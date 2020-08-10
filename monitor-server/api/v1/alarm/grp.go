package alarm

import (
	"github.com/gin-gonic/gin"
	"strconv"

	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"fmt"
	"strings"
	"net/http"
	"encoding/json"
	"io/ioutil"
	"time"
)

func ListGrp(c *gin.Context)  {
	id,_ := strconv.Atoi(c.Query("id"))
	search := c.Query("search")
	page,_ := strconv.Atoi(c.Query("page"))
	size,_ := strconv.Atoi(c.Query("size"))
	if page <= 0 || size <= 0 {
		mid.ReturnParamEmptyError(c, "page and size")
		return
	}
	query := m.GrpQuery{Id:id, Search:search, Page:page, Size:size}
	err := db.ListGrp(&query)
	if err != nil {
		mid.ReturnQueryTableError(c, "grp", err)
		return
	}
	mid.ReturnSuccessData(c, m.TableData{Data:query.Result, Num:query.ResultNum, Page:page, Size:size})
}

func AddGrp(c *gin.Context)  {
	var param m.GrpTable
	if err := c.ShouldBindJSON(&param); err==nil {
		if mid.IsIllegalName(param.Name) {
			mid.ReturnValidateError(c, "illegal name")
			return
		}
		query := m.GrpQuery{Name:param.Name}
		db.ListGrp(&query)
		if len(query.Result) > 0 {
			mid.ReturnValidateError(c, "name exists")
			return
		}
		operateUser := mid.GetOperateUser(c)
		err := db.UpdateGrp(&m.UpdateGrp{Groups:[]*m.GrpTable{&param}, Operation:"insert", OperateUser:operateUser})
		_,grpObj := db.GetSingleGrp(0, param.Name)
		if err != nil || grpObj.Id <= 0 {
			mid.ReturnUpdateTableError(c, "grp", err)
		}else{
			db.AddTpl(grpObj.Id,0, operateUser)
			mid.ReturnSuccess(c)
		}
	}else{
		mid.ReturnValidateError(c, err.Error())
	}
}

func UpdateGrp(c *gin.Context)  {
	var param m.GrpTable
	if err := c.ShouldBindJSON(&param); err==nil || param.Id <= 0 {
		if mid.IsIllegalName(param.Name) {
			mid.ReturnValidateError(c, "illegal name")
			return
		}
		query := m.GrpQuery{Name:param.Name}
		db.ListGrp(&query)
		if len(query.Result) > 0 {
			if query.Result[0].Id == param.Id && query.Result[0].Description == param.Description {
				mid.ReturnSuccess(c)
				return
			}
			if query.Result[0].Id != param.Id {
				mid.ReturnValidateError(c, "name exists")
				return
			}
		}
		err := db.UpdateGrp(&m.UpdateGrp{Groups:[]*m.GrpTable{&param}, Operation:"update", OperateUser:mid.GetOperateUser(c)})
		if err != nil {
			mid.ReturnUpdateTableError(c, "grp", err)
		}else{
			mid.ReturnSuccess(c)
		}
	}else{
		mid.ReturnValidateError(c, err.Error())
	}
}

func DeleteGrp(c *gin.Context)  {
	id,_ := strconv.Atoi(c.Query("id"))
	if id <= 0 {
		mid.ReturnParamTypeError(c, "id", "int")
		return
	}
	_,tplObj := db.GetTpl(0, id, 0)
	if tplObj.Id > 0 {
		db.DeleteStrategyByGrp(0, tplObj.Id)
		err := SaveConfigFile(tplObj.Id, false)
		if err != nil {
			mid.ReturnHandleError(c, "update prometheus config file fail", err)
			return
		}
		db.DeleteTpl(tplObj.Id)
	}
	db.DeleteStrategyByGrp(id, 0)
	err := db.UpdateGrp(&m.UpdateGrp{Groups:[]*m.GrpTable{&m.GrpTable{Id:id}}, Operation:"delete", OperateUser:mid.GetOperateUser(c)})
	if err != nil {
		mid.ReturnUpdateTableError(c, "grp", err)
	}else{
		mid.ReturnSuccess(c)
	}
}

