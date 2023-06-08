package funcs

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ProcessMapLock   = new(sync.RWMutex)
	GlobalProcessMap = make(map[string]*ProcessObj)
	portLock         = new(sync.RWMutex)
	portList         = []int{}
	autoRestartTime  = 0
	RemoteMode       = false
)

func StartManager() {
	log.Println("start manager")
	interval := 30
	if Config().Manager.AliveCheck > 0 {
		interval = Config().Manager.AliveCheck
	}
	if Config().Manager.AutoRestart {
		autoRestartTime = Config().Manager.Retry
	}
	t := time.NewTicker(time.Second * time.Duration(interval)).C
	for {
		<-t
		pids := getSystemProcessPid("", "")
		if len(pids) == 0 {
			continue
		}
		ProcessMapLock.RLock()
		for _, v := range GlobalProcessMap {
			tmpPid, tmpName, tmpStatus, tmpRetry := v.message()
			justDead := false
			if tmpStatus == "running" {
				if !containsInt(tmpPid, pids) {
					v.update(1)
					tmpPid, _, _, _ = v.message()
					if !containsInt(tmpPid, pids) {
						v.update(2)
						justDead = true
					}
				}
			}
			if justDead || (tmpStatus == "dead" && autoRestartTime > tmpRetry) {
				err := v.start("", "", "", 0, nil)
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
	var resultPort int
	existMaxPort := 0
	configMinPort := Config().Deploy.StartPort
	configMaxPort := configMinPort + 10000
	if len(portList) == 0 {
		resultPort = configMinPort
	} else {
		existMaxPort = portList[len(portList)-1]
		for _, v := range portList {
			if existMaxPort < v {
				existMaxPort = v
			}
		}
		resultPort = existMaxPort + 1
	}
	b, err := exec.Command(osBashCommand, "-c", "netstat -ltn | awk '{print $4}'").Output()
	if err != nil {
		portList = append(portList, resultPort)
		return resultPort
	}
	sysPortMap := make(map[int]int)
	for _, v := range strings.Split(string(b), "\n") {
		if strings.Contains(v, ":") {
			portString := strings.Split(v, ":")[strings.Count(v, ":")]
			sysPort, _ := strconv.Atoi(portString)
			if sysPort >= configMaxPort && sysPort < configMinPort {
				continue
			}
			sysPortMap[sysPort] = 1
		}
	}
	if _, ok := sysPortMap[resultPort]; ok {
		for i := resultPort; i < configMaxPort; i++ {
			if _, existFlag := sysPortMap[i]; !existFlag {
				resultPort = i
				break
			}
		}
	}
	portList = append(portList, resultPort)
	return resultPort
}

func containsInt(i int, l []int) bool {
	for _, v := range l {
		if i == v {
			return true
		}
	}
	return false
}

func PrintProcessList() []byte {
	var result string
	ProcessMapLock.RLock()
	for k, v := range GlobalProcessMap {
		result = result + fmt.Sprintf("%s : %s \n", k, v.print())
	}
	ProcessMapLock.RUnlock()
	return []byte(result)
}

func StopDeployProcess() {
	ProcessMapLock.RLock()
	for k, v := range GlobalProcessMap {
		err := v.stop()
		if err != nil {
			log.Printf("stop %s error : %v \n", k, err)
		} else {
			log.Printf("stop %s success \n", k)
		}
	}
	ProcessMapLock.RUnlock()
}

func SaveDeployProcess() {
	var processList []string
	filePath := "process.data"
	if Config().Manager.SaveFile != "" {
		filePath = Config().Manager.SaveFile
	}
	ProcessMapLock.RLock()
	for _, v := range GlobalProcessMap {
		if v.Deploy {
			processList = append(processList, v.print())
		}
	}
	var tmpBuffer bytes.Buffer
	enc := gob.NewEncoder(&tmpBuffer)
	err := enc.Encode(processList)
	if err != nil {
		log.Printf("save deploy process error : %v \n", err)
	} else {
		ioutil.WriteFile(filePath, tmpBuffer.Bytes(), 0644)
		log.Println("save deploy process success")
	}
	ProcessMapLock.RUnlock()
}

func LoadDeployProcess() {
	log.Println("load deploy process")
	var processList []string
	filePath := "process.data"
	if Config().Manager.SaveFile != "" {
		filePath = Config().Manager.SaveFile
	}
	file, err := os.Open(filePath)
	if err == nil {
		dec := gob.NewDecoder(file)
		err = dec.Decode(&processList)
		if err != nil {
			log.Printf("gob decode process.data error : %v \n", err)
			return
		}
		agentMangerLocalMode := strings.ToLower(os.Getenv("MONITOR_AGENT_MANAGER_REMOTE_MODE"))
		if agentMangerLocalMode == "y" || agentMangerLocalMode == "yes" || agentMangerLocalMode == "true" {
			RemoteMode = true
			return
		}
		for _, v := range processList {
			var p ProcessObj
			err = json.Unmarshal([]byte(v), &p)
			if err != nil {
				log.Printf("%s unmarshal error : %v \n", v, err)
			} else {
				p.Lock = new(sync.RWMutex)
				p.Pid = 0
				p.Deploy = true
				if p.Status == "running" {
					err = p.start("", "", "", 0, nil)
					if err != nil {
						log.Printf("process start error : %v \n", err)
					} else {
						GlobalProcessMap[p.Guid] = &p
						deployGuidStatus[p.Guid] = p.Status
					}
				}
			}
		}
	}
}

func clearUselessDir(path string) {
	if !strings.HasPrefix(path, Config().Deploy.DeployDir) || Config().Deploy.DeployDir == "" {
		return
	}
	_, err := os.Stat(path)
	if os.IsExist(err) {
		return
	}
	err = exec.Command(osBashCommand, "-c", fmt.Sprintf("rm -rf %s", path)).Run()
	if err != nil {
		log.Printf("clear useless dir error %v \n", err)
	}
}

func CleanDeployDir() {
	log.Println("start clean deploy dir")
	var dirList []string
	files, err := ioutil.ReadDir(Config().Deploy.DeployDir)
	if err != nil {
		log.Printf("read dir %s error %v \n", Config().Deploy.DeployDir, err)
	} else {
		for _, v := range files {
			if v.Name() == "process.data" {
				continue
			}
			dirList = append(dirList, v.Name())
		}
	}
	for _, v := range dirList {
		alive := false
		for _, vv := range GlobalProcessMap {
			if strings.Contains(vv.Path, v) {
				alive = true
				break
			}
		}
		if !alive {
			clearUselessDir(fmt.Sprintf("%s/%s", Config().Deploy.DeployDir, v))
		}
	}
}

func InitDeployDir(param []*AgentManagerTable) error {
	paramByte, _ := json.Marshal(param)
	log.Printf("init deploy dir : param -> %s \n", string(paramByte))
	var tmpDeleteList []string
	for k, v := range GlobalProcessMap {
		alive := false
		for _, vv := range param {
			if vv.EndpointGuid == v.Guid {
				configHash := fmt.Sprintf("%s_%s_%s", vv.InstanceAddress, vv.User, vv.Password)
				if v.ConfigHash != configHash {
					break
				}
				if strings.Contains(vv.AgentAddress, fmt.Sprintf(":%d", v.Port)) {
					alive = true
					break
				}
			}
		}
		if !alive {
			tmpDeleteList = append(tmpDeleteList, k)
		}
	}
	for _, v := range tmpDeleteList {
		DeleteDeploy(v)
	}
	var err error
	for _, v := range param {
		isExist := false
		for _, vv := range GlobalProcessMap {
			if vv.Guid == v.EndpointGuid {
				isExist = true
				break
			}
		}
		if !isExist {
			tmpParam := make(map[string]string)
			tmpParam["guid"] = v.EndpointGuid
			tmpParam["exporter"] = v.BinPath
			if v.ConfigFile != "" {
				tmpParam["config"] = v.ConfigFile
			}
			if strings.Contains(v.InstanceAddress, ":") {
				tmpParam["instance_server"] = v.InstanceAddress[:strings.Index(v.InstanceAddress, ":")]
				tmpParam["instance_port"] = v.InstanceAddress[strings.Index(v.InstanceAddress, ":")+1:]
			} else {
				err = fmt.Errorf("guid: %s instance address illegal: %s ", v.EndpointGuid, v.InstanceAddress)
				continue
			}
			if strings.Contains(v.AgentAddress, ":") {
				tmpParam["port"] = v.AgentAddress[strings.Index(v.AgentAddress, ":")+1:]
			} else {
				err = fmt.Errorf("guid: %s agent address illegal: %s ", v.EndpointGuid, v.AgentAddress)
				continue
			}
			tmpParam["auth_user"] = v.User
			tmpParam["auth_password"] = v.Password
			configHash := fmt.Sprintf("%s_%s_%s", v.InstanceAddress, v.User, v.Password)
			_, deployErr := AddDeploy(v.BinPath, v.ConfigFile, v.EndpointGuid, tmpParam, configHash)
			if deployErr != nil {
				err = deployErr
				continue
			}
		}
	}
	if err != nil {
		log.Printf("init deploy dir meet error but continue,message:%s  \n", err.Error())
	} else {
		log.Printf("init deploy dir done \n")
	}
	SaveDeployProcess()
	return err
}
