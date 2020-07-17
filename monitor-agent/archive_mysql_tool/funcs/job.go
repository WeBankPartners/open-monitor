package funcs

import (
	"log"
	"sort"
	"time"
	"fmt"
)

var jobChannelList chan ArchiveActionList

func StartCronJob()  {
	jobChannelList = make(chan ArchiveActionList, Config().Prometheus.MaxHttpOpen)
	go consumeJob()
	t,_ := time.Parse("2006-01-02 15:04:05 MST", fmt.Sprintf("%s 00:00:00 CST", time.Now().Format("2006-01-02")))
	subSecond := t.Unix()+86410-time.Now().Unix()
	time.Sleep(time.Duration(subSecond)*time.Second)
	c := time.NewTicker(24*time.Hour).C
	for {
		go CreateJob()
		<- c
	}
}

func CreateJob()  {
	log.Printf("start cron job \n")
	err,tableName := createTable()
	if err != nil {
		log.Printf("try to create table:%s error:%v \n", tableName, err)
		return
	}
	var unitPerJob int
	unitCount := 0
	for _,v := range MonitorObjList {
		unitPerJob += len(v.Metrics)
	}
	unitPerJob = unitPerJob/Config().Prometheus.MaxHttpOpen
	if unitPerJob > Config().Trans.MaxUnitSpeed {
		unitPerJob = Config().Trans.MaxUnitSpeed
	}
	if unitPerJob == 0 {
		unitPerJob = 1
	}
	var actionParamList []*ArchiveActionList
	var tmpActionParamObjList []*ArchiveActionParamObj
	for _,v := range MonitorObjList {
		for _,vv := range v.Metrics {
			unitCount++
			tmpActionParamObjList = append(tmpActionParamObjList, &ArchiveActionParamObj{Endpoint:v.Endpoint, Metric:vv.Metric, PromQl:vv.PromQl, TableName:tableName})
			if unitCount == unitPerJob {
				tmpArchiveActionList := ArchiveActionList{}
				for _,vvv := range tmpActionParamObjList {
					tmpArchiveActionList = append(tmpArchiveActionList, vvv)
				}
				actionParamList = append(actionParamList, &tmpArchiveActionList)
				tmpActionParamObjList = []*ArchiveActionParamObj{}
				unitCount = 0
			}
		}
	}
	go checkJobStatus()
	for _,v := range actionParamList {
		jobChannelList <- *v
	}
}

func consumeJob()  {
	for {
		param := <- jobChannelList
		go archiveAction(param)
	}
}

func checkJobStatus()  {
	time.Sleep(2*time.Second)
	for {
		log.Printf("job channel list length --> %d \n", len(jobChannelList))
		if len(jobChannelList) == 0 {
			log.Printf("archive job done \n")
			break
		}
	}
}

func archiveAction(param ArchiveActionList)  {
	if len(param) == 0 {
		return
	}
	var err error
	var rowData []*ArchiveTable
	for _,v :=range param {
		tmpPrometheusParam := PrometheusQueryParam{LastSecond:86400, PromQl:v.PromQl}
		err = getPrometheusData(&tmpPrometheusParam)
		if err != nil {
			log.Printf("acrhive action: endpoint->%s metric->%s get prometheus data error-> %v \n", v.Endpoint, v.Metric, err)
			continue
		}
		for _,vv := range tmpPrometheusParam.Data {
			tmpTagString := vv.Metric.ToTagString()
			tmpStartTime := vv.Start+60
			var tmpFloatList []float64
			for _,vvv := range vv.Values {
				if vvv[0] < float64(tmpStartTime) {
					tmpFloatList = append(tmpFloatList, vvv[1])
				}else{
					if len(tmpFloatList) > 0 {
						avg,min,max,p95 := calcData(tmpFloatList)
						rowData = append(rowData, &ArchiveTable{Endpoint:v.Endpoint,Metric:v.Metric,Tags:tmpTagString,UnixTime:tmpStartTime-60,Avg:avg,Min:min,Max:max,P95:p95})
					}
					tmpStartTime += 60
					tmpFloatList = []float64{vvv[1]}
				}
			}
			if len(tmpFloatList) > 0 {
				avg,min,max,p95 := calcData(tmpFloatList)
				rowData = append(rowData, &ArchiveTable{Endpoint:v.Endpoint,Metric:v.Metric,Tags:tmpTagString,UnixTime:tmpStartTime-60,Avg:avg,Min:min,Max:max,P95:p95})
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

func calcData(data []float64) (avg,min,max,p95 float64) {
	if len(data) == 1 {
		return data[0],data[0],data[0],data[0]
	}
	sort.Float64s(data)
	min = data[0]
	max = data[len(data)-1]
	p95 = data[len(data)-2]
	var sum float64
	for _,v := range data {
		sum += v
	}
	avg = sum/float64(len(data))
	return avg,min,max,p95
}