package funcs

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ProcessObj struct {
	Pid        int       `json:"pid"`
	Guid       string    `json:"guid"`
	Name       string    `json:"name"`
	Port       int       `json:"port"`
	Cmd        string    `json:"cmd"`
	RunCmd     string    `json:"run_cmd"`
	StartTime  time.Time `json:"start_time"`
	StopTime   time.Time `json:"stop_time"`
	Retry      int       `json:"retry"`
	Path       string    `json:"path"`
	Status     string    `json:"status"`
	Deploy     bool      `json:"deploy"`
	ConfigHash string    `json:"config_hash"`
	Lock       *sync.RWMutex
	Process    *os.Process
}

func (p *ProcessObj) init(name, path, cmd string) {
	p.Lock = new(sync.RWMutex)
	p.Retry = 0
	p.Name = name
	p.Path = path
	p.Cmd = cmd
	if path != "" {
		p.Cmd = "cd " + p.Path + " && " + p.Cmd
	}
	p.RunCmd = p.Cmd
	p.Status = "dead"
	p.Pid = 0
}

func (p *ProcessObj) start(configFile, startFile, guid string, port int, param map[string]string) error {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	if guid != "" {
		p.Guid = guid
	}
	if p.Guid != "" {
		p.Deploy = true
	}
	if port > 0 {
		p.Port = port
	}
	if param == nil {
		param = make(map[string]string)
	}
	param["abs_path"] = p.Path
	if configFile != "" {
		err := replaceParam(configFile, param)
		if err != nil {
			return err
		}
	}
	if startFile != "" {
		err := replaceParam(startFile, param)
		if err != nil {
			return err
		}
	}
	if len(param) > 0 {
		p.RunCmd = p.Cmd
		for k, v := range param {
			if strings.Contains(p.RunCmd, fmt.Sprintf("{{%s}}", k)) {
				p.RunCmd = strings.Replace(p.RunCmd, fmt.Sprintf("{{%s}}", k), v, -1)
			}
		}
	}
	cmd := exec.Command(osBashCommand, "-c", p.RunCmd)
	err := cmd.Run()
	if err != nil {
		return err
	}
	p.Process = cmd.Process
	var pids []int
	for i := 0; i < 5; i++ {
		pids = getSystemProcessPid("", p.Path)
		if len(pids) > 0 {
			break
		}
		time.Sleep(time.Second * time.Duration(1))
	}
	time.Sleep(time.Second * time.Duration(1))
	if len(pids) > 0 {
		p.Pid = pids[0]
	} else {
		//p.Pid = p.Process.Pid
		p.Status = "stop"
		return fmt.Errorf("start timeout")
	}
	//cErr := checkExporterAlive(p.Name, p.Port)
	//if cErr != nil {
	//	log.Printf("%s is broken \n", p.Name)
	//	p.Status = "broken"
	//	return cErr
	//}
	p.Status = "running"
	p.StartTime = time.Now()
	log.Printf("run: %s done\n", p.RunCmd)
	return nil
}

func (p *ProcessObj) stop() error {
	log.Println("start stop")
	p.Lock.Lock()
	defer p.Lock.Unlock()
	if p.Pid > 0 {
		log.Printf("try to stop pid %d... \n", p.Pid)
		err := exec.Command(osBashCommand, "-c", fmt.Sprintf("kill -9 %d", p.Pid)).Run()
		if err != nil {
			return err
		}
		log.Printf("stop pid %d done \n", p.Pid)
		p.Pid = 0
		p.Process = nil
		p.Status = "stop"
		p.StopTime = time.Now()
	}
	log.Println("stop done")
	return nil
}

func (p *ProcessObj) restart() error {
	err := p.stop()
	if err != nil {
		return err
	}
	err = p.start("", "", "", 0, make(map[string]string))
	return err
}

func (p *ProcessObj) print() string {
	p.Lock.RLock()
	result := fmt.Sprintf("{\"pid\":%d,\"guid\":\"%s\",\"port\":%d,\"name\":\"%s\",\"cmd\":\"%s\",\"run_cmd\":\"%s\",\"path\":\"%s\",\"status\":\"%s\"}", p.Pid, p.Guid, p.Port, p.Name, p.Cmd, p.RunCmd, p.Path, p.Status)
	p.Lock.RUnlock()
	return result
}

