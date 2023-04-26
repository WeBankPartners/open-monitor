package icmpping

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-agent/ping_exporter/funcs"
	"net"
	"time"
)

// icmp报头,8byte
type ICMP struct {
	Type        uint8  // 类型,8是请求,0是应答
	Code        uint8  // 代码,它与类型字段一起共同标识了ICMP报文的详细类型,比如说类型为3表示不可达,此时代码为0表示网络不可达,为1表示满意主机不可达等
	Checksum    uint16 // 校验和,对包括ICMP报文数据部分在内的整个ICMP数据报的校验和,以检验报文在传输过程中是否出现了差错,和IP报头中校验和计算方法一样
	Identifier  uint16 // 标识,用于标识本ICMP进程,但仅适用于回显请求和应答ICMP报文,对于目标不可达和超时,该字段为0
	SequenceNum uint16 // 序列号
}

var (
	icmpData     ICMP
	icmpBytes    []byte
	localAddress net.IPAddr = net.IPAddr{IP: net.ParseIP("0.0.0.0")}
)

func checkSum(data []byte) uint16 {
	var (
		sum    uint32
		length int = len(data)
		index  int
	)
	for length > 1 {
		sum += uint32(data[index])<<8 + uint32(data[index+1])
		index += 2
		length -= 2
	}
	if length > 0 {
		sum += uint32(data[index])
	}
	sum += (sum >> 16)
	return uint16(^sum)
}

func InitIcmpBytes() {
	icmpData.Type = 8 //8->echo message  0->reply message
	icmpData.Code = 0
	icmpData.Checksum = 0
	//icmpData.Identifier = 0
	//icmpData.SequenceNum = 0
	icmpData.Identifier = 1
	icmpData.SequenceNum = 1
	var buffer bytes.Buffer
	// 先在buffer中写入icmp数据报求去校验和
	binary.Write(&buffer, binary.BigEndian, icmpData)
	icmpData.Checksum = checkSum(buffer.Bytes())
	// 然后清空buffer并把求完校验和的icmp数据报写入其中准备发送
	buffer.Reset()
	binary.Write(&buffer, binary.BigEndian, icmpData)
	icmpBytes = buffer.Bytes()
}

// 统计一分钟的丢包率，ping 20个包，每个包3秒
func StartPingLossPacket(distIp string) float64 {
	raddr := net.IPAddr{IP: net.ParseIP(distIp)}
	//如果你要使用网络层的其他协议还可以设置成 ip:ospf、ip:arp 等
	conn, err := net.DialIP("ip4:icmp", &localAddress, &raddr)
	if err != nil {
		fmt.Println(err.Error())
		return 100
	}
	defer conn.Close()
	step := time.NewTicker(3 * time.Second).C
	var lossCount float64
	for i := 0; i < 20; i++ {
		r := doping(*conn, distIp, 3)
		if r == 0 {
			lossCount += 1
		}
		if i < 19 {
			<-step
		}
	}
	return (lossCount / 20) * 100
}

func StartPing(distIp string, timeout int) (int, float64, bool) {
	var raddr net.IPAddr = net.IPAddr{IP: net.ParseIP(distIp)}
	isConfused := false
	//如果你要使用网络层的其他协议还可以设置成 ip:ospf、ip:arp 等
	conn, err := net.DialIP("ip4:icmp", &localAddress, &raddr)
	if err != nil {
		fmt.Println(err.Error())
		return 3, 0, isConfused
	}
	defer conn.Close()
	re := 0
	tq := 0
	startTime := time.Now()
	for i := 0; i < 5; i++ {
		r := doping(*conn, distIp, timeout)
		if r == 2 {
			tq++
			r = 0
		}
		re += r
	}
	useTime := float64(time.Now().Sub(startTime).Nanoseconds()) / 1e6
	if re >= 2 { // 发5个ICMP包,如果有2个回复成功则算ping通
		addSuccessIp(distIp)
		return 0, useTime / float64(re), isConfused
	} else {
		isConfused = true
		if tq == 5 {
			//addSuccessIp(distIp)
			return 0, useTime, isConfused
		}
		if re == 1 && tq == 4 {
			//addSuccessIp(distIp)
			return 0, useTime, isConfused
		}
		if re == 2 && tq >= 1 { // 如果有2个回复成功和3个太快回复(下面把这当做了一种异常,有时候主机不通也会出现这种情况),也算主机是通的
			//addSuccessIp(distIp)
			return 0, useTime / 2, false
		}
		funcs.DebugLog("%s ping fail,%.3f ms, renum : %d ## ", distIp, useTime, re)
		return 1, useTime, false
		//if useTime < 6100 {  // 如果4个包不是全部2秒超时,则算异常需要重试
		//	funcs.DebugLog("%s ping retry,%.3f ms, renum : %d ## ", distIp, useTime, re)
		//	t := GetRetryMap(distIp, re)
		//	if t>=4{  // 如果这几次检测中有总数超过4个成功的包返回,也算是成功,经测试在网络流量高的时候会有大概5%-10%的测试IP会只返回2个响应成功的包
		//		addSuccessIp(distIp)
		//		return 0,useTime/float64(re)
		//	}
		//	addRetryIp(distIp)
		//	return 2,0
		//}else {
		//	funcs.DebugLog("%s ping fail,%.3f ms, renum : %d ## ", distIp, useTime, re)
		//	return 1,useTime
		//}
	}
}

// return 0->down 1->up 2->too quick(maybe get the result from system cache)
func doping(conn net.IPConn, distIp string, timeout int) int {
	// 发送请求
	if _, err := conn.Write(icmpBytes); err != nil {
		fmt.Println(err.Error())
		return 0
	}
	receiveByte := make([]byte, 1024)
	startTime := time.Now()
	isOk := 1
	// 设置超时时间
	conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(timeout)))
	// 读取返回报文
	_, err := conn.Read(receiveByte)
	useTime := float64(time.Now().Sub(startTime).Nanoseconds()) / 1e6
	if err != nil {
		//fmt.Printf("%s ping time out, use time %.3f \n", distIp, use_time)
		isOk = 0
	} else {
		if len(receiveByte) < 23 {
			funcs.DebugLog("icmp响应报文格式错误,无法解析")
			isOk = 0
		} else {
			var typeint uint8
			// IP报头20字节,ICMP报头8字节,从21到28,这里是取返回的ICMP类型1字节
			responseBuffer := bytes.NewBuffer(receiveByte[20:21])
			binary.Read(responseBuffer, binary.BigEndian, &typeint)
			if typeint == 0 {
				if useTime < 0.1 {
					// 返回太快了,无效
					funcs.DebugLog("%s ping fail,%.3f ms, to quick !! ", distIp, useTime)
					isOk = 2
				} else {
					funcs.DebugLog("%s ping success,%.3f ms, ok !! ", distIp, useTime)
				}
			} else {
				var codeint uint8
				// ICMP返回的代码1字节
				responseBuffer = bytes.NewBuffer(receiveByte[21:22])
				binary.Read(responseBuffer, binary.BigEndian, &codeint)
				msg := IcmpType(typeint, codeint)
				funcs.DebugLog("%s ping fail, error : %s ", distIp, msg)
				isOk = 0
			}
		}
	}
	return isOk
}
