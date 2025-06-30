package funcs

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
)

var (
	mysqlEngine         *xorm.Engine
	monitorMysqlEngine  *xorm.Engine
	databaseSelect      string
	hostIp              string
	maxUnitNum          int
	concurrentInsertNum int
	retryWaitSecond     int
	jobTimeout          int
)

func InitDbEngine(databaseName string) (err error) {
	if databaseName == "" {
		databaseName = "mysql"
	}
	connectStr := fmt.Sprintf("%s:%s@%s(%s:%s)/%s?collation=utf8mb4_unicode_ci&allowNativePasswords=true",
		Config().Mysql.User, Config().Mysql.Password, "tcp", Config().Mysql.Server, Config().Mysql.Port, databaseName)
	mysqlEngine, err = xorm.NewEngine("mysql", connectStr)
	if err != nil {
		log.Printf("init mysql fail with connect: %s error: %v \n", connectStr, err)
	} else {
		mysqlEngine.SetMaxIdleConns(Config().Mysql.MaxIdle)
		mysqlEngine.SetMaxOpenConns(Config().Mysql.MaxOpen)
		mysqlEngine.SetConnMaxLifetime(time.Duration(Config().Mysql.Timeout) * time.Second)
		mysqlEngine.Charset("utf8")
		// 使用驼峰式映射
		mysqlEngine.SetMapper(core.SnakeMapper{})
		if !strings.HasPrefix(databaseName, Config().Mysql.DatabasePrefix) {
			err = ChangeDatabase("")
		} else {
			databaseSelect = databaseName
			log.Printf("init mysql %s success \n", databaseSelect)
			err = initJobRecordTable()
		}
	}
	return err
}

func ResetDbEngine() {
	err := mysqlEngine.Close()
	if err != nil {
		log.Printf("close mysql engine fail,%s \n", err.Error())
	}
	time.Sleep(30 * time.Second)
	databaseName := Config().Mysql.DatabasePrefix + time.Now().Format("2006")
	connectStr := fmt.Sprintf("%s:%s@%s(%s:%s)/%s?collation=utf8mb4_unicode_ci&allowNativePasswords=true",
		Config().Mysql.User, Config().Mysql.Password, "tcp", Config().Mysql.Server, Config().Mysql.Port, databaseName)
	mysqlEngine, err = xorm.NewEngine("mysql", connectStr)
	if err != nil {
		log.Printf("init mysql fail with connect: %s error: %v \n", connectStr, err)
	} else {
		mysqlEngine.SetMaxIdleConns(Config().Mysql.MaxIdle)
		mysqlEngine.SetMaxOpenConns(Config().Mysql.MaxOpen)
		mysqlEngine.SetConnMaxLifetime(time.Duration(Config().Mysql.Timeout) * time.Second)
		mysqlEngine.Charset("utf8")
		// 使用驼峰式映射
		mysqlEngine.SetMapper(core.SnakeMapper{})
	}
	log.Println("Reset db engine done! ")
}

func InitMonitorDbEngine() (err error) {
	connectStr := fmt.Sprintf("%s:%s@%s(%s:%s)/%s?collation=utf8mb4_unicode_ci&allowNativePasswords=true",
		Config().Monitor.Mysql.User, Config().Monitor.Mysql.Password, "tcp", Config().Monitor.Mysql.Server, Config().Monitor.Mysql.Port, Config().Monitor.Mysql.DataBase)
	monitorMysqlEngine, err = xorm.NewEngine("mysql", connectStr)
	if err != nil {
		log.Printf("init monitor mysql fail with connect: %s error: %v \n", connectStr, err)
	} else {
		monitorMysqlEngine.SetMaxIdleConns(Config().Monitor.Mysql.MaxIdle)
		monitorMysqlEngine.SetMaxOpenConns(Config().Monitor.Mysql.MaxOpen)
		monitorMysqlEngine.SetConnMaxLifetime(time.Duration(Config().Monitor.Mysql.Timeout) * time.Second)
		monitorMysqlEngine.Charset("utf8")
		// 使用驼峰式映射
		monitorMysqlEngine.SetMapper(core.SnakeMapper{})
		log.Println("init monitor mysql success ")
	}
	return err
}