func (p *ProcessObj) message() (pid int, n string, status string, retry int) {
	p.Lock.RLock()
	pid = p.Pid
	n = p.Name
	status = p.Status
	retry = p.Retry
	p.Lock.RUnlock()
	return pid, n, status, retry
}

func (p *ProcessObj) update(signal int) {
	p.Lock.Lock()
	if signal == 1 {
		pids := getSystemProcessPid(p.Name, "")
		if len(pids) > 0 {
			p.Pid = pids[0]
		}
	} else if signal == 2 {
		p.Status = "dead"
	} else if signal == 3 {
		p.Retry = p.Retry + 1
	}
	p.Lock.Unlock()
}

func (p *ProcessObj) destroy() error {
	var err error
	if p.Status == "running" || p.Status == "broken" {
		err = p.stop()
		if err != nil {
			return err
		}
	}
	p.Pid = 0
	p.Guid = ""
	p.Name = ""
	p.Status = ""
	p.Cmd = ""
	p.RunCmd = ""
	p.Path = ""
	p.Process = nil
	p.Lock = nil
	return nil
}

func getSystemProcessPid(name, path string) []int {
	//log.Printf("name : %s \n", name)
	result := []int{}
	cmdString := "ps aux|grep -v '\\['|awk '{print " + osPsPidIndex + "}'"
	if name != "" {
		cmdString = fmt.Sprintf("ps a|grep %s|grep -v 'bash'|grep -v 'grep'|grep -v 'nohup'|grep -v 'start.sh'|awk '{print $1}'", name)
	}
	if path != "" {
		cmdString = fmt.Sprintf("ps aux|grep %s|grep -v 'bash'|grep -v 'grep'|grep -v 'nohup'|grep -v 'start.sh'|awk '{print "+osPsPidIndex+"}'", path)
	}
	//log.Println(cmdString)
	b, err := exec.Command(osBashCommand, "-c", cmdString).Output()
	if err != nil {
		log.Printf("get system process pid fail with command %s : %v \n", cmdString, err)
		return result
	}
	for _, v := range strings.Split(string(b), "\n") {
		if v != "" {
			findList := regexp.MustCompile(`(\d+)`).FindAllString(v, -1)
			if len(findList) > 0 {
				tmpPid, _ := strconv.Atoi(findList[0])
				if tmpPid > 0 {
					result = append(result, tmpPid)
				}
			}
		}
	}
	return result
}

func replaceParam(filePath string, paramMap map[string]string) error {
	log.Printf("filePath: %s \n", filePath)
	_, err := os.Stat(filePath)
	if os.IsExist(err) {
		return err
	}
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	configString := string(b)
	for k, v := range paramMap {
		if strings.Contains(configString, fmt.Sprintf("{{%s}}", k)) {
			configString = strings.Replace(configString, fmt.Sprintf("{{%s}}", k), v, -1)
		}
	}
	err = ioutil.WriteFile(filePath, []byte(configString), 0644)
	if err != nil {
		return err
	}
	return nil
}

func checkExporterAlive(name string, port int) error {
	scrapeKey := ""
	if strings.Contains(name, "mysql") {
		scrapeKey = "mysql_exporter_last_scrape_error"
	}
	if strings.Contains(name, "redis") {
		scrapeKey = "redis_exporter_last_scrape_error"
	}
	if strings.Contains(name, "tomcat") {
		scrapeKey = "jmx_scrape_error"
	}
	if scrapeKey == "" {
		return nil
	}
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/metrics", port))
	if err != nil {
		fmt.Printf("http get error %v \n", err)
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		fmt.Printf("read body error %v \n", err)
		return err
	}
	if resp.StatusCode/100 != 2 {
		fmt.Printf("response http code %v \n", resp.StatusCode)
		return err
	}
	for _, v := range strings.Split(string(body), "\n") {
		if strings.Contains(v, scrapeKey) {
			if strings.Contains(v, "1") {
				if !strings.Contains(v, "#") {
					log.Printf("scrape : %s \n", v)
					return fmt.Errorf("scrape error,please check connect param")
				}
			}
		}
	}
	return nil
}
