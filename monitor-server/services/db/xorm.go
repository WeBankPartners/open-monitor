package db

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"fmt"
	"github.com/go-xorm/core"
	mid "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/middleware"
	m "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/models"
	"time"
)

var stores []*xorm.Engine
var x *xorm.Engine
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
	mid.LogInfo("start init db")
	cnnstr := fmt.Sprintf("%s:%s@%s(%s)/%s?collation=utf8mb4_unicode_ci&allowNativePasswords=true",
		d.ConnUser, d.ConnPwd, d.ConnPtl, d.ConnHost, d.ConnDb)
	engine,err := xorm.NewEngine(d.DbType, cnnstr)
	if err!=nil {
		mid.LogError("init db fail", err)
	}
	engine.SetMaxIdleConns(d.MaxIdle)
	engine.SetMaxOpenConns(d.MaxOpen)
	// 使用驼峰式映射
	engine.SetMapper(core.SnakeMapper{})
	d.x = engine
	go keepAlive(d.Timeout)
}

func InitDbConn() {
	dbCfgList := m.Config().Store
	if len(dbCfgList) <= 0{
		mid.LogInfo("db config list is none")
		return
	}
	for _,v := range dbCfgList {
		if v.Name == "default" && v.Type == "mysql" {
			initDefaultMysql(*v)
		}
		//if v.Type == "redis" {
		//	initDefaultRedis(*v)
		//}
		//if v.Type == "mongodb" {
		//	InitMongodb(*v)
		//}
	}
}

func initDefaultMysql(dbCfg m.StoreConfig)  {
	dbObj := DBObj{DbType: dbCfg.Type, ConnUser: dbCfg.User, ConnPwd: dbCfg.Pwd, ConnHost: fmt.Sprintf("%s:%d", dbCfg.Server, dbCfg.Port), ConnDb: dbCfg.DataBase, ConnPtl: "tcp", MaxOpen: dbCfg.MaxOpen, MaxIdle: dbCfg.MaxIdle, Timeout: dbCfg.Timeout}
	dbObj.InitXorm()
	stores = []*xorm.Engine{dbObj.x}
	x = dbObj.x
	mid.LogInfo("default db init success")
}

func Default() *xorm.Engine {
	return stores[0]
}

type Action struct {
	Sql  string
	Param  []interface{}
}

func ExecuteTransactionSql(sqls []string) error {
	var actions []*Action
	for _,sql := range sqls {
		action := Action{Sql:sql}
		actions = append(actions, &action)
	}
	err := Transaction(actions)
	return err
}

func Transaction(actions []*Action) error {
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
