package funcs

import (
	"strings"
	"fmt"
	"os/exec"
)

var deployNumMap map[string]int
var deployPathMap map[string]string

func InitDeploy()  {
	deployNumMap = make(map[string]int)
	deployPathMap = make(map[string]string)
	if !Config().Deploy.Enable || len(Config().Deploy.PackagePath) == 0 {
		return
	}
	for _,v := range Config().Deploy.PackagePath {
		tmpName := strings.Split(v, "/")[strings.Count(v, "/")]
		deployNumMap[tmpName] = 0
		deployPathMap[tmpName] = v
	}
}

func AddDeploy(name,configFile string, param map[string]string) error {
	for k,v := range param {
		fmt.Printf("add deploy param k:%s v:%s \n", k, v)
	}
	if _,b := deployNumMap[name]; !b {
		return fmt.Errorf("%s can not find in the config file", name)
	}
	var p ProcessObj
	tmpName := fmt.Sprintf("%s_%d", name, deployNumMap[name]+1)
	deployPath := fmt.Sprintf("%s/%s", Config().Deploy.DeployDir, tmpName)
	err := exec.Command("bash", "-c", fmt.Sprintf("mkdir -p %s && cp -r %s/* %s/", deployPath, deployPathMap[name], deployPath)).Run()
	if err != nil {
		return err
	}
	configFile = deployPath + "/" + configFile
	startFile := deployPath + "/start.sh"
	p.init(tmpName, deployPath, "./start.sh")
	ProcessMapLock.Lock()
	GlobalProcessMap[tmpName] = &p
	ProcessMapLock.Unlock()
	deployNumMap[name] = deployNumMap[name] + 1
	port := GetPort()
	param["port"] = fmt.Sprintf("%d", port)
	err = p.start(configFile, startFile, param)
	return err
}