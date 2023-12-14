package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	configFile := flag.String("c", "config.json", "config json file")
	flag.Parse()
	configBytes, configErr := ioutil.ReadFile(*configFile)
	if configErr != nil {
		log.Printf("read config file fail,%s ", configErr.Error())
		return
	}
	dpList := []*DaemonProcConfig{}
	if err := json.Unmarshal(configBytes, &dpList); err != nil {
		log.Printf("json unmarshal config file fail,%s ", err.Error())
		return
	}
	if len(dpList) == 0 {
		log.Println("config file is empty,done")
		return
	}
	for _, v := range dpList {
		dp := DaemonProc{}
		dp.Init(v)
		dp.Start()
	}
	WaitProcessSignal()
}

func WaitProcessSignal() {
	sg := make(chan os.Signal)
	signal.Notify(sg, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)
	s := <-sg
	log.Printf("get signal %s \n", s.String())
}

type DaemonProcConfig struct {
	Name      string   `json:"name"`
	Args      []string `json:"args"`
	MaxTry    int      `json:"maxTry"`
	WorkDir   string   `json:"workDir"`
	StdOutLog string   `json:"stdOutLog"`
	LocalBin  bool     `json:"localBin"`
	WithBash  bool     `json:"withBash"`
}

func (d *DaemonProcConfig) CommandLine() string {
	cline := ""
	if !strings.HasSuffix(d.WorkDir, "/") {
		d.WorkDir = d.WorkDir + "/"
	}
	if d.WorkDir != "/" {
		cline += fmt.Sprintf("cd %s && ", d.WorkDir)
	}
	if d.LocalBin {
		cline += fmt.Sprintf("./%s", d.Name)
	} else {
		cline += d.Name
	}
	if len(d.Args) > 0 {
		cline += " " + strings.Join(d.Args, " ")
	}
	if d.StdOutLog != "" {
		cline += fmt.Sprintf(" >> %s 2>&1", d.StdOutLog)
	}
	return cline
}

type DaemonProc struct {
	procName      string
	procArgs      []string
	createTime    time.Time
	startTime     time.Time
	cmdString     string
	currentCmd    *exec.Cmd
	stopChan      chan int
	waitExistChan chan int
	err           error
	daemonRunning bool
	maxTry        int
	stdOutLog     string
	withBash      bool
}

func (p *DaemonProc) Init(dpConfig *DaemonProcConfig) {
	p.cmdString = dpConfig.CommandLine()
	p.procName = dpConfig.Name
	p.procArgs = dpConfig.Args
	p.maxTry = dpConfig.MaxTry
	p.stdOutLog = dpConfig.StdOutLog
	p.withBash = dpConfig.WithBash
	log.Printf("cmdString : %s \n", p.cmdString)
	p.createTime = time.Now()
	p.stopChan = make(chan int, 1)
	p.waitExistChan = make(chan int, 1)
}

func (p *DaemonProc) Start() {
	go p.startDaemon()
}

func (p *DaemonProc) startDaemon() {
	p.daemonRunning = true
	stopFlag := false
	execCount := 0
	log.Println("daemon start")
	for {
		go p.execProc()
		select {
		case <-p.stopChan:
			stopFlag = true
			p.currentCmd.Process.Kill()
		case <-p.waitExistChan:
			if p.err != nil {
				log.Printf("err: %s \n", p.err.Error())
				stopFlag = true
			} else {
				execCount = execCount + 1
				if p.maxTry > 0 && p.maxTry < execCount {
					stopFlag = true
				}
			}
		}
		if stopFlag {
			break
		}
	}
	p.daemonRunning = false
	log.Println("daemon end")
}

func (p *DaemonProc) execProc() {
	p.startTime = time.Now()
	if p.withBash {
		p.currentCmd = exec.Command("bash", "-c", p.cmdString)
	} else {
		p.currentCmd = exec.Command("sh", "-c", p.cmdString)
	}
	defer func() {
		p.waitExistChan <- 1
	}()
	if err := p.currentCmd.Start(); err != nil {
		p.err = fmt.Errorf("cmd %s start fail,%s ", p.procName, err.Error())
		return
	}
	log.Printf("start exec proc: %s, pid: %d \n", p.procName, p.currentCmd.Process.Pid)
	pcState, pcErr := p.currentCmd.Process.Wait()
	if pcErr != nil {
		p.err = fmt.Errorf("cmd %s wait process running fail,%s ", p.procName, pcErr.Error())
		return
	}
	log.Printf("proc %s wait return,stateCode: %d \n", p.procName, pcState.ExitCode())
}

func (p *DaemonProc) Stop() {
	if p.daemonRunning {
		p.stopChan <- 1
	}
}

func (p *DaemonProc) Status() string {
	var status, pid, startTime, runningTime string
	if !p.daemonRunning {
		status = "stop"
	} else {
		status = "running"
		pid = fmt.Sprintf("%d", p.currentCmd.Process.Pid)
		startTime = p.startTime.Format(time.RFC3339)
		runningTime = fmt.Sprintf("%.0fs", time.Now().Sub(p.startTime).Seconds())
	}
	return fmt.Sprintf("proc: %s | args: %s | status: %s | pid: %s | createTime: %s | startTime: %s | runningTime: %s ", p.procName, p.procArgs, status, pid, p.createTime.Format(time.RFC3339), startTime, runningTime)
}
