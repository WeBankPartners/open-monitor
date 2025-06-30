package funcs

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"
)

var jobChannelList chan ArchiveActionList
var archiveTime int64 = 3600

func StartCronJob() {
	concurrentInsertNum = 50
	if Config().Trans.ConcurrentInsertNum > 0 {
		concurrentInsertNum = Config().Trans.ConcurrentInsertNum
	}
	maxUnitNum = 5
	if Config().Trans.MaxUnitSpeed > 0 {
		maxUnitNum = Config().Trans.MaxUnitSpeed
	}
	retryWaitSecond = 60
	if Config().Trans.RetryWaitSecond > 0 {
		retryWaitSecond = Config().Trans.RetryWaitSecond
	}
	jobTimeout = 1800
	if Config().Trans.JobTimeout > 0 {
		jobTimeout = Config().Trans.JobTimeout
	}
	jobChannelList = make(chan ArchiveActionList, Config().Prometheus.MaxHttpOpen)
	go consumeJob()
	t, _ := time.Parse("2006-01-02 15:04:05 MST", fmt.Sprintf("%s :00:00 "+DefaultLocalTimeZone, time.Now().Format("2006-01-02 15")))
	subSecond := t.Unix() + archiveTime + 10 - time.Now().Unix()
	time.Sleep(time.Duration(subSecond) * time.Second)
	c := time.NewTicker(1 * time.Hour).C
	for {
		go func() {
			jobId := time.Now().Add(-1 * time.Hour).Format("2006-01-02_15")
			if checkJobState(jobId) {
				CreateJob("")
				time.Sleep(10 * time.Minute)
				ArchiveFromMysql(0)
			}
		}()
		<-c
	}
}

// jobId -> 2025-06-30_10
func CreateJob(jobId string) {
	log.Printf("start CreateJob %s \n", jobId)
	log.Printf("start InitMonitorMetricMap %s \n", jobId)
	err := InitMonitorMetricMap()
	if err != nil {
		log.Printf("fail InitMonitorMetricMap,init monitor metric map error: %v \n", err)
		return
	}
	log.Printf("end InitMonitorMetricMap %s \n", jobId)
	var start, end int64
	if jobId == "" {
		t, _ := time.Parse("2006-01-02_15:04:05 MST", fmt.Sprintf("%s:00:00 "+DefaultLocalTimeZone, time.Now().Format("2006-01-02_15")))
		start = t.Unix() - archiveTime
		end = t.Unix()
	} else {
		t, err := time.Parse("2006-01-02_15:04:05 MST", fmt.Sprintf("%s:00:00 "+DefaultLocalTimeZone, jobId))
		if err != nil {
			log.Printf("dateString validate fail,must format like 2006-01-02 \n")
			return
		}
		start = t.Unix()
		end = t.Unix() + archiveTime
	}
	log.Printf("start createTable start:%s end:%s \n", time.Unix(start, 0).Format("2006-01-02_15:04:05"), time.Unix(end, 0).Format("2006-01-02_15:04:05"))
	err, tableName := createTable(start, false)
	if err != nil {
		log.Printf("try to create table:%s error:%v \n", tableName, err)
		return
	}
	log.Printf("end createTable %s \n", jobId)
	unitCount := 0
	actionParamObjLength := maxUnitNum * Config().Prometheus.MaxHttpOpen
	var actionParamList []*ArchiveActionList
	var tmpActionParamObjList []*ArchiveActionParamObj
	for _, v := range MonitorObjList {
		for _, vv := range v.Metrics {
			unitCount++
			tmpActionParamObjList = append(tmpActionParamObjList, &ArchiveActionParamObj{Endpoint: v.Endpoint, Metric: vv.Metric, PromQl: vv.PromQl, TableName: tableName, Start: start, End: end})
			if unitCount == actionParamObjLength {
				tmpArchiveActionList := ArchiveActionList{}
				for _, vvv := range tmpActionParamObjList {
					tmpArchiveActionList = append(tmpArchiveActionList, vvv)
				}
				actionParamList = append(actionParamList, &tmpArchiveActionList)
				tmpActionParamObjList = []*ArchiveActionParamObj{}
				unitCount = 0
			}
		}
	}
	if len(tmpActionParamObjList) > 0 {
		tmpArchiveActionList := ArchiveActionList{}
		for _, vvv := range tmpActionParamObjList {
			tmpArchiveActionList = append(tmpArchiveActionList, vvv)
		}
		actionParamList = append(actionParamList, &tmpArchiveActionList)
	}
	go checkJobStatus()
	for _, v := range actionParamList {
		jobChannelList <- *v
	}
}

func consumeJob() {
	for {
		param := <-jobChannelList
		if len(param) == 0 {
			continue
		}
		tmpUnixCount := 0
		var concurrentJobList []ArchiveActionList
		tmpJobList := ArchiveActionList{}
		for _, v := range param {
			tmpUnixCount++
			tmpJobList = append(tmpJobList, v)
			if tmpUnixCount >= maxUnitNum {
				concurrentJobList = append(concurrentJobList, tmpJobList)
				tmpJobList = ArchiveActionList{}
				tmpUnixCount = 0
			}
		}
		if len(tmpJobList) > 0 {
			concurrentJobList = append(concurrentJobList, tmpJobList)
		}
		log.Printf("start consume job,length:%d ,concurrent:%d \n", len(param), len(concurrentJobList))
		startTime := time.Now()
		wg := sync.WaitGroup{}
		for _, job := range concurrentJobList {
			wg.Add(1)
			go func(jobList ArchiveActionList, tmpWg *sync.WaitGroup) {
				//archiveAction(jobList)
				archiveTimeoutAction(jobList)
				tmpWg.Done()
			}(job, &wg)
		}
		wg.Wait()
		endTime := time.Now()
		useTime := float64(endTime.Sub(startTime).Nanoseconds()) / 1e6
		log.Printf("done with consume job,use time: %.3f ms", useTime)
		if int(endTime.Sub(startTime).Seconds()) >= jobTimeout {
			log.Println("job timeout,try to reset db connection ")
			ResetDbEngine()
		}
		gTransport.CloseIdleConnections()
	}
}

