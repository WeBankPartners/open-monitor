package db

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	"reflect"
	"strconv"
	"strings"
	"time"
	"xorm.io/core"
	"xorm.io/xorm"
	xorm_log "xorm.io/xorm/log"
)

var (
	x               *xorm.Engine
	archiveMysql    *xorm.Engine
	archiveDatabase string
	ArchiveEnable   bool
)

//var RedisStore sessions.RedisStore

func InitDatabase() error {
	connStr := fmt.Sprintf("%s:%s@%s(%s)/%s?collation=utf8mb4_unicode_ci&allowNativePasswords=true",
		models.Config().Store.User, models.Config().Store.Pwd, "tcp", fmt.Sprintf("%s:%s", models.Config().Store.Server, models.Config().Store.Port), models.Config().Store.DataBase)
	engine, err := xorm.NewEngine("mysql", connStr)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Init database connect fail", zap.Error(err))
		return err
	}
	engine.SetMaxIdleConns(models.Config().Store.MaxIdle)
	engine.SetMaxOpenConns(models.Config().Store.MaxOpen)
	engine.SetConnMaxLifetime(time.Duration(models.Config().Store.Timeout) * time.Second)
	engine.SetLogger(&dbLogger{LogLevel: 1, ShowSql: true, Logger: log.DatabaseLogger})
	// 使用驼峰式映射
	engine.SetMapper(core.SnakeMapper{})
	x = engine
	log.Info(nil, log.LOGGER_APP, "Success init database connect !!")
	tmpEnable := strings.ToLower(models.Config().ArchiveMysql.Enable)
	if tmpEnable == "y" || tmpEnable == "yes" || tmpEnable == "true" {
		initArchiveDbEngine()
	} else {
		ArchiveEnable = false
	}
	return nil
}

func initArchiveDbEngine() {
	databaseName := models.Config().ArchiveMysql.DatabasePrefix + time.Now().Format("2006")
	var err error
	connectStr := fmt.Sprintf("%s:%s@%s(%s:%s)/%s?collation=utf8mb4_unicode_ci&allowNativePasswords=true",
		models.Config().ArchiveMysql.User, models.Config().ArchiveMysql.Password, "tcp", models.Config().ArchiveMysql.Server, models.Config().ArchiveMysql.Port, databaseName)
	archiveMysql, err = xorm.NewEngine("mysql", connectStr)
	if err != nil {
		ArchiveEnable = false
		log.Error(nil, log.LOGGER_APP, "Init archive mysql fail", zap.String("connectStr", connectStr), zap.Error(err))
	} else {
		ArchiveEnable = true
		archiveMysql.SetMaxIdleConns(models.Config().ArchiveMysql.MaxIdle)
		archiveMysql.SetMaxOpenConns(models.Config().ArchiveMysql.MaxOpen)
		archiveMysql.SetConnMaxLifetime(time.Duration(models.Config().ArchiveMysql.Timeout) * time.Second)
		archiveMysql.Charset("utf8")
		// 使用驼峰式映射
		archiveMysql.SetMapper(core.SnakeMapper{})
		archiveDatabase = databaseName
		log.Info(nil, log.LOGGER_APP, "Init archive mysql "+archiveDatabase+" success")
	}
}

func checkArchiveDatabase() {
	databaseName := models.Config().ArchiveMysql.DatabasePrefix + time.Now().Format("2006")
	if databaseName == archiveDatabase {
		return
	}
	initArchiveDbEngine()
}

type Action struct {
	Sql   string
	Param []interface{}
}

func Transaction(actions []*Action) error {
	if len(actions) == 0 {
		return fmt.Errorf("transaction actions is null")
	}
	session := x.NewSession()
	err := session.Begin()
	for _, action := range actions {
		params := make([]interface{}, 0)
		params = append(params, action.Sql)
		for _, v := range action.Param {
			params = append(params, v)
		}
		_, err = session.Exec(params...)
		if err != nil {
			session.Rollback()
			break
		}
	}
	if err == nil {
		err = session.Commit()
	}
	session.Close()
	return err
}

func Classify(obj interface{}, operation string, table string, force bool) Action {
	var action Action
	if operation == "insert" {
		action = insert(obj, table)
	} else if operation == "update" {
		action = update(obj, table, force)
	} else if operation == "delete" {
		action = actionDelete(obj, table)
	}
	return action
}

