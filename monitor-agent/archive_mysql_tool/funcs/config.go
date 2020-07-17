package funcs

import (
	"sync"
	"log"
	"os"
	"io/ioutil"
	"strings"
	"encoding/json"
)

type MysqlConfig struct {
	Type  string  `json:"type"`
	Server  string  `json:"server"`
	Port  int     `json:"port"`
	User  string  `json:"user"`
	Password   string  `json:"password"`
	DataBase  string  `json:"database"`
	MaxOpen  int  `json:"maxOpen"`
	MaxIdle  int  `json:"maxIdle"`
	Timeout  int  `json:"timeout"`
}

type PrometheusConfig struct {
	Server  string  `json:"server"`
	Port  int     `json:"port"`
	MaxHttpOpen  int  `json:"max_http_open"`
	MaxHttpIdle  int  `json:"max_http_idle"`
	HttpIdleTimeout  int  `json:"http_idle_timeout"`
	QueryStep    int  `json:"query_step"`
	IgnoreTags   []string  `json:"ignore_tags"`
}

type MonitorConfig struct {
	Mysql  MysqlConfig  `json:"mysql"`
}

type TransConfig struct {
	MaxUnitSpeed  int  `json:"max_unit_speed"`
}

type GlobalConfig struct {
	Mysql  MysqlConfig  `json:"mysql"`
	Prometheus  PrometheusConfig `json:"prometheus"`
	Monitor  MonitorConfig  `json:"monitor"`
	Trans    TransConfig    `json:"trans"`
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
	config = &c
	log.Println("read config file:", cfg, "successfully")
	lock.Unlock()
	return nil
}