func insertMysql(rows []*ArchiveTable, tableName string) error {
	startTime := time.Now()
	log.Printf("start insert mysql table:%s,row num:%d,concurrentInsertNum:%d \n", tableName, len(rows), concurrentInsertNum)
	var sqlList []string
	var rowCountList []int
	tmpCount := 0
	sqlString := fmt.Sprintf("INSERT INTO %s(endpoint,metric,tags,unix_time,`avg`,`min`,`max`,`p95`,`sum`,`create_time`) VALUES ", tableName)
	for i, v := range rows {
		tmpCount += 1
		sqlString += fmt.Sprintf("('%s','%s','%s',%d,%.3f,%.3f,%.3f,%.3f,%.3f,'%s')", strings.ReplaceAll(v.Endpoint, "'", ""), strings.ReplaceAll(v.Metric, "'", ""), strings.ReplaceAll(v.Tags, "'", ""), v.UnixTime, v.Avg, v.Min, v.Max, v.P95, v.Sum, transUnixTime(v.UnixTime))
		if (i+1)%concurrentInsertNum == 0 || i == len(rows)-1 {
			rowCountList = append(rowCountList, tmpCount)
			tmpCount = 0
			sqlList = append(sqlList, sqlString)
			sqlString = fmt.Sprintf("INSERT INTO %s(endpoint,metric,tags,unix_time,`avg`,`min`,`max`,`p95`,`sum`,`create_time`) VALUES ", tableName)
		} else {
			sqlString += ","
		}
	}
	gErrMessage := ""
	for sqlIndex, v := range sqlList {
		// var tmpErr error
		_, tmpErr := mysqlEngine.Exec(v)
		// for i := 0; i < 3; i++ {
		// 	//log.Printf("start try %d to insert mysql,data num:%d \n", i+1, rowCountList[sqlIndex])
		// 	_, err := mysqlEngine.Exec(v)
		// 	if err != nil {
		// 		tmpErr = err
		// 	} else {
		// 		tmpErr = nil
		// 		break
		// 	}
		// 	if i < 2 {
		// 		time.Sleep(time.Duration(retryWaitSecond) * time.Second)
		// 	}
		// }
		if tmpErr != nil {
			log.Printf("Exec sql error:%s sql:%s \n", tmpErr.Error(), v)
			tmpErrorString := tmpErr.Error()
			if len(tmpErrorString) > 200 {
				tmpErrorString = tmpErrorString[:200]
			}
			gErrMessage += fmt.Sprintf("fail with rows length:%d error:%s \n", rowCountList[sqlIndex], tmpErrorString)
		}
	}
	if gErrMessage == "" {
		log.Printf("done insert mysql table:%s,row num:%d,use time:%.3f s \n", tableName, len(rows), time.Now().Sub(startTime).Seconds())
		return nil
	} else {
		return fmt.Errorf(gErrMessage)
	}
}

func createTable(start int64, isFiveArchive bool) (err error, tableName string) {
	tableName = fmt.Sprintf("archive_%s", time.Unix(start, 0).Format("2006_01_02"))
	if isFiveArchive {
		tableName = tableName + "_5min"
	}
	err = ChangeDatabase(time.Unix(start, 0).Format("2006"))
	if err != nil {
		return err, tableName
	}
	if checkTableExists(tableName) {
		return nil, tableName
	}
	tableDate := time.Unix(start, 0).Format("2006_01_02")
	if isFiveArchive {
		tableDate = tableDate + "_5m"
	}
	createSql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (`id` int(11) unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,`endpoint` VARCHAR(255) NOT NULL,`metric` VARCHAR(255) NOT NULL,`tags` VARCHAR(1024) NOT NULL DEFAULT '',`unix_time` INT(11) NOT NULL,`avg` DOUBLE NOT NULL DEFAULT 0,`min` DOUBLE NOT NULL DEFAULT 0,`max` DOUBLE NOT NULL DEFAULT 0,`p95` DOUBLE NOT NULL DEFAULT 0,`sum` DOUBLE NOT NULL DEFAULT 0,`create_time` VARCHAR(64) DEFAULT NULL,INDEX idx_%s_endpoint (`endpoint`),INDEX idx_%s_metric (`metric`)) ENGINE=INNODB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8", tableName, tableDate, tableDate)
	_, err = mysqlEngine.Exec(createSql)
	if err != nil {
		log.Printf("create table %s error: %v \n", tableName, err)
	}
	return err, tableName
}

func checkTableExists(tableName string) bool {
	var tables []*PrometheusArchiveTables
	err := mysqlEngine.SQL(fmt.Sprintf("SELECT `TABLE_NAME` FROM information_schema.`TABLES` WHERE TABLE_SCHEMA='%s'", databaseSelect)).Find(&tables)
	if err != nil {
		log.Printf("show tables error: %v \n", err)
		return false
	}
	for _, v := range tables {
		if v.TableName == tableName {
			return true
		}
	}
	return false
}