func insert(obj interface{}, table string) Action {
	var action Action
	params := make([]interface{}, 0)
	column := `(`
	value := ` value (`
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	length := t.NumField()
	for i := 0; i < length; i++ {
		if t.Field(i).Name == "Id" {
			if v.Field(i).Int() == 0 {
				continue
			}
		}
		fetchType := false
		f := v.Field(i).Type().String()
		if f == "int" || f == "int64" {
			params = append(params, v.Field(i).Int())
			fetchType = true
		}
		if f == "time.Time" {
			params = append(params, v.Field(i).Interface().(time.Time).Format(models.DatetimeFormat))
			fetchType = true
		}
		if f == "string" {
			params = append(params, v.Field(i).String())
			fetchType = true
		}
		if !fetchType {
			continue
		}
		if i == length-1 {
			column = column + transColumn(t.Field(i).Name) + `)`
			value = value + `?)`
		} else {
			column = column + transColumn(t.Field(i).Name) + `,`
			value = value + `?,`
		}
	}
	action.Sql = `insert into ` + table + column + value
	action.Param = params
	return action
}

func update(obj interface{}, table string, force bool) Action {
	var action Action
	params := make([]interface{}, 0)
	var where, value string
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	var tmpId int64
	var tmpGuid string
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Name == "Id" {
			tmpId = v.Field(i).Int()
			where = ` where id=?`
			continue
		}
		if where == "" && t.Field(i).Name == "Guid" {
			tmpGuid = v.Field(i).String()
			where = ` where guid=?`
			continue
		}
		fetchType := false
		f := v.Field(i).Type().String()
		if strings.Contains(f, "int") {
			if v.Field(i).Int() > 0 || force {
				params = append(params, v.Field(i).Int())
				fetchType = true
			}
		}
		if f == "string" {
			if v.Field(i).String() != "" || force {
				params = append(params, v.Field(i).String())
				fetchType = true
			}
		}
		if f == "time.Time" {
			tt := v.Field(i).Interface().(time.Time)
			if tt.Unix() > 0 {
				params = append(params, tt.Format(models.DatetimeFormat))
				fetchType = true
			}
		}
		if !fetchType {
			continue
		}
		value = value + transColumn(t.Field(i).Name) + `=?,`
	}
	if tmpGuid != "" {
		params = append(params, tmpGuid)
	} else {
		if tmpId > 0 {
			params = append(params, tmpId)
		}
	}
	if len(params) > 0 {
		value = value[0 : len(value)-1]
		action.Sql = `update ` + table + ` set ` + value + where
		action.Param = params
	} else {
		action.Sql = ""
		action.Param = params
	}
	return action
}

func actionDelete(obj interface{}, table string) Action {
	var action Action
	params := make([]interface{}, 0)
	var where string
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Name == "Id" {
			where = ` where id=?`
			params = append(params, v.Field(i).Int())
			break
		}
		if t.Field(i).Name == "Guid" {
			where = ` where guid=?`
			params = append(params, v.Field(i).String())
			break
		}
		fetchType := false
		f := v.Field(i).Type().String()
		if strings.Contains(f, "int") {
			if v.Field(i).Int() > 0 {
				params = append(params, v.Field(i).Int())
				fetchType = true
			}
		}
		if f == "string" {
			if v.Field(i).String() != "" {
				params = append(params, v.Field(i).String())
				fetchType = true
			}
		}
		if !fetchType {
			continue
		}
		where = ` where ` + transColumn(t.Field(i).Name) + `=? and `
	}
	if strings.Contains(where, " and ") {
		where = where[:len(where)-4]
	}
	action.Sql = `delete from ` + table + where
	action.Param = params
	if len(params) == 0 {
		action.Sql = ""
	}
	return action
}

func transColumn(s string) string {
	r := []byte(s)
	var v []byte
	for i := 0; i < len(r); i++ {
		rr := r[i]
		if 'A' <= rr && rr <= 'Z' {
			rr += 'a' - 'A'
			if i != 0 {
				v = append(v, '_')
			}
		}
		v = append(v, rr)
	}
	return string(v)
}

type dbLogger struct {
	LogLevel xorm_log.LogLevel
	ShowSql  bool
	Logger   *zap.SugaredLogger
}

