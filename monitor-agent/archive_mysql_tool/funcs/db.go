package funcs

import (
	"github.com/go-xorm/xorm"
	"github.com/go-xorm/core"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"time"
	"fmt"
)

var (
	mysqlEngine *xorm.Engine
	monitorMysqlEngine *xorm.Engine
)

func InitDbEngine() (err error) {
	connectStr := fmt.Sprintf("%s:%s@%s(%s:%d)/%s?collation=utf8mb4_unicode_ci&allowNativePasswords=true",
		Config().Mysql.User, Config().Mysql.Password, "tcp", Config().Mysql.Server, Config().Mysql.Port, Config().Mysql.DataBase)
	mysqlEngine,err = xorm.NewEngine("mysql", connectStr)
	if err != nil {
		log.Printf("init mysql fail with connect: %s error: %v \n", connectStr, err)
	}else{
		mysqlEngine.SetMaxIdleConns(Config().Mysql.MaxIdle)
		mysqlEngine.SetMaxOpenConns(Config().Mysql.MaxOpen)
		mysqlEngine.SetConnMaxLifetime(time.Duration(Config().Mysql.Timeout)*time.Second)
		mysqlEngine.Charset("utf8")
		// 使用驼峰式映射
		mysqlEngine.SetMapper(core.SnakeMapper{})
		log.Println("init mysql success ")
	}
	return err
}

func InitMonitorDbEngine() (err error) {
	connectStr := fmt.Sprintf("%s:%s@%s(%s:%d)/%s?collation=utf8mb4_unicode_ci&allowNativePasswords=true",
		Config().Monitor.Mysql.User, Config().Monitor.Mysql.Password, "tcp", Config().Monitor.Mysql.Server, Config().Monitor.Mysql.Port, Config().Monitor.Mysql.DataBase)
	monitorMysqlEngine,err = xorm.NewEngine("mysql", connectStr)
	if err != nil {
		log.Printf("init monitor mysql fail with connect: %s error: %v \n", connectStr, err)
	}else{
		monitorMysqlEngine.SetMaxIdleConns(Config().Monitor.Mysql.MaxIdle)
		monitorMysqlEngine.SetMaxOpenConns(Config().Monitor.Mysql.MaxOpen)
		monitorMysqlEngine.SetConnMaxLifetime(time.Duration(Config().Monitor.Mysql.Timeout)*time.Second)
		monitorMysqlEngine.Charset("utf8")
		// 使用驼峰式映射
		monitorMysqlEngine.SetMapper(core.SnakeMapper{})
		log.Println("init monitor mysql success ")
	}
	return err
}

func insertMysql(rows []*ArchiveTable,tableName string) error {
	var sqlList []string
	sqlString := fmt.Sprintf("INSERT INTO %s VALUES ", tableName)
	for i,v := range rows {
		sqlString += fmt.Sprintf("('%s','%s','%s',%d,%.3f,%.3f,%.3f,%.3f)", v.Endpoint, v.Metric, v.Tags, v.UnixTime, v.Avg, v.Min, v.Max, v.P95)
		if (i+1)%100 == 0 {
			sqlList = append(sqlList, sqlString)
			sqlString = fmt.Sprintf("INSERT INTO %s VALUES ", tableName)
		}else{
			sqlString += ","
		}
	}
	for _,v := range sqlList {
		_,err := mysqlEngine.Exec(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func createTable(start int64) (err error, tableName string) {
	tableName = fmt.Sprintf("archive_%s", time.Unix(start, 0).Format("2006_01_02"))
	if checkTableExists(tableName) {
		return nil,tableName
	}
	tableDate := time.Unix(start, 0).Format("2006_01_02")
	createSql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (`endpoint` VARCHAR(255) NOT NULL,`metric` VARCHAR(255) NOT NULL,`tags` VARCHAR(500) NOT NULL DEFAULT '',`unix_time` INT(11) NOT NULL,`avg` DOUBLE NOT NULL DEFAULT 0,`min` DOUBLE NOT NULL DEFAULT 0,`max` DOUBLE NOT NULL DEFAULT 0,`p95` DOUBLE NOT NULL DEFAULT 0,INDEX idx_%s_endpoint (`endpoint`),INDEX idx_%s_metric (`metric`)) ENGINE=INNODB DEFAULT CHARSET=utf8",tableName,tableDate,tableDate)
	_,err = mysqlEngine.Exec(createSql)
	if err != nil {
		log.Printf("create table %s error: %v \n", tableName, err)
	}
	return err,tableName
}

func checkTableExists(tableName string) bool {
	var tables []*PrometheusArchiveTables
	err := mysqlEngine.SQL("show tables").Find(&tables)
	if err != nil {
		log.Printf("show tables error: %v \n", err)
		return false
	}
	for _,v := range tables {
		if v.TablesInPrometheusArchive == tableName {
			return true
		}
	}
	return false
}