func checkJobStatus() {
	time.Sleep(2 * time.Second)
	for {
		log.Printf("job channel list length --> %d \n", len(jobChannelList))
		if len(jobChannelList) == 0 {
			log.Printf("archive job done \n")
			break
		}
		time.Sleep(30 * time.Second)
	}
}

func archiveTimeoutAction(param ArchiveActionList) {
	ctx, cancel := context.WithCancel(context.Background())
	go func(p ArchiveActionList, c context.CancelFunc) {
		archiveAction(p)
		c()
	}(param, cancel)
	select {
	case <-ctx.Done():
		log.Printf("done archive action,job length:%d \n", len(param))
	case <-time.After(time.Duration(jobTimeout) * time.Second):
		log.Printf("timeout archive action in %d s,job length:%d \n", jobTimeout, len(param))
	}
}

func archiveAction(param ArchiveActionList) {
	log.Printf("start archive action,job length:%d \n", len(param))
	if len(param) == 0 {
		return
	}
	var err error
	var rowData []*ArchiveTable
	for i, v := range param {
		log.Printf("start build archive row data,index:%d \n", i)
		tmpPrometheusParam := PrometheusQueryParam{Start: v.Start, End: v.End, PromQl: v.PromQl}
		err = getPrometheusData(&tmpPrometheusParam)
		if err != nil {
			log.Printf("acrhive action: endpoint->%s metric->%s get prometheus data error-> %v \n", v.Endpoint, v.Metric, err)
			continue
		}
		for _, vv := range tmpPrometheusParam.Data {
			tmpTagString := vv.Metric.ToTagString()
			tmpStartTime := vv.Start + 60
			var tmpFloatList []float64
			for _, vvv := range vv.Values {
				pointTime := int64(vvv[0])
				if pointTime < tmpStartTime {
					tmpFloatList = append(tmpFloatList, vvv[1])
				} else {
					if len(tmpFloatList) > 0 {
						avg, min, max, p95, sum := calcData(tmpFloatList)
						rowData = append(rowData, &ArchiveTable{Endpoint: v.Endpoint, Metric: v.Metric, Tags: tmpTagString, UnixTime: tmpStartTime - 60, Avg: avg, Min: min, Max: max, P95: p95, Sum: sum})
					}
					pointStepTime := pointTime - 60
					for tmpStartTime < pointStepTime {
						tmpStartTime += 60
					}
					tmpStartTime += 60
					tmpFloatList = []float64{vvv[1]}
				}
			}
			if len(tmpFloatList) > 0 && tmpStartTime <= v.End {
				avg, min, max, p95, sum := calcData(tmpFloatList)
				rowData = append(rowData, &ArchiveTable{Endpoint: v.Endpoint, Metric: v.Metric, Tags: tmpTagString, UnixTime: tmpStartTime - 60, Avg: avg, Min: min, Max: max, P95: p95, Sum: sum})
			}
		}
	}
	if len(rowData) == 0 {
		log.Printf("acrhive action: endpoint->%s unit_num->%d row data is empty \n", param[0].Endpoint, len(param))
		return
	}
	err = insertMysql(rowData, param[0].TableName)
	if err != nil {
		log.Printf("acrhive action: endpoint->%s unit_num->%d row_num->%d insert to mysql error-> %v \n", param[0].Endpoint, len(param), len(rowData), err)
	}
}

func calcData(data []float64) (avg, min, max, p95, sum float64) {
	if len(data) == 1 {
		return data[0], data[0], data[0], data[0], data[0]
	}
	sort.Float64s(data)
	min = data[0]
	max = data[len(data)-1]
	p95 = data[len(data)-2]
	for _, v := range data {
		sum += v
	}
	avg = sum / float64(len(data))
	return avg, min, max, p95, sum
}

func ArchiveFromMysql(tableUnixTime int64) {
	if tableUnixTime <= 0 {
		var startDays int64 = 90
		if Config().Trans.FiveMinStartDay > 0 {
			startDays = Config().Trans.FiveMinStartDay
		}
		t, _ := time.Parse("2006-01-02 15:04:05 MST", fmt.Sprintf("%s 00:00:00 "+DefaultLocalTimeZone, time.Now().Format("2006-01-02")))
		tableUnixTime = t.Unix() - (startDays * 86400)
	}
	oldTableName := fmt.Sprintf("archive_%s", time.Unix(tableUnixTime, 0).Format("2006_01_02"))
	if !checkTableExists(oldTableName) {
		return
	}
	err, newTableName := createTable(tableUnixTime, true)
	if err != nil {
		log.Printf("archive 5 min job,create table:%s error:%v \n", newTableName, err)
		return
	}
	err, countNowTable := getArchiveTableCountData(oldTableName)
	if err != nil {
		log.Printf("archive 5 min job,get count data from table:%s error:%v \n", oldTableName, err)
		return
	}
	for _, v := range countNowTable {
		tmpErr := archiveOneToFive(oldTableName, newTableName, v.Endpoint, v.Metric)
		if tmpErr != nil {
			log.Printf("archive 5 min job,archive 1 min to 5 min job error: %v \n", tmpErr)
		}
	}
	err = renameFiveToOne(oldTableName, newTableName)
	if err != nil {
		log.Printf("archive 5 min job,rename %s to %s error: %v \n", oldTableName, newTableName, err)
	}
}
