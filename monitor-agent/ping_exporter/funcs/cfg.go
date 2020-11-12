package funcs

import (
	"bufio"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type TransferConfig struct {
	Addrs    []string `json:"addrs"`
	Interval int      `json:"interval"`
	Timeout  int      `json:"timeout"`
	Sn       int      `json:"sn"`
}

type OpenFalconConfig struct {
	Enabled  bool     `json:"enabled"`
	Transfer      *TransferConfig   `json:"transfer"`
}

type PrometheusCOnfig struct {
	Enabled  bool     `json:"enabled"`
	Port  string  `json:"port"`
	Path  string  `json:"path"`
}

type SourceConfig struct {
	Const  SourceConstConfig  `json:"const"`
	File  SourceFileConfig  `json:"file"`
	Remote  SourceRemoteConfig  `json:"remote"`
	Listen  SourceListenConfig  `json:"listen"`
}

type SourceConstConfig struct {
	Enabled  bool     `json:"enabled"`
	Ips  []string  `json:"ips"`
	Weight  int  `json:"weight"`
}

type SourceFileConfig struct {
	Enabled  bool     `json:"enabled"`
	Path  string  `json:"path"`
	Weight  int  `json:"weight"`
}

type SourceRemoteConfig  struct {
	Enabled  bool     `json:"enabled"`
	Header  []string  `json:"header"`
	GroupTag  string  `json:"group_tag"`
	Url  string  `json:"url"`
	Interval  int  `json:"interval"`
	Weight  int  `json:"weight"`
}

type SourceListenConfig  struct {
	Enabled  bool     `json:"enabled"`
	Port  string  `json:"port"`
	Path  string  `json:"path"`
	Weight  int  `json:"weight"`
}

type MetricConfig  struct {
	Ping  string  `json:"ping"`
	PingUseTime  string  `json:"ping_use_time"`
	PingCountNum  string  `json:"ping_count_num"`
	PingCountSuccess  string  `json:"ping_count_success"`
	PingCountFail  string  `json:"ping_count_fail"`
	Telnet  string  `json:"telnet"`
	TelnetCountNum  string  `json:"telnet_count_num"`
	TelnetCountSuccess  string  `json:"telnet_count_success"`
	TelnetCountFail  string  `json:"telnet_count_fail"`
	HttpCheck  string  `json:"http_check"`
	HttpCheckCountNum  string  `json:"http_check_count_num"`
	HttpCheckCountSuccess  string  `json:"http_check_count_success"`
	HttpCheckCountFail  string  `json:"http_check_count_fail"`
}

type GlobalConfig struct {
	Debug         bool              `json:"debug"`
	Interval      int               `json:"interval"`
	PingEnable    bool              `json:"ping_enable"`
	TelnetEnable  bool              `json:"telnet_enable"`
	HttpCheckEnable  bool           `json:"http_check_enable"`
	HttpProxyEnable  bool           `json:"http_proxy_enable"`
	HttpProxyAddress  string           `json:"http_proxy"`
	OpenFalcon    OpenFalconConfig  `json:"open-falcon"`
	Prometheus    PrometheusCOnfig  `json:"prometheus"`
	Source      SourceConfig    `json:"source"`
	Metrics       MetricConfig      `json:"metrics"`
}

var (
	ConfigFile string
	config     *GlobalConfig
	lock       = new(sync.RWMutex)
)

func Config() *GlobalConfig {
	lock.RLock()
	defer lock.RUnlock()
	return config
}

func ParseConfig(cfg string) error {
	if cfg == "" {
		log.Fatalln("use -c to specify configuration file")
	}
	_, err := os.Stat(cfg)
	if os.IsExist(err) {
		log.Fatalln("config file not found")
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
		log.Fatalln("parse config file:", cfg, "fail:", err)
		return err
	}
	lock.Lock()
	defer lock.Unlock()
	config = &c
	log.Println("read config file:", cfg, "successfully")
	return nil
}

func Uuid() (string) {
	commandName := "/usr/sbin/dmidecode"
	params := []string{"|", "grep UUID"}
	cmd := exec.Command(commandName, params...)
	//显示运行的命令
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalln(os.Stderr, "error=>", err.Error())
		return ""
	}
	cmd.Start() // Start开始执行c包含的命令，但并不会等待该命令完成即返回。Wait方法会返回命令的返回状态码并在命令返回后释放相关的资源。

	reader := bufio.NewReader(stdout)

	var index int
	var uuid string
	//实时循环读取输出流中的一行内容
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		s := strings.Split(line, ":")
		t := strings.TrimSpace(s[0])
		if t == "UUID" {
			uuid = strings.TrimSpace(s[1])
		}
		index++
	}
	cmd.Wait()
	return uuid
}