package funcs

import (
	"sync"
	"os/exec"
	"strings"
	"strconv"
	"time"
	"log"
	"fmt"
)

var (
	ProcessMapLock = new(sync.RWMutex)
	GlobalProcessMap = make(map[string]*ProcessObj)
	portLock = new(sync.RWMutex)
	portList = []int{}
	autoRestartTime = 0
)

func StartManager()  {
	interval := 30
	if Config().Manager.AliveCheck > 0 {
		interval = Config().Manager.AliveCheck
	}
	if Config().Manager.AutoRestart {
		autoRestartTime = Config().Manager.Retry
	}
	t := time.NewTicker(time.Second*time.Duration(interval)).C
	for {
		<- t
		pids := getSystemProcessPid("")
		if len(pids) == 0 {
			continue
		}
		ProcessMapLock.RLock()
		for _,v := range GlobalProcessMap {
			tmpPid,tmpName,tmpStatus,tmpRetry := v.message()
			if tmpStatus == "running" {
				if !containsInt(tmpPid, pids) {
					v.update(1)
					tmpPid,_,_,_ = v.message()
					if !containsInt(tmpPid, pids) {
						v.update(2)
					}
				}
			}
			if tmpStatus == "dead" && autoRestartTime > tmpRetry {
				err := v.start("", "", nil)
				v.update(3)
				if err != nil {
					log.Printf("retry to start %s fail,error : %v \n", tmpName, err)
				}
			}
		}
		ProcessMapLock.RUnlock()
	}
}

func GetPort() int {
	portLock.Lock()
	defer portLock.Unlock()
	var tmpPort int
	maxNum := 0
	if len(portList) == 0 {
		tmpPort = Config().Deploy.StartPort
	}else {
		maxNum = portList[len(portList)-1]
		tmpPort = maxNum+1
	}
	b,err := exec.Command("bash", "-c", "netstat -ltn | awk '{print $4}'").Output()
	if err != nil {
		portList = append(portList, tmpPort)
		return tmpPort
	}
	for _,v := range strings.Split(string(b),"\n") {
		if strings.Contains(v, ":") {
			portString := strings.Split(v, ":")[strings.Count(v, ":")]
			sysMaxPort,_ := strconv.Atoi(portString)
			if sysMaxPort > tmpPort {
				tmpPort = sysMaxPort
			}
		}
	}
	if tmpPort != maxNum+1 {
		tmpPort = tmpPort + 1
	}
	portList = append(portList, tmpPort)
	return tmpPort
}

func containsInt(i int, l []int) bool {
	for _,v := range l {
		if i == v {
			return true
		}
	}
	return false
}

func PrintProcessList() []byte {
	var result string
	ProcessMapLock.RLock()
	for k,v := range GlobalProcessMap {
		result = result + fmt.Sprintf("%s : %s \n", k, v.print())
	}
	ProcessMapLock.RUnlock()
	return []byte(result)
}