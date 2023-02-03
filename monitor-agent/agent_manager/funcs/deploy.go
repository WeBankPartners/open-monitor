package funcs

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os/exec"
	"strconv"
	"strings"
)

var deployPathMap map[string]string
var deployGuidStatus map[string]string
var LocalIp string
var osBashCommand string
var osPsPidIndex string

func InitDeploy() {
	log.Println("init deploy")
	deployPathMap = make(map[string]string)
	deployGuidStatus = make(map[string]string)
	initOsCommand()
	if !Config().Deploy.Enable || len(Config().Deploy.PackagePath) == 0 {
		return
	}
	for _, v := range Config().Deploy.PackagePath {
		tmpName := strings.Split(v, "/")[strings.Count(v, "/")]
		deployPathMap[tmpName] = v
	}
}

func AddDeploy(name, configFile, guid string, param map[string]string, configHash string) (port int, err error) {
	port = 0
	if _, b := deployGuidStatus[guid]; b {
		log.Println("enter exist action")
		port = GlobalProcessMap[guid].Port
		DeleteDeploy(guid)
	}
	var p ProcessObj
	p.ConfigHash = configHash
	tmpName := fmt.Sprintf("%s_%d", name, getNextDirIndex(name))
	deployPath := fmt.Sprintf("%s/%s", Config().Deploy.DeployDir, tmpName)
	err = exec.Command(osBashCommand, "-c", fmt.Sprintf("mkdir -p %s && cp -r %s/* %s/", deployPath, deployPathMap[name], deployPath)).Run()
	if err != nil {
		return port, err
	}
	if configFile != "" {
		configFile = deployPath + "/" + configFile
	}
	startFile := deployPath + "/start.sh"
	p.init(tmpName, deployPath, "./start.sh")
	ProcessMapLock.Lock()
	GlobalProcessMap[guid] = &p
	ProcessMapLock.Unlock()
	if _, b := param["port"]; !b {
		if port == 0 {
			port = GetPort()
		}
		param["port"] = fmt.Sprintf("%d", port)
	} else {
		port, _ = strconv.Atoi(param["port"])
	}
	err = p.start(configFile, startFile, guid, port, param)
	if err != nil && p.Status == "broken" {
		p.destroy()
		delete(GlobalProcessMap, guid)
		return 0, err
	}
	deployGuidStatus[guid] = p.Status
	for k, v := range deployGuidStatus {
		log.Printf("deploy guid status ---> k:%s  v:%s \n", k, v)
	}
	return port, err
}

func DeleteDeploy(guid string) error {
	log.Printf("try to delete %s \n", guid)
	if v, b := GlobalProcessMap[guid]; b {
		v.stop()
		clearUselessDir(v.Path)
		delete(GlobalProcessMap, guid)
		delete(deployGuidStatus, guid)
		return nil
	} else {
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
	} else {
		LocalIp = re[0]
		log.Printf("local ip : %s \n", LocalIp)
		return true
	}
}

func initOsCommand() {
	for _, v := range Config().OsBash {
		_, err := exec.Command(v, "-c", "date").Output()
		if err == nil {
			osBashCommand = v
			break
		}
	}
	if osBashCommand == "" {
		osBashCommand = "bash"
	}
	b, err := exec.Command(osBashCommand, "-c", "ps aux|grep PID|grep -v grep").Output()
	if err == nil {
		index := 0
		for _, v := range strings.Split(string(b), " ") {
			if v != "" {
				index += 1
				if v == "PID" {
					break
				}
			}
		}
		osPsPidIndex = fmt.Sprintf("$%d", index)
	} else {
		osPsPidIndex = "$2"
	}
	log.Printf("init os command done, bash: %s  index:%s \n", osBashCommand, osPsPidIndex)
}

func getNextDirIndex(name string) int {
	log.Printf("start get next dir index for %s \n", name)
	index := 0
	files, err := ioutil.ReadDir(Config().Deploy.DeployDir)
	if err != nil {
		log.Printf("read dir %s error %v \n", Config().Deploy.DeployDir, err)
	} else {
		for _, v := range files {
			if strings.Contains(v.Name(), name) {
				tmpList := strings.Split(v.Name(), "_")
				tmpIndex, _ := strconv.Atoi(tmpList[len(tmpList)-1])
				if tmpIndex > index {
					index = tmpIndex
				}
			}
		}
	}
	return index + 1
}
