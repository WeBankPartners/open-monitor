package db

import (
	_ "github.com/go-sql-driver/mysql"
	"xorm.io/core"
	"xorm.io/xorm"
	"fmt"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"time"
	"reflect"
	"strings"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
)

var (
	x *xorm.Engine
    archiveMysql *xorm.Engine
    archiveDatabase  string
    ArchiveEnable bool
)
//var RedisStore sessions.RedisStore

type DBObj struct {
	x *xorm.Engine
	DbType  string
	ConnUser  string
	ConnPwd   string
	ConnHost  string
	ConnDb    string
	ConnPtl   string
	MaxIdle   int
	MaxOpen   int
	Timeout   int
}

func (d *DBObj) InitXorm()  {
	log.Logger.Info("Start init db")
	cnnstr := fmt.Sprintf("%s:%s@%s(%s)/%s?collation=utf8mb4_unicode_ci&allowNativePasswords=true",
		d.ConnUser, d.ConnPwd, d.ConnPtl, d.ConnHost, d.ConnDb)
	engine,err := xorm.NewEngine(d.DbType, cnnstr)
	if err!=nil {
		log.Logger.Error("Init db fail", log.Error(err))
	}
	engine.SetMaxIdleConns(d.MaxIdle)
	engine.SetMaxOpenConns(d.MaxOpen)
	// 使用驼峰式映射
	engine.SetMapper(core.SnakeMapper{})
	d.x = engine
	go keepAlive(d.Timeout)
}

func InitDbConn() {
	dbCfg := m.Config().Store
	if dbCfg.Type == "mysql" {
		initDefaultMysql(dbCfg)
	}
	tmpEnable := strings.ToLower(m.Config().ArchiveMysql.Enable)
	if tmpEnable == "y" || tmpEnable == "yes" || tmpEnable == "true" {
		initArchiveDbEngine()
	}else{
		ArchiveEnable = false
	}
}

func initDefaultMysql(dbCfg m.StoreConfig)  {
	dbObj := DBObj{DbType: dbCfg.Type, ConnUser: dbCfg.User, ConnPwd: dbCfg.Pwd, ConnHost: fmt.Sprintf("%s:%d", dbCfg.Server, dbCfg.Port), ConnDb: dbCfg.DataBase, ConnPtl: "tcp", MaxOpen: dbCfg.MaxOpen, MaxIdle: dbCfg.MaxIdle, Timeout: dbCfg.Timeout}
	dbObj.InitXorm()
	x = dbObj.x
	log.Logger.Info("Default db init success")
}

func initArchiveDbEngine() {
	databaseName := m.Config().ArchiveMysql.DatabasePrefix + time.Now().Format("2006")
	var err error
	connectStr := fmt.Sprintf("%s:%s@%s(%s:%s)/%s?collation=utf8mb4_unicode_ci&allowNativePasswords=true",
		m.Config().ArchiveMysql.User, m.Config().ArchiveMysql.Password, "tcp", m.Config().ArchiveMysql.Server, m.Config().ArchiveMysql.Port, databaseName)
	archiveMysql,err = xorm.NewEngine("mysql", connectStr)
	if err != nil {
		ArchiveEnable = false
		log.Logger.Error("Init archive mysql fail", log.String("connectStr",connectStr), log.Error(err))
	}else{
		ArchiveEnable = true
		archiveMysql.SetMaxIdleConns(m.Config().ArchiveMysql.MaxIdle)
		archiveMysql.SetMaxOpenConns(m.Config().ArchiveMysql.MaxOpen)
		archiveMysql.SetConnMaxLifetime(time.Duration(m.Config().ArchiveMysql.Timeout)*time.Second)
		archiveMysql.Charset("utf8")
		// 使用驼峰式映射
		archiveMysql.SetMapper(core.SnakeMapper{})
		archiveDatabase = databaseName
		log.Logger.Info("Init archive mysql "+archiveDatabase+" success")
	}
}

func checkArchiveDatabase()  {
	databaseName := m.Config().ArchiveMysql.DatabasePrefix + time.Now().Format("2006")
	if databaseName == archiveDatabase {
		return
	}
	initArchiveDbEngine()
}

type Action struct {
	Sql  string
	Param  []interface{}
}

func Transaction(actions []*Action) error {
	if len(actions) == 0 {
		return fmt.Errorf("transaction actions is null")
	}
	session := x.NewSession()
	err := session.Begin()
	for _,action := range actions {
		params := make([]interface{}, 0)
		params = append(params, action.Sql)
		for _,v := range action.Param {
			params = append(params, v)
		}
		_,err = session.Exec(params...)
		if err != nil {
			session.Rollback()
			break
		}
	}
	if err==nil {
		err = session.Commit()
	}
	session.Close()
	return err
}

type dbTransactionFunc func(session *xorm.Session) error

func InTransaction(callback dbTransactionFunc) error {
	return inTransactionWithRetry(callback)
}

func inTransactionWithRetry(callback dbTransactionFunc) error {
	var err error
	session := x.NewSession()
	defer session.Close()
	if err = session.Begin(); err != nil {
		return err
	}
	err = callback(session)
	if err != nil {
		session.Rollback()
		return err
	} else if err = session.Commit(); err != nil {
		return err
	}
	return nil
}

func keepAlive(interval int)  {
	for {
		x.Exec(`select 1`)
		time.Sleep(time.Duration(interval)*time.Second)
	}
}

func Classify(obj interface{}, operation string, table string, force bool) Action {
	var action Action
	if operation == "insert" {
		action = insert(obj, table)
	}else if operation == "update" {
		action = update(obj, table, force)
	}else if operation == "delete" {
		action = delete(obj, table)
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
			params = append(params, v.Field(i).Interface().(time.Time).Format(m.DatetimeFormat))
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
		}else{
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
	var where,value string
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
				params = append(params, tt.Format(m.DatetimeFormat))
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
	}else{
		if tmpId > 0 {
			params = append(params, tmpId)
		}
	}
	if len(params) > 0 {
		value = value[0:len(value)-1]
		action.Sql = `update ` + table + ` set ` + value + where
		action.Param = params
	}else{
		action.Sql = ""
		action.Param = params
	}
	return action
}

func delete(obj interface{}, table string) Action {
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