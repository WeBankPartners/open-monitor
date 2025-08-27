package funcs

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"database/sql"
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

	// 连接复用趋势监控
	lastConnectionStats sql.DBStats
	connectionTrendData []ConnectionTrendPoint
)

const (
	// WaitCountThreshold 等待连接数阈值，超过此值才触发打印
	WaitCountThreshold = 500
)

// ConnectionTrendPoint 连接趋势数据点
type ConnectionTrendPoint struct {
	Timestamp        time.Time
	OpenConnections  int
	InUseConnections int
	IdleConnections  int
	WaitCount        int64
	WaitDuration     time.Duration
	UsageRate        float64
	ReuseRate        float64
}

func InitDbEngine(databaseName string) (err error) {
	if databaseName == "" {
		databaseName = "mysql"
	}

	// 关闭旧的连接引擎
	if mysqlEngine != nil {
		log.Printf("InitDbEngine - Closing old engine for database: %s", databaseSelect)
		err := mysqlEngine.Close()
		if err != nil {
			log.Printf("InitDbEngine - Error closing old engine: %v", err)
		}
		// 等待连接关闭
		time.Sleep(1 * time.Second)
	}

	connectStr := fmt.Sprintf("%s:%s@%s(%s:%s)/%s?collation=utf8mb4_unicode_ci&allowNativePasswords=true",
		Config().Mysql.User, Config().Mysql.Password, "tcp", Config().Mysql.Server, Config().Mysql.Port, databaseName)
	mysqlEngine, err = xorm.NewEngine("mysql", connectStr)
	if err != nil {
		log.Printf("init mysql fail with connect: %s error: %v \n", connectStr, err)
	} else {
		maxOpen := Config().Mysql.MaxOpen
		maxIdle := Config().Mysql.MaxIdle
		timeout := Config().Mysql.Timeout
		if maxOpen <= 0 {
			maxOpen = 150
		}
		if maxIdle <= 0 {
			maxIdle = 100
		}
		if timeout <= 0 {
			timeout = 60
		}
		log.Printf("InitDbEngine - Setting MySQL connection pool - MaxOpen: %d, MaxIdle: %d, Timeout: %d", maxOpen, maxIdle, timeout)
		mysqlEngine.SetMaxIdleConns(maxIdle)
		mysqlEngine.SetMaxOpenConns(maxOpen)
		mysqlEngine.SetConnMaxLifetime(time.Duration(timeout) * time.Second)
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
	log.Println("start reset db engine...")

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
		log.Printf("ResetDbEngine - Setting MySQL connection pool - MaxOpen: %d, MaxIdle: %d, Timeout: %d", Config().Mysql.MaxOpen, Config().Mysql.MaxIdle, Config().Mysql.Timeout)
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
		log.Printf("InitMonitorDbEngine - Setting Monitor MySQL connection pool - MaxOpen: %d, MaxIdle: %d, Timeout: %d", Config().Monitor.Mysql.MaxOpen, Config().Monitor.Mysql.MaxIdle, Config().Monitor.Mysql.Timeout)
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
		var tmpErr error
		// _, tmpErr := mysqlEngine.Exec(v)
		for i := 0; i < 3; i++ {
			//log.Printf("start try %d to insert mysql,data num:%d \n", i+1, rowCountList[sqlIndex])
			_, err := mysqlEngine.Exec(v)
			if err != nil {
				tmpErr = err
			} else {
				tmpErr = nil
				break
			}
			if i < 2 {
				time.Sleep(time.Duration(retryWaitSecond) * time.Second)
			}
		}
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

// PrintDBConnectionStatsConditional 条件性打印数据库连接池统计信息（性能优化版本）
func PrintDBConnectionStatsConditional(prefix string, forcePrint bool) {
	if mysqlEngine == nil {
		return
	}

	stats := mysqlEngine.DB().Stats()

	// 只在以下情况打印详细日志：
	// 1. 强制打印（forcePrint=true）
	// 2. 连接使用率超过80%
	// 3. 等待连接数超过阈值（默认10）
	// 4. 连接数接近最大值
	shouldPrint := forcePrint
	if !shouldPrint && stats.MaxOpenConnections > 0 {
		usageRate := float64(stats.OpenConnections) / float64(stats.MaxOpenConnections)
		shouldPrint = usageRate > 0.8 || stats.WaitCount > WaitCountThreshold || stats.OpenConnections >= stats.MaxOpenConnections-2
	}

	if shouldPrint {
		log.Printf("[%s] DB Connection Stats - Open: %d/%d, InUse: %d, Idle: %d, WaitCount: %d, WaitDuration: %v",
			prefix,
			stats.OpenConnections,
			stats.MaxOpenConnections,
			stats.InUse,
			stats.Idle,
			stats.WaitCount,
			stats.WaitDuration)

		if stats.MaxOpenConnections > 0 {
			usageRate := float64(stats.OpenConnections) / float64(stats.MaxOpenConnections) * 100
			log.Printf("[%s] DB Connection Usage Rate: %.2f%%", prefix, usageRate)
		}

		if stats.WaitCount > 0 {
			avgWaitTime := stats.WaitDuration / time.Duration(stats.WaitCount)
			log.Printf("[%s] DB Connection Avg Wait Time: %v", prefix, avgWaitTime)
		}

		// 添加连接复用情况监控
		printConnectionReuseStats(prefix, stats)

		// 只在强制打印时调用趋势分析
		if forcePrint {
			PrintConnectionTrendAnalysis(prefix)
		}
	}
}

// printConnectionReuseStats 打印连接复用统计信息
func printConnectionReuseStats(prefix string, stats sql.DBStats) {
	// 基础连接统计
	log.Printf("[%s] === DB Connection Reuse Analysis ===", prefix)
	log.Printf("[%s] Basic Stats - MaxOpen: %d, Open: %d, InUse: %d, Idle: %d",
		prefix, stats.MaxOpenConnections, stats.OpenConnections, stats.InUse, stats.Idle)

	if stats.MaxOpenConnections > 0 {
		// 连接池利用率 = 当前连接数 / 最大连接数
		poolUtilization := float64(stats.OpenConnections) / float64(stats.MaxOpenConnections) * 100

		// 连接复用率 = 空闲连接数 / 总连接数（如果总连接数>0）
		var reuseRate float64
		if stats.OpenConnections > 0 {
			reuseRate = float64(stats.Idle) / float64(stats.OpenConnections) * 100
		}

		// 连接压力指数 = 使用中连接数 / 最大连接数
		pressureIndex := float64(stats.InUse) / float64(stats.MaxOpenConnections) * 100

		log.Printf("[%s] Connection Metrics - Pool Utilization: %.2f%%, Reuse Rate: %.2f%%, Pressure Index: %.2f%%",
			prefix, poolUtilization, reuseRate, pressureIndex)

		// 连接池健康度评估
		healthStatus := "Healthy"
		if pressureIndex > 90 {
			healthStatus = "Critical"
		} else if pressureIndex > 70 {
			healthStatus = "Warning"
		} else if reuseRate < 20 && stats.OpenConnections > 0 {
			healthStatus = "Low Reuse"
		}

		log.Printf("[%s] Health Status: %s", prefix, healthStatus)

		// 连接池饱和度分析
		saturationLevel := "Low"
		if poolUtilization > 90 {
			saturationLevel = "Critical"
		} else if poolUtilization > 70 {
			saturationLevel = "High"
		} else if poolUtilization > 50 {
			saturationLevel = "Medium"
		}

		log.Printf("[%s] Pool Saturation: %s (%.2f%%)", prefix, saturationLevel, poolUtilization)
	}

	// 等待连接情况分析
	if stats.WaitCount > 0 {
		avgWaitTime := stats.WaitDuration / time.Duration(stats.WaitCount)
		waitSeverity := "Low"
		if avgWaitTime > 5*time.Second {
			waitSeverity = "Critical"
		} else if avgWaitTime > 1*time.Second {
			waitSeverity = "High"
		} else if avgWaitTime > 100*time.Millisecond {
			waitSeverity = "Medium"
		}

		log.Printf("[%s] Wait Analysis - Severity: %s, Count: %d, Avg Wait: %v, Total Wait: %v",
			prefix, waitSeverity, stats.WaitCount, avgWaitTime, stats.WaitDuration)
	} else {
		log.Printf("[%s] Wait Analysis - No connection waits detected", prefix)
	}

	// 连接池效率指标
	if stats.OpenConnections > 0 {
		// 活跃连接比例
		activeRatio := float64(stats.InUse) / float64(stats.OpenConnections) * 100
		// 空闲连接比例
		idleRatio := float64(stats.Idle) / float64(stats.OpenConnections) * 100

		log.Printf("[%s] Connection Distribution - Active: %.2f%% (%d), Idle: %.2f%% (%d)",
			prefix, activeRatio, stats.InUse, idleRatio, stats.Idle)
	}

	// 性能建议
	log.Printf("[%s] === Performance Recommendations ===", prefix)
	if stats.MaxOpenConnections > 0 {
		utilization := float64(stats.OpenConnections) / float64(stats.MaxOpenConnections)
		if utilization > 0.8 {
			log.Printf("[%s] Consider increasing MaxOpenConnections (current: %d)",
				prefix, stats.MaxOpenConnections)
		} else if utilization < 0.2 && stats.OpenConnections > 5 {
			log.Printf("[%s] Consider decreasing MaxOpenConnections (current: %d)",
				prefix, stats.MaxOpenConnections)
		}

		if stats.WaitCount > 0 {
			log.Printf("[%s] Connection pool may be undersized, consider tuning pool parameters", prefix)
		}

		if stats.Idle > stats.MaxOpenConnections/2 {
			log.Printf("[%s] High idle connections, consider reducing MaxIdleConns", prefix)
		}
	}

	log.Printf("[%s] === End Analysis ===", prefix)
}

// PrintConnectionTrendAnalysis 打印连接复用趋势分析
func PrintConnectionTrendAnalysis(prefix string) {
	if mysqlEngine == nil {
		return
	}

	currentStats := mysqlEngine.DB().Stats()

	// 计算当前指标
	var currentUsageRate, currentReuseRate float64
	if currentStats.MaxOpenConnections > 0 {
		currentUsageRate = float64(currentStats.OpenConnections) / float64(currentStats.MaxOpenConnections) * 100
	}
	if currentStats.OpenConnections > 0 {
		currentReuseRate = float64(currentStats.Idle) / float64(currentStats.OpenConnections) * 100
	}

	// 创建当前数据点
	currentPoint := ConnectionTrendPoint{
		Timestamp:        time.Now(),
		OpenConnections:  currentStats.OpenConnections,
		InUseConnections: currentStats.InUse,
		IdleConnections:  currentStats.Idle,
		WaitCount:        currentStats.WaitCount,
		WaitDuration:     currentStats.WaitDuration,
		UsageRate:        currentUsageRate,
		ReuseRate:        currentReuseRate,
	}

	// 添加到趋势数据
	connectionTrendData = append(connectionTrendData, currentPoint)

	// 保持最近10个数据点
	if len(connectionTrendData) > 10 {
		connectionTrendData = connectionTrendData[1:]
	}

	// 分析趋势
	if len(connectionTrendData) >= 2 {
		analyzeConnectionTrend(prefix, connectionTrendData)
	}

	// 更新上次统计
	lastConnectionStats = currentStats
}

// analyzeConnectionTrend 分析连接复用趋势
func analyzeConnectionTrend(prefix string, trendData []ConnectionTrendPoint) {
	if len(trendData) < 2 {
		return
	}

	latest := trendData[len(trendData)-1]
	previous := trendData[len(trendData)-2]

	log.Printf("[%s] === Connection Trend Analysis ===", prefix)

	// 连接数变化趋势
	openChange := latest.OpenConnections - previous.OpenConnections
	openTrend := "Stable"
	if openChange > 2 {
		openTrend = "Increasing"
	} else if openChange < -2 {
		openTrend = "Decreasing"
	}

	log.Printf("[%s] Connection Count Trend - %s (change: %+d)", prefix, openTrend, openChange)

	// 使用率变化趋势
	usageChange := latest.UsageRate - previous.UsageRate
	usageTrend := "Stable"
	if usageChange > 5 {
		usageTrend = "Increasing"
	} else if usageChange < -5 {
		usageTrend = "Decreasing"
	}

	log.Printf("[%s] Usage Rate Trend - %s (change: %+.2f%%)", prefix, usageTrend, usageChange)

	// 复用率变化趋势
	reuseChange := latest.ReuseRate - previous.ReuseRate
	reuseTrend := "Stable"
	if reuseChange > 5 {
		reuseTrend = "Improving"
	} else if reuseChange < -5 {
		reuseTrend = "Deteriorating"
	}

	log.Printf("[%s] Reuse Rate Trend - %s (change: %+.2f%%)", prefix, reuseTrend, reuseChange)

	// 等待连接趋势
	waitChange := latest.WaitCount - previous.WaitCount
	waitTrend := "Stable"
	if waitChange > 0 {
		waitTrend = "Increasing"
	} else if waitChange < 0 {
		waitTrend = "Decreasing"
	}

	log.Printf("[%s] Wait Count Trend - %s (change: %+d)", prefix, waitTrend, waitChange)

	// 性能趋势评估
	performanceTrend := "Good"
	if latest.UsageRate > 90 || latest.WaitCount > 0 || latest.ReuseRate < 20 {
		performanceTrend = "Concerning"
	}
	if latest.UsageRate > 95 || latest.WaitCount > 5 {
		performanceTrend = "Poor"
	}

	log.Printf("[%s] Overall Performance Trend: %s", prefix, performanceTrend)
	log.Printf("[%s] === End Trend Analysis ===", prefix)
}
