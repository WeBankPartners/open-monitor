package funcs

import (
	"sync"
	"os/exec"
	"strings"
	"strconv"
	"time"
	"log"
	"fmt"
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"os"
	"encoding/json"
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
			justDead := false
			if tmpStatus == "running" {
				if !containsInt(tmpPid, pids) {
					v.update(1)
					tmpPid,_,_,_ = v.message()
					if !containsInt(tmpPid, pids) {
						v.update(2)
						justDead = true
					}
				}
			}
			if justDead || (tmpStatus == "dead" && autoRestartTime > tmpRetry) {
				err := v.start("", "","", 0, nil)
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
	b,err := exec.Command(osBashCommand, "-c", "netstat -ltn | awk '{print $4}'").Output()
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

func StopDeployProcess()  {
	ProcessMapLock.RLock()
	for k,v := range GlobalProcessMap {
		err := v.stop()
		if err != nil {
			log.Printf("stop %s error : %v \n", k, err)
		}else{
			log.Printf("stop %s success \n", k)
		}
	}
	ProcessMapLock.RUnlock()
}

func SaveDeployProcess()  {
	var processList []string
	ProcessMapLock.RLock()
	for _,v := range GlobalProcessMap {
		if v.Deploy {
			processList = append(processList, v.print())
		}
	}
	var tmpBuffer bytes.Buffer
	enc := gob.NewEncoder(&tmpBuffer)
	err := enc.Encode(processList)
	if err != nil {
		log.Printf("save deploy process error : %v \n", err)
	}else{
		ioutil.WriteFile("process.data", tmpBuffer.Bytes(), 0644)
		log.Println("save deploy process success")
	}
	ProcessMapLock.RUnlock()
}

func LoadDeployProcess()  {
	var processList []string
	file,err := os.Open("process.data")
	if err == nil {
		dec := gob.NewDecoder(file)
		err = dec.Decode(&processList)
		if err != nil {
			log.Printf("gob decode process.data error : %v \n", err)
			return
		}
		for _,v := range processList {
			var p ProcessObj
			err = json.Unmarshal([]byte(v), &p)
			if err != nil {
				log.Printf("%s unmarshal error : %v \n", v, err)
			}else{
				p.Lock = new(sync.RWMutex)
				p.Pid = 0
				p.Deploy = true
				if p.Status == "running" {
					err = p.start("", "", "", 0, nil)
					if err != nil {
						log.Printf("process start error : %v \n", err)
					} else {
						GlobalProcessMap[p.Guid] = &p
					}
				}
			}
		}
	}
}