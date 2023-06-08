package manager

import (
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-agent/agent_manager/api/v1/redirect"
	"github.com/WeBankPartners/open-monitor/monitor-agent/agent_manager/funcs"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

func AddDeploy(w http.ResponseWriter, r *http.Request) {
	var resp funcs.HttpResponse
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("error : %v \n", err)
		resp.Code = 500
		resp.Message = fmt.Sprintf("error:%v", err)
	} else {
		var tmpParamMap map[string]string
		err = json.Unmarshal(b, &tmpParamMap)
		if err != nil {
			resp.Code = 500
			resp.Message = fmt.Sprintf("error:%v", err)
		} else {
			var exporter, configFile, guid string
			if _, b := tmpParamMap["guid"]; !b {
				resp.Code = 400
				resp.Message = "param guid can not find!"
			}
			guid = tmpParamMap["guid"]
			if v, b := tmpParamMap["exporter"]; !b {
				resp.Code = 400
				resp.Message = "param exporter can not find!"
			} else {
				exporter = v
				if v, b := tmpParamMap["config"]; b {
					configFile = v
					if !illegalPath(configFile) {
						resp.Code = 400
						resp.Message = "param config illegal "
					}
				}
				if !illegalName(exporter) {
					resp.Code = 400
					resp.Message = "param exporter illegal "
				}
				if !illegalGuid(guid) {
					resp.Code = 400
					resp.Message = "param guid illegal "
				}
				if resp.Code < 200 {
					var port int
					if tmpParamMap["agentManagerRemoteIp"] != "" {
						port, err = redirect.Add(tmpParamMap["agentManagerRemoteIp"], tmpParamMap)
					} else {
						configHash := fmt.Sprintf("%s:%s_%s_%s", tmpParamMap["instance_server"], tmpParamMap["instance_port"], tmpParamMap["auth_user"], tmpParamMap["auth_password"])
						port, err = funcs.AddDeploy(exporter, configFile, guid, tmpParamMap, configHash)
					}
					if err != nil {
						resp.Code = 500
						resp.Message = fmt.Sprintf("error:%v", err)
					} else {
						resp.Code = 200
						if port > 0 {
							resp.Message = fmt.Sprintf("%s:%d", funcs.LocalIp, port)
						} else {
							resp.Message = "exist"
						}
					}
				}
			}
		}
	}
	log.Printf("response code:%d message:%s \n", resp.Code, resp.Message)
	funcs.SaveDeployProcess()
	w.Write(resp.Byte())
}

func DelDeploy(w http.ResponseWriter, r *http.Request) {
	var resp funcs.HttpResponse
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("error : %v \n", err)
		resp.Code = 500
		resp.Message = fmt.Sprintf("error:%v", err)
	} else {
		var tmpParamMap map[string]string
		err = json.Unmarshal(b, &tmpParamMap)
		if err != nil {
			resp.Code = 500
			resp.Message = fmt.Sprintf("error:%v", err)
		} else {
			if v, b := tmpParamMap["guid"]; b {
				if tmpParamMap["agentManagerRemoteIp"] != "" {
					err = redirect.Delete(tmpParamMap["agentManagerRemoteIp"], tmpParamMap)
				} else {
					err = funcs.DeleteDeploy(v)
				}
				if err != nil {
					resp.Code = 500
					resp.Message = fmt.Sprintf("error:%v", err)
				} else {
					resp.Code = 200
					resp.Message = "success"
				}
			} else {
				resp.Code = 400
				resp.Message = "Param guid not exist"
			}
		}
	}
	funcs.SaveDeployProcess()
	w.Write(resp.Byte())
}

func InitDeploy(w http.ResponseWriter, r *http.Request) {
	log.Println("start init deploy")
	var resp funcs.HttpResponse
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("error : %v \n", err)
		resp.Code = 500
		resp.Message = fmt.Sprintf("error:%v", err)
	} else {
		var param funcs.InitDeployParam
		err = json.Unmarshal(b, &param)
		if err != nil {
			resp.Code = 500
			resp.Message = fmt.Sprintf("error:%v", err)
		} else {
			if param.AgentManagerRemoteIp != "" {
				err = redirect.Init(&param)
			} else {
				err = funcs.InitDeployDir(param.Config)
			}
			if err != nil {
				resp.Code = 500
				resp.Message = fmt.Sprintf("error:%v", err)
			} else {
				resp.Code = 200
				resp.Message = "success"
			}
		}
	}
	w.Write(resp.Byte())
}

func DisplayProcess(w http.ResponseWriter, r *http.Request) {
	w.Write(funcs.PrintProcessList())
}

func illegalPath(input string) bool {
	if input == "" {
		return true
	}
	var regPath = regexp.MustCompile(`^\/?([\w|\.|\-]+\/?)+$`)
	return regPath.MatchString(input)
}

func illegalName(input string) bool {
	if input == "" {
		return true
	}
	var regPath = regexp.MustCompile(`^\w+$`)
	return regPath.MatchString(input)
}

func illegalGuid(input string) bool {
	if input == "" {
		return true
	}
	var regPath = regexp.MustCompile(`^[\w|\.|\-:\*@!~(){}\[\]=,]+$`)
	return regPath.MatchString(input)
}
