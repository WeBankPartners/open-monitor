package api

import (
	"encoding/json"
	"github.com/WeBankPartners/open-monitor/monitor-server/api/v1/k8smonitor"
	"github.com/gin-gonic/gin"
	"os"
	"strings"
	"testing"

	"github.com/WeBankPartners/open-monitor/monitor-server/api/v1/user"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
)

func TestKubernetesMenuContainsAuthenticatedRoutes(t *testing.T) {
	raw, err := os.ReadFile("../conf/menu-api-map.json")
	if err != nil {
		t.Fatal(err)
	}
	var menus []*models.MenuApiMapObj
	if err = json.Unmarshal(raw, &menus); err != nil {
		t.Fatal(err)
	}
	oldMenus, oldMap := models.MenuApiGlobalList, user.ApiMenuMap
	t.Cleanup(func() { models.MenuApiGlobalList, user.ApiMenuMap = oldMenus, oldMap })
	models.MenuApiGlobalList = menus
	user.ApiMenuMap = make(map[string][]string)
	previousCodes := apiCodeMap
	t.Cleanup(func() { apiCodeMap = previousCodes })
	apiCodeMap = make(map[string]string)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	registerKubernetesRoutes(router, "/monitor", k8smonitor.Handler{}, k8smonitor.Handler{})
	codes := apiCodeMap
	user.InitApiMenuMap(codes)
	for _, route := range router.Routes() {
		if !strings.HasPrefix(route.Path, "/monitor/api/v1/k8s/") {
			continue
		}
		key := route.Method + "_" + route.Path
		code := codes[key]
		if code == "" {
			t.Errorf("route %s has no API code", key)
			continue
		}
		if strings.Contains(route.Path, "/plugin/") {
			continue
		} // Existing SUB_SYSTEM authentication.
		mapped := false
		for _, menu := range user.ApiMenuMap[code] {
			if menu == "MONITOR_K8S_CONFIG" {
				mapped = true
			}
		}
		if !mapped {
			t.Errorf("route %s has no Kubernetes menu permission mapping", key)
		}
	}
	for key := range codes {
		found := false
		for _, route := range router.Routes() {
			if route.Method+"_"+route.Path == key {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("API code %s has no registered route", key)
		}
	}
}
