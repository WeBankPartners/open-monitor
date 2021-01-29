package funcs

import (
	"os/exec"
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
	Port  string     `json:"port"`
	User  string  `json:"user"`
	Password   string  `json:"password"`
	DataBase  string  `json:"database"`
	DatabasePrefix string  `json:"database_prefix"`
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
	FiveMinStartDay int64 `json:"five_min_start_day"`
	ConcurrentInsertNum  int  `json:"concurrent_insert_num"`
	RetryWaitSecond  int  `json:"retry_wait_second"`
}

type HttpConfig struct {
	Enable  bool  `json:"enable"`
	Port    int   `json:"port"`
}

type GlobalConfig struct {
	Enable  string  `json:"enable"`
	Mysql  MysqlConfig  `json:"mysql"`
	Prometheus  PrometheusConfig `json:"prometheus"`
	Monitor  MonitorConfig  `json:"monitor"`
	Trans    TransConfig    `json:"trans"`
	Http     HttpConfig     `json:"http"`
}

var (
	config     *GlobalConfig
	lock       = new(sync.RWMutex)
	DefaultLocalTimeZone string
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
	initLocalTimeZone()
	hostIp = "127.0.0.1"
	if os.Getenv("MONITOR_HOST_IP") != "" {
		hostIp = os.Getenv("MONITOR_HOST_IP")
	}
	return nil
}

func initLocalTimeZone()  {
	cmdOut,err := exec.Command("/bin/sh", "-c", "date|awk '{print $5}'").Output()
	if err != nil {
		log.Printf("init local time zone fail,%s \n", err.Error())
	}else{
		cmdOutString := strings.TrimSpace(string(cmdOut))
		if cmdOutString != "" {
			DefaultLocalTimeZone = cmdOutString
			log.Printf("init local time zone to %s \n", DefaultLocalTimeZone)
		}else{
			DefaultLocalTimeZone = "CST"
		}
	}
}