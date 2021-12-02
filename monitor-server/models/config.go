package models

import (
	"encoding/json"
	"github.com/WeBankPartners/go-common-lib/cipher"
	"github.com/toolkits/file"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

type LogConfig struct {
	Level            string `json:"level"`
	File             string `json:"file"`
	AccessFile       string `json:"access_file"`
	ArchiveMaxSize   int    `json:"archive_max_size"`
	ArchiveMaxBackup int    `json:"archive_max_backup"`
	ArchiveMaxDay    int    `json:"archive_max_day"`
	Compress         bool   `json:"compress"`
}

type LdapConfig struct {
	Enable     bool     `json:"enable"`
	Server     string   `json:"server"`
	Port       int      `json:"port"`
	BindDN     string   `json:"bindDN"`
	BaseDN     string   `json:"baseDN"`
	Filter     string   `json:"filter"`
	Attributes []string `json:"attributes"`
}

type SessionRedisConfig struct {
	Enable  bool   `json:"enable"`
	Server  string `json:"server"`
	Port    int    `json:"port"`
	Pwd     string `json:"pwd"`
	Db      int    `json:"db"`
	MaxIdle int    `json:"max_idle"`
}

type SessionConfig struct {
	Enable       bool               `json:"enable"`
	Expire       int64              `json:"expire"`
	ServerEnable bool               `json:"server_enable"`
	ServerToken  string             `json:"server_token"`
	Redis        SessionRedisConfig `json:"redis"`
}

type HttpConfig struct {
	Port            string         `json:"port"`
	Swagger         bool           `json:"swagger"`
	Cross           bool           `json:"cross"`
	ReturnError     bool           `json:"return_error"`
	Alive           int64          `json:"alive"`
	Ldap            *LdapConfig    `json:"ldap"`
	Log             *LogConfig     `json:"log"`
	Session         *SessionConfig `json:"session"`
	DefaultLanguage string         `json:"default_language"`
}

type StoreConfig struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Server   string `json:"server"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Pwd      string `json:"pwd"`
	DataBase string `json:"database"`
	MaxOpen  int    `json:"maxOpen"`
	MaxIdle  int    `json:"maxIdle"`
	Timeout  int    `json:"timeout"`
}

type DataSourceConfig struct {
	Env        string               `json:"env"`
	Servers    []*DatasourceServers `json:"servers"`
	DivideTime int64                `json:"divide_time"`
	WaitTime   int                  `json:"wait_time"`
}

type DependenceConfig struct {
	Name     string `json:"name"`
	Server   string `json:"server"`
	Username string `json:"username"`
	Password string `json:"password"`
	Expire   int    `json:"expire"`
}

type AgentConfig struct {
	AgentType  string `json:"agent_type"`
	AgentBin   string `json:"agent_bin"`
	Port       string `json:"port"`
	User       string `json:"user"`
	Password   string `json:"password"`
	ConfigFile string `json:"config_file"`
}

type DatasourceServers struct {
	Id    int    `json:"id"`
	Type  string `json:"type"`
	Env   string `json:"env"`
	Host  string `json:"host"`
	Token string `json:"token"`
}

type PrometheusConfig struct {
	SdConfigPath   string `json:"sd_config_path"`
	RuleConfigPath string `json:"rule_config_path"`
	ConfigReload   string `json:"config_reload"`
}

type AlertMailConfig struct {
	Enable   bool   `json:"enable"`
	Protocol string `json:"protocol"`
	Tls      bool   `json:"tls"`
	Sender   string `json:"sender"`
	User     string `json:"user"`
	Password string `json:"password"`
	Server   string `json:"server"`
	Token    string `json:"token"`
}

type AlertConfig struct {
	Enable bool            `json:"enable"`
	Mail   AlertMailConfig `json:"mail"`
}

type PeerConfig struct {
	Enable          bool     `json:"enable"`
	InstanceHostIp  string   `json:"instance_host_ip"`
	HttpPort        string   `json:"http_port"`
	OtherServerList []string `json:"other_server_list"`
}

type CronJobConfig struct {
	Enable   bool `json:"enable"`
	Interval int  `json:"interval"`
}

type SdFileConfig struct {
	Enable bool   `json:"enable"`
	Path   string `json:"path"`
}

type ArchiveMysqlConfig struct {
	Enable             string `json:"enable"`
	Type               string `json:"type"`
	Server             string `json:"server"`
	Port               string `json:"port"`
	User               string `json:"user"`
	Password           string `json:"password"`
	DataBase           string `json:"database"`
	DatabasePrefix     string `json:"database_prefix"`
	MaxOpen            int    `json:"maxOpen"`
	MaxIdle            int    `json:"maxIdle"`
	Timeout            int    `json:"timeout"`
	LocalStorageMaxDay int64  `json:"local_storage_max_day"`
	FiveMinStartDay    int64  `json:"five_min_start_day"`
}

type CapacityServerConfig struct {
	Server             string `json:"server"`
	Port               string `json:"port"`
}

type GlobalConfig struct {
	IsPluginMode     string              `json:"is_plugin_mode"`
	Http             *HttpConfig         `json:"http"`
	Store            StoreConfig         `json:"store"`
	Datasource       DataSourceConfig    `json:"datasource"`
	LimitIp          []string            `json:"limitIp"`
	Dependence       []*DependenceConfig `json:"dependence"`
	Prometheus       PrometheusConfig    `json:"prometheus"`
	TagBlacklist     []string            `json:"tag_blacklist"`
	Agent            []*AgentConfig      `json:"agent"`
	Alert            AlertConfig         `json:"alert"`
	Peer             PeerConfig          `json:"peer"`
	CronJob          CronJobConfig       `json:"cron_job"`
	SdFile           SdFileConfig        `json:"sd_file"`
	ArchiveMysql     ArchiveMysqlConfig  `json:"archive_mysql"`
	ProcessCheckList []string            `json:"process_check_list"`
	DefaultAdminRole string              `json:"default_admin_role"`
	AlarmAliveMaxDay int                 `json:"alarm_alive_max_day"`
	CapacityServer   CapacityServerConfig `json:"capacity_server"`
}

var (
	ConfigFile           string
	config               *GlobalConfig
	lock                 = new(sync.RWMutex)
	CoreUrl              string
	CoreJwtKey           string
	FiringCallback       string
	RecoverCallback      string
	SubSystemCode        string
	SubSystemKey         string
	DefaultMailReceiver  []string
	DefaultLocalTimeZone string
	PluginRunningMode    bool
	SmsParamMaxLength    int
)

func Config() *GlobalConfig {
	lock.RLock()
	defer lock.RUnlock()
	return config
}

func InitConfig(cfg string) {
	if cfg == "" {
		log.Fatalln("use -c to specify configuration file")
	}

	if !file.IsExist(cfg) {
		log.Fatalln("config file:", cfg, "is not existent. maybe you need `mv cfg.example.json cfg.json`")
	}

	ConfigFile = cfg

	configContent, err := file.ToTrimString(cfg)
	if err != nil {
		log.Fatalln("read config file:", cfg, "fail:", err)
	}

	var c GlobalConfig
	err = json.Unmarshal([]byte(configContent), &c)
	if err != nil {
		log.Fatalln("parse config file:", cfg, "fail:", err)
	}
	rsaPemByte, readRsaErr := ioutil.ReadFile(RsaPemPath)
	if readRsaErr != nil {
		log.Printf("Read rsa pem file fail,%s ", readRsaErr.Error())
	}
	c.Store.Pwd, _ = cipher.DecryptRsa(c.Store.Pwd, string(rsaPemByte))
	c.ArchiveMysql.Password, _ = cipher.DecryptRsa(c.ArchiveMysql.Password, string(rsaPemByte))
	if c.IsPluginMode == "yes" || c.IsPluginMode == "y" || c.IsPluginMode == "true" {
		PluginRunningMode = true
	} else {
		PluginRunningMode = false
	}
	lock.Lock()
	defer lock.Unlock()

	config = &c
	for _, v := range config.Dependence {
		if v.Name == "core" {
			CoreUrl = v.Server
			if strings.HasSuffix(CoreUrl, "/") {
				CoreUrl = CoreUrl[:len(CoreUrl)-1]
			}
			break
		}
	}
	CoreJwtKey, _ = cipher.DecryptRsa(os.Getenv("JWT_SIGNING_KEY"), string(rsaPemByte))
	SubSystemCode = os.Getenv("SUB_SYSTEM_CODE")
	SubSystemKey = os.Getenv("SUB_SYSTEM_KEY")
	FiringCallback = os.Getenv("ALARM_FIRING_CALLBACK")
	RecoverCallback = os.Getenv("ALARM_RECOVER_CALLBACK")
	log.Println("read config file:", cfg, "successfully")
	if CoreUrl != "" && SubSystemCode != "" && SubSystemKey != "" {
		InitCoreToken()
	} else {
		log.Printf("Init core token fail,coreUrl & subSystemCode & subSystemKey can not empty")
	}
	SmsParamMaxLength, _ = strconv.Atoi(os.Getenv("MONITOR_SMS_PARAM_LENGTH"))
	initLocalTimeZone()
}

func initLocalTimeZone() {
	cmdOut, err := exec.Command("/bin/sh", "-c", "date|awk '{print $5}'").Output()
	if err != nil {
		log.Printf("init local time zone fail,%s \n", err.Error())
	} else {
		cmdOutString := strings.TrimSpace(string(cmdOut))
		if cmdOutString != "" {
			DefaultLocalTimeZone = cmdOutString
			log.Printf("init local time zone to %s \n", DefaultLocalTimeZone)
		} else {
			DefaultLocalTimeZone = "CST"
		}
	}
}
