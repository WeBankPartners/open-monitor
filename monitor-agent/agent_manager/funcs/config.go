package funcs

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
)

type HttpConfig struct {
	Port int `json:"port"`
}

type DeployConfig struct {
	Enable         bool     `json:"enable"`
	StartPort      int      `json:"start_port"`
	PackagePath    []string `json:"package_path"`
	DeployDir      string   `json:"deploy_dir"`
	EachMaxProcess int      `json:"each_max_process"`
}

type ManagerConfig struct {
	AliveCheck  int    `json:"alive_check"`
	AutoRestart bool   `json:"auto_restart"`
	Retry       int    `json:"retry"`
	SaveFile    string `json:"save_file"`
}

type ProcessConfig struct {
	Name string `json:"name"`
	Cmd  string `json:"cmd"`
}

type AgentsConfig struct {
	Process            []*ProcessConfig `json:"process"`
	HttpRegisterEnable bool             `json:"http_register_enable"`
}

type GlobalConfig struct {
	Http       *HttpConfig    `json:"http"`
	Deploy     *DeployConfig  `json:"deploy"`
	Manager    *ManagerConfig `json:"manager"`
	Agents     *AgentsConfig  `json:"agents"`
	OsBash     []string       `json:"os_bash"`
	RemoteMode string         `json:"remote_mode"`
}

var (
	config *GlobalConfig
	lock   = new(sync.RWMutex)
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
	b, err := ioutil.ReadFile(cfg)
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

type AgentManagerTable struct {
	EndpointGuid    string `json:"endpoint_guid"`
	Name            string `json:"name"`
	User            string `json:"user"`
	Password        string `json:"password"`
	InstanceAddress string `json:"instance_address"`
	AgentAddress    string `json:"agent_address"`
	ConfigFile      string `json:"config_file"`
	BinPath         string `json:"bin_path"`
	AgentRemotePort string `json:"agent_remote_port"`
}

type InitDeployParam struct {
	AgentManagerRemoteIp string               `json:"agentManagerRemoteIp"`
	Config               []*AgentManagerTable `json:"config"`
}
