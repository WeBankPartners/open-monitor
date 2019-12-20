package funcs

import (
	"time"
	"sync"
	"os"
	"io/ioutil"
	"strings"
	"fmt"
	"os/exec"
	"log"
	"strconv"
)

type ProcessObj  struct {
	Pid  int  `json:"pid"`
	Name  string  `json:"name"`
	Cmd  string  `json:"cmd"`
	RunCmd  string  `json:"run_cmd"`
	StartTime  time.Time  `json:"start_time"`
	StopTime   time.Time  `json:"stop_time"`
	Retry  int  `json:"retry"`
	Path  string  `json:"path"`
	Status  string  `json:"status"`
	Lock  *sync.RWMutex
	Process  *os.Process
}

func (p *ProcessObj)init(name,path,cmd string)  {
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

func (p *ProcessObj)start(configFile,startFile string,param map[string]string) error {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	for k,v := range param {
		fmt.Printf("start param k:%s v:%s \n", k, v)
	}
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
	cmd := exec.Command("bash", "-c", p.RunCmd)
	err := cmd.Run()
	if err != nil {
		return err
	}
	p.Process = cmd.Process
	pids := getSystemProcessPid(p.Name)
	if len(pids) > 0 {
		p.Pid = pids[0]
	}else {
		pids = getSystemProcessPid(p.RunCmd)
		if len(pids) > 0 {
			p.Pid = pids[0]
		}else {
			p.Pid = p.Process.Pid
		}
	}
	p.Status = "running"
	return nil
}

func (p *ProcessObj)stop() error {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	if p.Pid > 0 && p.Process != nil {
		err := p.Process.Kill()
		if err != nil {
			return err
		}
		p.Pid = 0
		p.Process = nil
		p.Status = "stop"
	}
	return nil
}

func (p *ProcessObj)restart() error {
	err := p.stop()
	if err != nil {
		return err
	}
	err = p.start("","", make(map[string]string))
	return err
}

func (p *ProcessObj)print() string {
	p.Lock.RLock()
	result := fmt.Sprintf("{\"pid\":%d,\"name\":\"%s\",\"cmd\":\"%s\",\"run_cmd\":\"%s\",\"start_time\":%d,\"stop_time\":%d,\"path\":\"%s\",\"status\":\"%s\"}", p.Pid,p.Name,p.Cmd,p.RunCmd,p.StartTime.Unix(),p.StopTime.Unix(),p.Path,p.Status)
	p.Lock.RUnlock()
	return result
}

func (p *ProcessObj)message() (pid int,n string,status string,retry int) {
	p.Lock.RLock()
	pid = p.Pid
	n = p.Name
	status = p.Status
	retry = p.Retry
	p.Lock.RUnlock()
	return pid,n,status,retry
}

func (p *ProcessObj)update(signal int) {
	p.Lock.Lock()
	if signal == 1 {
		pids := getSystemProcessPid(p.Name)
		if len(pids) == 0 {
			pids = getSystemProcessPid(p.RunCmd)
		}
		if len(pids) > 0 {
			p.Pid = pids[0]
		}
	}else if signal == 2 {
		p.Status = "dead"
	}else if signal == 3 {
		p.Retry = p.Retry + 1
	}
	p.Lock.Unlock()
}

func (p *ProcessObj)destroy() error {
	var err error
	if p.Status == "running" {
		err = p.stop()
		if err != nil {
			return err
		}
	}
	p.Pid = 0
	p.Name = ""
	p.Status = ""
	p.Cmd = ""
	p.RunCmd = ""
	p.Path = ""
	p.Process = nil
	p.Lock = nil
	return nil
}

func getSystemProcessPid(name string) []int {
	result := []int{}
	cmdString := "ps aux|grep -v '\\['|awk '{print $2}'"
	if name != "" {
		cmdString = fmt.Sprintf("ps aux|grep %s|grep -v 'bash'|grep -v 'grep'|awk '{print $2}'", name)
	}
	b,err := exec.Command("bash", "-c", cmdString).Output()
	if err != nil {
		log.Printf("get system process pid fail : %v \n", err)
		return result
	}
	for _,v := range strings.Split(string(b), "\n") {
		if v != "" {
			tmpPid,_ := strconv.Atoi(v)
			if tmpPid > 0 {
				result = append(result, tmpPid)
			}
		}
	}
	return result
}

func replaceParam(filePath string,paramMap map[string]string) error {
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
	for k,v := range paramMap {
		fmt.Printf("param k:%s  v:%s \n", k ,v)
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