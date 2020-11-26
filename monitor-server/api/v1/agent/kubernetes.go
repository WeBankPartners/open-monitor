package agent

import (
	"github.com/gin-gonic/gin"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"strconv"
	"strings"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
)

func UpdateKubernetesCluster(c *gin.Context)  {
	var param m.KubernetesClusterParam
	operation := c.Param("operation")
	var err error
	if operation == "get" || operation == "list" {
		result,err := db.ListKubernetesCluster()
		if err != nil {
			mid.ReturnHandleError(c, err.Error(), err)
		}else{
			mid.ReturnSuccessData(c, result)
		}
		return
	}
	if operation == "delete" {
		var tmpParam m.KubernetesClusterTable
		if err = c.ShouldBindJSON(&tmpParam);err==nil {
			if param.Id <= 0 {
				mid.ReturnParamEmptyError(c, "id")
				return
			}
			err = db.DeleteKubernetesCluster(param.Id)
		}else{
			mid.ReturnValidateError(c, err.Error())
			return
		}
	}
	if err = c.ShouldBindJSON(&param);err==nil {
		if mid.IsIllegalIp(param.Ip) {
			mid.ReturnValidateError(c, "param ip is illegal")
			return
		}
		portInt,_ := strconv.Atoi(param.Port)
		if portInt <= 0 {
			mid.ReturnValidateError(c, "param port is illegal")
			return
		}
		param.ClusterName = strings.TrimSpace(param.ClusterName)
		if !mid.IsIllegalNormalInput(param.ClusterName) {
			mid.ReturnValidateError(c, "param cluster_name is illegal")
			return
		}
		if operation == "update" {
			if param.Id <= 0 {
				mid.ReturnValidateError(c, "param id is empty")
				return
			}
			err = db.UpdateKubernetesCluster(param)
		}else {
			err = db.AddKubernetesCluster(param)
		}
	}else{
		mid.ReturnValidateError(c, err.Error())
		return
	}
	if err != nil {
		mid.ReturnHandleError(c, err.Error(), err)
	}else{
		mid.ReturnSuccess(c)
	}
}