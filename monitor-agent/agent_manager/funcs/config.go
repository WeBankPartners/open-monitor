package funcs

import (
	"sync"
	"log"
	"encoding/json"
	"io/ioutil"
	"strings"
	"os"
)

type HttpConfig struct {
	Port  int  `json:"port"`
}

type DeployConfig struct {
	Enable  bool  `json:"enable"`
	StartPort  int  `json:"start_port"`
	PackagePath  []string  `json:"package_path"`
	DeployDir  string  `json:"deploy_dir"`
	EachMaxProcess  int  `json:"each_max_process"`
}

type ManagerConfig struct {
	AliveCheck  int  `json:"alive_check"`
	AutoRestart  bool  `json:"auto_restart"`
	Retry  int  `json:"retry"`
}

type ProcessConfig  struct {
	Name  string  `json:"name"`
	Cmd  string  `json:"cmd"`
}

type AgentsConfig struct {
	Process  []*ProcessConfig  `json:"process"`
	HttpRegisterEnable  bool  `json:"http_register_enable"`
}

type GlobalConfig struct {
	Http  *HttpConfig  `json:"http"`
	Deploy  *DeployConfig  `json:"deploy"`
	Manager  *ManagerConfig  `json:"manager"`
	Agents  *AgentsConfig  `json:"agents"`
}

var (
	config     *GlobalConfig
	lock       = new(sync.RWMutex)
)

func Config() *GlobalConfig {
	lock.RLock()
	defer lock.RUnlock()
	return config
}

func InitConfig(cfg string) error {
	if cfg == "" {
		log.Println("use -c to specify configuration file")
	}
	_, err := os.Stat(cfg)
	if os.IsExist(err) {
		log.Println("config file not found")
		return err
	}
	b,err := ioutil.ReadFile(cfg)
	if err != nil {
		log.Printf("read file %s error %v \n", cfg, err)
		return err
	}
	configContent := strings.TrimSpace(string(b))
	var c GlobalConfig
	err = json.Unmarshal([]byte(configContent), &c)
	if err != nil {
		log.Println("parse config file:", cfg, "fail:", err)
		return err
	}
	lock.Lock()
	defer lock.Unlock()
	config = &c
	log.Println("read config file:", cfg, "successfully")
	return nil
}