func ListEndpoint(c *gin.Context)  {
	search := c.Query("search")
	page,_ := strconv.Atoi(c.Query("page"))
	size,_ := strconv.Atoi(c.Query("size"))
	grp,_ := strconv.Atoi(c.Query("grp"))
	if page <= 0 || size <= 0 {
		mid.ReturnParamEmptyError(c, "page and size")
		return
	}
	query := m.AlarmEndpointQuery{Search:search, Page:page, Size:size, Grp:grp}
	err := db.ListAlarmEndpoints(&query)
	if err != nil {
		mid.ReturnQueryTableError(c, "alarm endpoints", err)
		return
	}
	mid.ReturnSuccessData(c, m.TableData{Data:query.Result, Num:query.ResultNum, Page:page, Size:size})
}

func EditGrpEndpoint(c *gin.Context)  {
	var param m.GrpEndpointParamNew
	if err := c.ShouldBindJSON(&param); err==nil {
		if param.Operation != "add" && param.Operation != "delete" {
			mid.ReturnValidateError(c, "operation must be add or delete")
			return
		}
		err,isUpdate := db.UpdateGrpEndpoint(param)
		if err != nil {
			mid.ReturnUpdateTableError(c, "grp endpoint", err)
			return
		}
		if isUpdate {
			err,tplObj := db.GetTpl(0, param.Grp, 0)
			if err != nil {
				mid.ReturnFetchDataError(c, "tpl", "grp_id", strconv.Itoa(param.Grp))
				return
			}
			if tplObj.Id <= 0 {
				err,tplObj = db.AddTpl(param.Grp, 0, mid.GetOperateUser(c))
				if err != nil {
					mid.ReturnUpdateTableError(c, "tpl", err)
					return
				}
			}
			err = SaveConfigFile(tplObj.Id, false)
			if err != nil {
				mid.ReturnHandleError(c, "save configuration file failed", err)
				return
			}
		}
		mid.ReturnSuccess(c)
	}else{
		mid.ReturnValidateError(c, err.Error())
	}
}

func ExportGrpStrategy(c *gin.Context)  {
	idStringList := strings.Split(c.Query("id"), ",")
	var idList []string
	for _,v := range idStringList {
		tmpId,_ := strconv.Atoi(v)
		if tmpId > 0 {
			idList = append(idList, v)
		}
	}
	if len(idList) == 0 {
		mid.ReturnParamEmptyError(c, "id")
		return
	}
	err,result := db.GetGrpStrategy(idList)
	if err != nil {
		mid.ReturnQueryTableError(c, "grp strategy", err)
		return
	}

	b,err := json.Marshal(result)
	if err != nil {
		mid.ReturnHandleError(c, "export group strategy fail, json marshal object error", err)
		return
	}
	c.Writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s_%s.json", "monitor_group_", time.Now().Format("20060102150405")))
	c.Data(http.StatusOK, "application/octet-stream", b)
}

func ImportGrpStrategy(c *gin.Context)  {
	file,err := c.FormFile("file")
	if err != nil {
		mid.ReturnValidateError(c, err.Error())
		return
	}
	f,err := file.Open()
	if err != nil {
		mid.ReturnHandleError(c, "file open error ", err)
		return
	}
	var paramObj []*m.GrpStrategyExportObj
	b,err := ioutil.ReadAll(f)
	defer f.Close()
	if err != nil {
		mid.ReturnHandleError(c, "read content fail error ", err)
		return
	}
	err = json.Unmarshal(b, &paramObj)
	if err != nil {
		mid.ReturnHandleError(c, "json unmarshal fail error ", err)
		return
	}
	err = db.SetGrpStrategy(paramObj)
	if err != nil {
		mid.ReturnHandleError(c, "save group strategy error", err)
		return
	}
	mid.ReturnSuccess(c)
}

func UpdateGrpRole(c *gin.Context)  {
	var param m.RoleGrpDto
	if err := c.ShouldBindJSON(&param); err==nil {
		err = db.UpdateGrpRole(param)
		if err != nil {
			mid.ReturnUpdateTableError(c, "grp role", err)
		}else{
			mid.ReturnSuccess(c)
		}
	}else{
		mid.ReturnValidateError(c, err.Error())
	}
}

func GetGrpRole(c *gin.Context)  {
	grpId,_ := strconv.Atoi(c.Query("grp_id"))
	if grpId <= 0 {
		mid.ReturnParamTypeError(c, "grp_id", "int")
		return
	}
	err,result := db.GetGrpRole(grpId)
	if err != nil {
		mid.ReturnFetchDataError(c, "rel_role_grp", "grp_id", strconv.Itoa(grpId))
	}else{
		mid.ReturnSuccessData(c, result)
	}
}