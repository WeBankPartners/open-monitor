package api

import (
	"github.com/WeBankPartners/open-monitor/monitor-agent/agent_manager/api/v1/manager"
	"net/http"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-agent/agent_manager/funcs"
)

func InitHttpServer() {
	http.Handle("/deploy/add", http.HandlerFunc(manager.AddDeploy))
	http.Handle("/deploy/delete", http.HandlerFunc(manager.DelDeploy))
	http.Handle("/process/list", http.HandlerFunc(manager.DisplayProcess))
	http.ListenAndServe(fmt.Sprintf(":%d", funcs.Config().Http.Port), nil)
}