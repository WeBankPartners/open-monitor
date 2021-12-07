package db

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"golang.org/x/net/context/ctxhttp"
	"io/ioutil"
	"net/http"
	"strings"
)

func SyncCoreRoleList()  {
	if m.CoreUrl == "" {
		return
	}
	request,err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/platform/v1/roles/retrieve", m.CoreUrl), strings.NewReader(""))
	if err != nil {
		log.Logger.Error("Get core role key new request fail", log.Error(err))
		return
	}
	request.Header.Set("Authorization", m.GetCoreToken())
	res,err := ctxhttp.Do(context.Background(), http.DefaultClient, request)
	if err != nil {
		log.Logger.Error("Get core role key ctxhttp request fail", log.Error(err))
		return
	}
	b,_ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	var result m.CoreRoleDto
	err = json.Unmarshal(b, &result)
	if err != nil {
		log.Logger.Error("Get core role key json unmarshal result", log.Error(err))
		return
	}
}