func (d *dbLogger) Debug(v ...interface{}) {
	d.Logger.Debugw(fmt.Sprint(v...))
}

func (d *dbLogger) Debugf(format string, v ...interface{}) {
	d.Logger.Debugw(fmt.Sprintf(format, v...))
}

func (d *dbLogger) Error(v ...interface{}) {
	d.Logger.Errorw(fmt.Sprint(v...))
}

func (d *dbLogger) Errorf(format string, v ...interface{}) {
	d.Logger.Errorw(fmt.Sprintf(format, v...))
}

func (d *dbLogger) Info(v ...interface{}) {
	d.Logger.Infow(fmt.Sprint(v...))
}

func (d *dbLogger) Infof(format string, v ...interface{}) {
	if len(v) < 4 {
		d.Logger.Infow(fmt.Sprintf(format, v...))
		return
	}
	var costMs float64 = 0
	costTime := fmt.Sprintf("%s", v[3])
	if strings.Contains(costTime, "µs") {
		costMs, _ = strconv.ParseFloat(strings.ReplaceAll(costTime, "µs", ""), 64)
		costMs = costMs / 1000
	} else if strings.Contains(costTime, "ms") {
		costMs, _ = strconv.ParseFloat(costTime[:len(costTime)-2], 64)
	} else if strings.Contains(costTime, "s") && !strings.Contains(costTime, "m") {
		costMs, _ = strconv.ParseFloat(costTime[:len(costTime)-1], 64)
		costMs = costMs * 1000
	} else {
		costTime = costTime[:len(costTime)-1]
		mIndex := strings.Index(costTime, "m")
		minTime, _ := strconv.ParseFloat(costTime[:mIndex], 64)
		secTime, _ := strconv.ParseFloat(costTime[mIndex+1:], 64)
		costMs = (minTime*60 + secTime) * 1000
	}
	d.Logger.Infow("db_log", zap.String("sql", fmt.Sprintf("%s", v[1])), zap.String("param", fmt.Sprintf("%v", v[2])), zap.Float64("cost_ms", costMs))
}

func (d *dbLogger) Warn(v ...interface{}) {
	d.Logger.Warnw(fmt.Sprint(v...))
}

func (d *dbLogger) Warnf(format string, v ...interface{}) {
	d.Logger.Warnw(fmt.Sprintf(format, v...))
}

func (d *dbLogger) Level() xorm_log.LogLevel {
	return d.LogLevel
}

func (d *dbLogger) SetLevel(l xorm_log.LogLevel) {
	d.LogLevel = l
}

func (d *dbLogger) ShowSQL(b ...bool) {
	d.ShowSql = b[0]
}

func (d *dbLogger) IsShowSQL() bool {
	return d.ShowSql
}

func transPageInfoToSQL(pageInfo models.PageInfo) (pageSql string, param []interface{}) {
	pageSql = " LIMIT ?,? "
	param = append(param, pageInfo.StartIndex*pageInfo.PageSize)
	param = append(param, pageInfo.PageSize)
	return
}

func queryCount(sql string, params ...interface{}) int {
	sql = "SELECT COUNT(1) FROM ( " + sql + " ) sub_query"
	params = append([]interface{}{sql}, params...)
	queryRows, err := x.QueryString(params...)
	if err != nil || len(queryRows) == 0 {
		log.Error(nil, log.LOGGER_APP, "Query sql count message fail", zap.Error(err))
		return 0
	}
	if _, b := queryRows[0]["COUNT(1)"]; b {
		countNum, _ := strconv.Atoi(queryRows[0]["COUNT(1)"])
		return countNum
	}
	return 0
}

func getJsonToXormMap(input interface{}) (resultMap map[string]string, idKeyName string) {
	resultMap = make(map[string]string)
	t := reflect.TypeOf(input)
	for i := 0; i < t.NumField(); i++ {
		resultMap[t.Field(i).Tag.Get("json")] = t.Field(i).Tag.Get("xorm")
		if i == 0 {
			idKeyName = t.Field(i).Tag.Get("xorm")
		}
	}
	return resultMap, idKeyName
}

func createListParams(inputList []string, prefix string) (specSql string, paramList []interface{}) {
	if len(inputList) > 0 {
		var specList []string
		for _, v := range inputList {
			specList = append(specList, "?")
			paramList = append(paramList, prefix+v)
		}
		specSql = strings.Join(specList, ",")
	}
	return
}

