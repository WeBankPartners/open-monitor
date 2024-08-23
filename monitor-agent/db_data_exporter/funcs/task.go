package funcs

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	"log"
	"strconv"
	"sync"
	"time"
)

type DbMonitorTaskObj struct {
	DbType         string       `json:"db_type"`
	Endpoint       string       `json:"endpoint"`
	Name           string       `json:"name"`
	Server         string       `json:"server"`
	Port           string       `json:"port"`
	User           string       `json:"user"`
	Password       string       `json:"password"`
	Sql            string       `json:"sql"`
	Step           int64        `json:"step"`
	LastTime       int64        `json:"last_time"`
	ServiceGroup   string       `json:"service_group"`
	Session        *xorm.Engine `json:"session"`
	KeywordGuid    string       `json:"keyword_guid"`
	KeywordCount   int64        `json:"keyword_count"`
	KeywordContent string       `json:"keyword_content"`
}

type DbMonitorResultObj struct {
	Name         string  `json:"name"`
	Endpoint     string  `json:"endpoint"`
	Server       string  `json:"server"`
	Port         string  `json:"port"`
	Value        float64 `json:"value"`
	ServiceGroup string  `json:"service_group"`
	KeywordGuid  string  `json:"keyword_guid"`
	KeywordCount int64   `json:"keyword_count"`
	Step         int64   `json:"step"`
}

type DbLastKeywordDto struct {
	KeywordGuid    string `json:"keyword_guid"`
	KeywordContent string `json:"keyword_content"`
	Endpoint       string `json:"endpoint"`
}

var (
	taskList        []*DbMonitorTaskObj
	taskLock        = new(sync.RWMutex)
	resultList      []*DbMonitorResultObj
	resultLock      = new(sync.RWMutex)
	taskInterval    = 10
	maxIdle         = 2
	maxOpen         = 5
	timeOut         = 10
	metricString    = "db_monitor_value"
	dbKeywordMetric = "db_keyword_value"
)

func StartCronTask() {
	log.Println("start cron task")
	minuteTime, _ := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprintf("%s:00", time.Now().Format("2006-01-02 15:04")), time.Local)
	time.Sleep(time.Duration((minuteTime.Unix()+60)-time.Now().Unix()) * time.Second)
	t := time.NewTicker(time.Duration(taskInterval) * time.Second).C
	for {
		<-t
		go doTask()
	}
}

func doTask() {
	taskLock.RLock()
	if len(taskList) == 0 {
		taskLock.RUnlock()
		return
	}
	var newResultList []*DbMonitorResultObj
	nowTime := time.Now().Unix()
	for _, taskObj := range taskList {
		if !checkStepActive(taskObj.LastTime, nowTime, taskObj.Step) {
			continue
		} else if taskObj.Step > 10 && taskObj.LastTime == 0 {
			log.Printf("step:%d task:%s start doTask \n", taskObj.Step, taskObj.Name)
		}
		var resultValue float64
		if taskObj.DbType == "mysql" {
			resultValue = mysqlTask(taskObj)
		}
		newResultList = append(newResultList, &DbMonitorResultObj{Name: taskObj.Name, Endpoint: taskObj.Endpoint, Server: taskObj.Server, Port: taskObj.Port, Value: resultValue, ServiceGroup: taskObj.ServiceGroup, KeywordGuid: taskObj.KeywordGuid, KeywordCount: taskObj.KeywordCount, Step: taskObj.Step})
		taskObj.LastTime = nowTime
	}
	taskLock.RUnlock()
	resultLock.Lock()
	resultList = newResultList
	resultLock.Unlock()
}

func checkStepActive(lastTime, nowTime, step int64) bool {
	if step <= 10 {
		return true
	}
	if lastTime == 0 && step > 60 {
		if checkStepNearbyTime(nowTime, step) {
			return true
		} else {
			return false
		}
	}
	if (nowTime - lastTime) >= step {
		return true
	}
	return false
}

func mysqlTask(config *DbMonitorTaskObj) float64 {
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
	queryStringMap, err := config.Session.QueryString(config.Sql)
	if err != nil {
		log.Printf("mysql query data fail with sql:%s,error: %s\n", config.Sql, err.Error())
		return -2
	}
	var resultValue float64
	if config.KeywordGuid != "" {
		if len(queryStringMap) > 0 {
			config.KeywordCount = config.KeywordCount + 1
			rowOneBytes, _ := json.Marshal(queryStringMap[0])
			config.KeywordContent = string(rowOneBytes)
		}
	}
	if len(queryStringMap) > 0 {
		for _, v := range queryStringMap[0] {
			resultValue, _ = strconv.ParseFloat(v, 64)
		}
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

		queryStringMap, err := tmpSession.QueryString(param.Sql)
		if err != nil {
			log.Printf("check illegal, mysql query data fail with sql:%s,error: %s\n", param.Sql, err.Error())
			return fmt.Errorf("Mysql query data fail,%s ", err.Error())
		}
		if len(queryStringMap) != 1 {
			return fmt.Errorf("Query result row num %d ", len(queryStringMap))
		}
		if len(queryStringMap[0]) != 1 {
			return fmt.Errorf("Query result return column num %d ", len(queryStringMap[0]))
		}
		for _, v := range queryStringMap[0] {
			_, err = strconv.ParseFloat(v, 64)
			if err != nil {
				err = fmt.Errorf("Query result:%s format float type fail,%s ", v, err.Error())
			}
		}
		return err
	}
}

func checkStepNearbyTime(nowTime, step int64) bool {
	dayTime, _ := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprintf("%s 00:00:00", time.Now().Format("2006-01-02")), time.Local)
	lastDayUnix := dayTime.Unix()
	for lastDayUnix < nowTime {
		lastDayUnix = lastDayUnix + step
	}
	if (lastDayUnix - nowTime) <= 10 {
		return true
	}
	return false
}
