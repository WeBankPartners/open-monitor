package funcs

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	"log"
	"sync"
	"time"
)

type DbMonitorTaskObj struct {
	DbType   string       `json:"db_type"`
	Endpoint string       `json:"endpoint"`
	Name     string       `json:"name"`
	Server   string       `json:"server"`
	Port     string       `json:"port"`
	User     string       `json:"user"`
	Password string       `json:"password"`
	Sql      string       `json:"sql"`
	Session  *xorm.Engine `json:"session"`
}

type DbMonitorResultObj struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
	Server   string `json:"server"`
	Port     string `json:"port"`
	Value    int    `json:"value"`
}

var (
	taskList     []*DbMonitorTaskObj
	taskLock     = new(sync.RWMutex)
	resultList   []*DbMonitorResultObj
	resultLock   = new(sync.RWMutex)
	taskInterval = 10
	maxIdle      = 2
	maxOpen      = 5
	timeOut      = 10
	metricString = "db_monitor_value"
)

func StartCronTask() {
	t := time.NewTicker(time.Duration(taskInterval) * time.Second).C
	for {
		<-t
		go doTask()
	}
}

func doTask() {
	taskLock.RLock()
	defer taskLock.RUnlock()
	if len(taskList) == 0 {
		return
	}
	var newResultList []*DbMonitorResultObj
	for _, taskObj := range taskList {
		var resultValue int
		if taskObj.DbType == "mysql" {
			resultValue = mysqlTask(taskObj)
		}
		newResultList = append(newResultList, &DbMonitorResultObj{Name: taskObj.Name, Endpoint: taskObj.Endpoint, Server: taskObj.Server, Port: taskObj.Port, Value: resultValue})
	}
	resultLock.Lock()
	resultList = newResultList
	resultLock.Unlock()
}

func mysqlTask(config *DbMonitorTaskObj) int {
	if config.Session == nil {
		connectStr := fmt.Sprintf("%s:%s@%s(%s:%s)/?collation=utf8mb4_unicode_ci&allowNativePasswords=true",
			config.User, config.Password, "tcp", config.Server, config.Port)
		tmpSession, err := xorm.NewEngine("mysql", connectStr)
		if err != nil {
			log.Printf("mysql connect fail with connectStr %s,error: %s\n", connectStr, err.Error())
			return -1
		} else {
			tmpSession.SetMaxIdleConns(maxIdle)
			tmpSession.SetMaxOpenConns(maxOpen)
			tmpSession.SetConnMaxLifetime(time.Duration(timeOut) * time.Second)
			tmpSession.Charset("utf8")
			// 使用驼峰式映射
			tmpSession.SetMapper(core.SnakeMapper{})
			config.Session = tmpSession
		}
	}
	queryMap := make(map[string]int)
	_, err := config.Session.SQL(config.Sql).Get(&queryMap)
	if err != nil {
		log.Printf("mysql query data fail with sql:%s,error: %s\n", config.Sql, err.Error())
		return -2
	}
	var resultValue int
	for _, v := range queryMap {
		resultValue = v
	}
	return resultValue
}

func checkIllegal(param DbMonitorTaskObj) error {
	connectStr := fmt.Sprintf("%s:%s@%s(%s:%s)/?collation=utf8mb4_unicode_ci&allowNativePasswords=true",
		param.User, param.Password, "tcp", param.Server, param.Port)
	tmpSession, err := xorm.NewEngine("mysql", connectStr)
	if err != nil {
		log.Printf("check illegal, mysql connect fail with connectStr %s,error: %s\n", connectStr, err.Error())
		return fmt.Errorf("Mysql connect fail,%s ", err.Error())
	} else {
		tmpSession.SetMaxIdleConns(maxIdle)
		tmpSession.SetMaxOpenConns(maxOpen)
		tmpSession.SetConnMaxLifetime(time.Duration(timeOut) * time.Second)
		tmpSession.Charset("utf8")
		// 使用驼峰式映射
		tmpSession.SetMapper(core.SnakeMapper{})
		queryMap := make(map[string]int)
		_, err := tmpSession.SQL(param.Sql).Get(&queryMap)
		if err != nil {
			log.Printf("check illegal, mysql query data fail with sql:%s,error: %s\n", param.Sql, err.Error())
			return fmt.Errorf("Mysql query data fail,%s ", err.Error())
		}
		if len(queryMap) != 1 {
			err = fmt.Errorf("Query result row equal %d,must be one ", len(queryMap))
		}
		return err
	}
}