func transFiltersToSQL(queryParam *models.QueryRequestParam, transParam *models.TransFiltersParam) (filterSql, queryColumn string, param []interface{}) {
	if transParam.Prefix != "" && !strings.HasSuffix(transParam.Prefix, ".") {
		transParam.Prefix = transParam.Prefix + "."
	}
	if transParam.IsStruct {
		transParam.KeyMap, transParam.PrimaryKey = getJsonToXormMap(transParam.StructObj)
	}
	for _, filter := range queryParam.Filters {
		if transParam.KeyMap[filter.Name] == "" || transParam.KeyMap[filter.Name] == "-" {
			continue
		}
		if filter.Operator == "eq" {
			filterSql += fmt.Sprintf(" AND %s%s=? ", transParam.Prefix, transParam.KeyMap[filter.Name])
			param = append(param, filter.Value)
		} else if filter.Operator == "contains" || filter.Operator == "like" {
			filterSql += fmt.Sprintf(" AND %s%s LIKE ? ", transParam.Prefix, transParam.KeyMap[filter.Name])
			param = append(param, fmt.Sprintf("%%%s%%", filter.Value))
		} else if filter.Operator == "in" {
			inValueList := filter.Value.([]interface{})
			inValueStringList := []string{}
			for _, inValueInterfaceObj := range inValueList {
				if inValueInterfaceObj == nil {
					inValueStringList = append(inValueStringList, "")
				} else {
					inValueStringList = append(inValueStringList, inValueInterfaceObj.(string))
				}
			}
			tmpSpecSql, tmpListParams := createListParams(inValueStringList, "")
			filterSql += fmt.Sprintf(" AND %s%s in (%s) ", transParam.Prefix, transParam.KeyMap[filter.Name], tmpSpecSql)
			param = append(param, tmpListParams...)
		} else if filter.Operator == "lt" {
			filterSql += fmt.Sprintf(" AND %s%s<=? ", transParam.Prefix, transParam.KeyMap[filter.Name])
			param = append(param, filter.Value)
		} else if filter.Operator == "gt" {
			filterSql += fmt.Sprintf(" AND %s%s>=? ", transParam.Prefix, transParam.KeyMap[filter.Name])
			param = append(param, filter.Value)
		} else if filter.Operator == "ne" || filter.Operator == "neq" {
			filterSql += fmt.Sprintf(" AND %s%s!=? ", transParam.Prefix, transParam.KeyMap[filter.Name])
			param = append(param, filter.Value)
		} else if filter.Operator == "notNull" || filter.Operator == "isnot" {
			filterSql += fmt.Sprintf(" AND %s%s is not null ", transParam.Prefix, transParam.KeyMap[filter.Name])
		} else if filter.Operator == "null" || filter.Operator == "is" {
			filterSql += fmt.Sprintf(" AND %s%s is null ", transParam.Prefix, transParam.KeyMap[filter.Name])
		}
	}
	if queryParam.Sorting != nil {
		if transParam.KeyMap[queryParam.Sorting.Field] == "" || transParam.KeyMap[queryParam.Sorting.Field] == "-" {
			queryParam.Sorting.Field = transParam.PrimaryKey
		} else {
			queryParam.Sorting.Field = transParam.KeyMap[queryParam.Sorting.Field]
		}
		if queryParam.Sorting.Asc {
			filterSql += fmt.Sprintf(" ORDER BY %s%s ASC ", transParam.Prefix, queryParam.Sorting.Field)
		} else {
			filterSql += fmt.Sprintf(" ORDER BY %s%s DESC ", transParam.Prefix, queryParam.Sorting.Field)
		}
	}
	if len(queryParam.ResultColumns) > 0 {
		for _, resultColumn := range queryParam.ResultColumns {
			if transParam.KeyMap[resultColumn] == "" || transParam.KeyMap[resultColumn] == "-" {
				continue
			}
			queryColumn += fmt.Sprintf("%s%s,", transParam.Prefix, transParam.KeyMap[resultColumn])
		}
	}
	if queryColumn == "" {
		queryColumn = " * "
	} else {
		queryColumn = queryColumn[:len(queryColumn)-1]
	}
	return
}