func ChangeDatabase(year string) error {
	if year == "" {
		year = time.Now().Format("2006")
	}
	databaseName := Config().Mysql.DatabasePrefix + year
	if databaseName == databaseSelect {
		return nil
	}
	_, err := mysqlEngine.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", databaseName))
	if err != nil {
		log.Printf("create database error -> %v \n", err)
		return err
	}
	err = InitDbEngine(databaseName)
	return err
}

func getArchiveTableCountData(tableName string) (err error, result []*ArchiveCountQueryObj) {
	err = mysqlEngine.SQL(fmt.Sprintf("SELECT endpoint,metric,COUNT(1) AS num FROM %s GROUP BY endpoint,metric", tableName)).Find(&result)
	return err, result
}

func archiveOneToFive(oldTable, newTable, endpoint, metric string) error {
	var oldTableData []*ArchiveTable
	var newTableData []*ArchiveTable
	err := mysqlEngine.SQL(fmt.Sprintf("SELECT tags,unix_time,`avg`,`min`,`max`,`p95`,`sum` FROM %s WHERE endpoint='%s' AND metric='%s' ORDER BY tags,unix_time", oldTable, endpoint, metric)).Find(&oldTableData)
	if err != nil {
		return err
	}
	if len(oldTableData) == 0 {
		return fmt.Errorf("table:%s endpoint:%s metric:%s empty data", oldTable, endpoint, metric)
	}
	var tmpRowObj ArchiveFiveRowObj
	tmpCountIndex := 0
	for i, v := range oldTableData {
		if v.Tags != tmpRowObj.Tags {
			if tmpRowObj.UnixTime > 0 {
				newArchiveRowData := tmpRowObj.CalcArchiveTable()
				newTableData = append(newTableData, &newArchiveRowData)
			}
			tmpCountIndex = 0
		}
		if tmpCountIndex == 0 {
			tmpRowObj = ArchiveFiveRowObj{Endpoint: endpoint, Metric: metric, Tags: v.Tags, UnixTime: v.UnixTime, Avg: []float64{}, Min: []float64{}, Max: []float64{}, P95: []float64{}, Sum: []float64{}}
		}
		tmpCountIndex += 1
		tmpRowObj.Avg = append(tmpRowObj.Avg, v.Avg)
		tmpRowObj.Min = append(tmpRowObj.Min, v.Min)
		tmpRowObj.Max = append(tmpRowObj.Max, v.Max)
		tmpRowObj.P95 = append(tmpRowObj.P95, v.P95)
		tmpRowObj.Sum = append(tmpRowObj.Sum, v.Sum)
		if tmpCountIndex == 5 || i == len(oldTableData)-1 {
			newArchiveRowData := tmpRowObj.CalcArchiveTable()
			newTableData = append(newTableData, &newArchiveRowData)
			tmpCountIndex = 0
		}
	}
	err = insertMysql(newTableData, newTable)
	return err
}

func renameFiveToOne(oldTable, newTable string) error {
	var err error
	_, err = mysqlEngine.Exec(fmt.Sprintf("drop table %s", oldTable))
	if err != nil {
		return err
	}
	_, err = mysqlEngine.Exec(fmt.Sprintf("ALTER TABLE %s RENAME `%s`", newTable, oldTable))
	return err
}

func initJobRecordTable() error {
	_, err := mysqlEngine.Exec("CREATE TABLE IF NOT EXISTS `job_record` (`id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,`job_time` VARCHAR(64) unique not null,`host_ip` VARCHAR(255) NOT NULL,`update_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,  PRIMARY KEY (`id`)) ENGINE=INNODB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8")
	if err != nil {
		err = fmt.Errorf("init job_record table error: %v", err)
	}
	return err
}

func checkJobState(jobId string) bool {
	ipInt, _ := strconv.Atoi(strings.Replace(hostIp, ".", "", -1))
	rand.Seed(time.Now().UnixNano() + int64(ipInt))
	waitSecond := rand.Intn(10)
	log.Printf("host:%s run wait job %d second...\n", hostIp, waitSecond)
	time.Sleep(time.Duration(waitSecond) * time.Second)
	var jobTables []*JobRecordTable
	mysqlEngine.SQL(fmt.Sprintf("SELECT * FROM job_record WHERE job_time='%s'", jobId)).Find(&jobTables)
	if len(jobTables) > 0 {
		return false
	}
	_, err := mysqlEngine.Exec(fmt.Sprintf("INSERT INTO job_record(job_time,host_ip,update_at) VALUE ('%s','%s','%s')", jobId, hostIp, time.Now().Format("2006-01-02 15:04:05")))
	if err != nil {
		log.Printf("update job_record table with host:%s error: %v \n", hostIp, err)
		return false
	}
	return true
}

func transUnixTime(input int64) (output string) {
	output = time.Unix(input, 0).Format("2006-01-02 15:04:05")
	return
}
