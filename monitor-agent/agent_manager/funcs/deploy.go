package funcs

import (
	"strings"
	"fmt"
	"os/exec"
	"log"
	"net"
	"time"
)

var deployNumMap map[string]int
var deployPathMap map[string]string
var deployGuidStatus map[string]string
var LocalIp string
var osBashCommand string
var osPsPidIndex string

func InitDeploy()  {
	deployNumMap = make(map[string]int)
	deployPathMap = make(map[string]string)
	deployGuidStatus = make(map[string]string)
	initOsCommand()
	if !Config().Deploy.Enable || len(Config().Deploy.PackagePath) == 0 {
		return
	}
	for _,v := range Config().Deploy.PackagePath {
		tmpName := strings.Split(v, "/")[strings.Count(v, "/")]
		deployNumMap[tmpName] = 0
		deployPathMap[tmpName] = v
	}
}

func AddDeploy(name,configFile,guid string, param map[string]string) (port int,err error) {
	port = 0
	if _,b := deployGuidStatus[guid]; b {
		log.Println("enter exist action")
		port = GlobalProcessMap[guid].Port
		DeleteDeploy(guid)
		//if v == "running" {
		//	DeleteDeploy(guid)
		//	return GlobalProcessMap[guid].Port,nil
		//}
		//if v == "stop" {
		//	err := GlobalProcessMap[guid].start("","","",0,nil)
		//	return GlobalProcessMap[guid].Port,err
		//}
	}
	if _,b := deployNumMap[name]; !b {
		return port,fmt.Errorf("%s can not find in the config file", name)
	}
	var p ProcessObj
	tmpName := fmt.Sprintf("%s_%d", name, deployNumMap[name]+1)
	deployNumMap[name] = deployNumMap[name] + 1
	deployPath := fmt.Sprintf("%s/%s", Config().Deploy.DeployDir, tmpName)
	err = exec.Command(osBashCommand, "-c", fmt.Sprintf("mkdir -p %s && cp -r %s/* %s/", deployPath, deployPathMap[name], deployPath)).Run()
	if err != nil {
		return port,err
	}
	if configFile != "" {
		configFile = deployPath + "/" + configFile
	}
	startFile := deployPath + "/start.sh"
	p.init(tmpName, deployPath, "./start.sh")
	ProcessMapLock.Lock()
	GlobalProcessMap[guid] = &p
	ProcessMapLock.Unlock()
	if port == 0 {
		port = GetPort()
	}
	param["port"] = fmt.Sprintf("%d", port)
	err = p.start(configFile, startFile, guid, port, param)
	if err != nil && p.Status == "broken" {
		p.destroy()
		delete(GlobalProcessMap, guid)
		return 0,err
	}
	deployGuidStatus[guid] = p.Status
	for k,v := range deployGuidStatus {
		log.Printf("deploy guid status ---> k:%s  v:%s \n", k, v)
	}
	return port,err
}

func DeleteDeploy(guid string) error {
	if v,b := GlobalProcessMap[guid]; b {
		err := v.stop()
		if err == nil {
			deployGuidStatus[guid] = "stop"
			for k,v := range deployGuidStatus {
				log.Printf("deploy guid status ---> k:%s  v:%s \n", k, v)
			}
			err = exec.Command(osBashCommand, "-c", fmt.Sprintf("mv -f %s /tmp/%s_%s", v.Path, v.Name, time.Now().Format("2006_01_02_15_04_05"))).Run()
			if err != nil {
				log.Printf("remove dir %s fail : %v \n", v.Path, err)
			}
		}
		return err
	}else{
		return fmt.Errorf("guid:%s not exist", guid)
	}
}

func InitLocalIp() bool {
	addrs, err := net.InterfaceAddrs()
	re := []string{}
	if err != nil {
		log.Println(err)
		return false
	}
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				re = append(re, ipNet.IP.String())
			}
		}
	}
	if len(re) == 0 {
		return false
	}else{
		LocalIp = re[0]
		log.Printf("local ip : %s \n", LocalIp)
		return true
	}
}

func initOsCommand()  {
	for _,v := range Config().OsBash {
		_,err := exec.Command(v, "-c", "date").Output()
		if err==nil {
			osBashCommand = v
			break
		}
	}
	if osBashCommand == "" {
		osBashCommand = "bash"
	}
	b,err := exec.Command(osBashCommand, "-c", "ps aux|grep PID|grep -v grep").Output()
	if err == nil {
		index := 0
		for _,v := range strings.Split(string(b), " ") {
			if v != "" {
				index += 1
				if v == "PID" {
					break
				}
			}
		}
		osPsPidIndex = fmt.Sprintf("$%d", index)
	}else{
		osPsPidIndex = "$2"
	}
	log.Printf("init os command done, bash: %s  index:%s \n", osBashCommand, osPsPidIndex)
}