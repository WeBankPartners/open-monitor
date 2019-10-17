package alarm

import (
	"github.com/gin-gonic/gin"
	"strconv"

	mid "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/middleware"
	m "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/models"
	"github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/services/db"
	"fmt"
)

func ListGrp(c *gin.Context)  {
	id,_ := strconv.Atoi(c.Query("id"))
	search := c.Query("search")
	page,_ := strconv.Atoi(c.Query("page"))
	size,_ := strconv.Atoi(c.Query("size"))
	if page <= 0 || size <= 0 {
		mid.ReturnValidateFail(c, "page and size can't be none")
		return
	}
	query := m.GrpQuery{Id:id, Search:search, Page:page, Size:size}
	err := db.ListGrp(&query)
	if err != nil {
		mid.ReturnError(c, "get alarm group error", err)
		return
	}
	mid.ReturnData(c, m.TableData{Data:query.Result, Num:query.ResultNum, Page:page, Size:size})
}

func AddGrp(c *gin.Context)  {
	var param m.GrpTable
	if err := c.ShouldBindJSON(&param); err==nil {
		if mid.IsIllegalName(param.Name) {
			mid.ReturnValidateFail(c, "Name illegal")
			return
		}
		query := m.GrpQuery{Name:param.Name}
		db.ListGrp(&query)
		if len(query.Result) > 0 {
			mid.ReturnError(c, "Name exist", nil)
			return
		}
		err := db.UpdateGrp(&m.UpdateGrp{Groups:[]*m.GrpTable{&param}, Operation:"insert", OperateUser:""})
		if err != nil {
			mid.ReturnError(c, "Fail", err)
		}else{
			mid.ReturnSuccess(c, "Success")
		}
	}else{
		mid.ReturnValidateFail(c, "Param validate fail")
	}
}

func UpdateGrp(c *gin.Context)  {
	var param m.GrpTable
	if err := c.ShouldBindJSON(&param); err==nil || param.Id <= 0 {
		if mid.IsIllegalName(param.Name) {
			mid.ReturnValidateFail(c, "Name illegal")
			return
		}
		query := m.GrpQuery{Name:param.Name}
		db.ListGrp(&query)
		if len(query.Result) > 0 {
			if query.Result[0].Id == param.Id && query.Result[0].Description == param.Description {
				mid.ReturnSuccess(c, "same content")
				return
			}
			if query.Result[0].Id != param.Id {
				mid.ReturnError(c, "the group name already exist", nil)
				return
			}
		}
		err := db.UpdateGrp(&m.UpdateGrp{Groups:[]*m.GrpTable{&param}, Operation:"update", OperateUser:""})
		if err != nil {
			mid.ReturnError(c, "Fail", err)
		}else{
			mid.ReturnSuccess(c, "Success")
		}
	}else{
		mid.ReturnValidateFail(c, "Param validate fail")
	}
}

func DeleteGrp(c *gin.Context)  {
	id,_ := strconv.Atoi(c.Query("id"))
	if id <= 0 {
		mid.ReturnValidateFail(c,"Id is none")
		return
	}
	err := db.UpdateGrp(&m.UpdateGrp{Groups:[]*m.GrpTable{&m.GrpTable{Id:id}}, Operation:"delete", OperateUser:""})
	if err != nil {
		mid.ReturnError(c, "Fail", err)
	}else{
		mid.ReturnSuccess(c, "Success")
	}
}

func ListEndpoint(c *gin.Context)  {
	search := c.Query("search")
	page,_ := strconv.Atoi(c.Query("page"))
	size,_ := strconv.Atoi(c.Query("size"))
	grp,_ := strconv.Atoi(c.Query("grp"))
	if page <= 0 || size <= 0 {
		mid.ReturnValidateFail(c, "page and size can't be none")
		return
	}
	query := m.AlarmEndpointQuery{Search:search, Page:page, Size:size, Grp:grp}
	err := db.ListAlarmEndpoints(&query)
	if err != nil {
		mid.ReturnError(c, "get endpoints error", err)
		return
	}
	mid.ReturnData(c, m.TableData{Data:query.Result, Num:query.ResultNum, Page:page, Size:size})
}

func EditGrpEndpoint(c *gin.Context)  {
	var param m.GrpEndpointParamNew
	if err := c.ShouldBindJSON(&param); err==nil {
		if param.Operation != "add" && param.Operation != "delete" {
			mid.ReturnValidateFail(c, "operation must be add or delete")
			return
		}
		err,isUpdate := db.UpdateGrpEndpoint(param)
		if err != nil {
			mid.ReturnError(c, "update grp endpoint fail", err)
			return
		}
		if isUpdate {
			err,tplObj := db.GetTpl(0, param.Grp, 0)
			if err != nil || tplObj.Id <= 0 {
				mid.ReturnError(c, fmt.Sprintf("edit grp endpoint, get tpl fail with grp id:%d", param.Grp), err)
				return
			}
			err = SaveConfigFile(tplObj.Id)
			if err != nil {
				mid.ReturnError(c, "edit grp endpoint, save prometheus config file fail", err)
				return
			}
		}
		mid.ReturnSuccess(c, "Success")
	}else{
		mid.ReturnValidateFail(c, "Param validate fail")
	}
}

