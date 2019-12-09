package db

import (
	"sort"
	"strconv"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
)

func CheckAggregate(start int64, end int64, endpoint string, step,num int) int {
	if num == 0 {
		return 0
	}
	if step <= 0 {
		host := models.EndpointTable{Guid: endpoint}
		GetEndpoint(&host)
		if host.Id == 0 {
			return 0
		}
		step = host.Step
	}
	var agg int
	limit := 5
	subtime := end - start
	hour := int64(3600)
	if subtime <= hour {
		if num <=limit {
			agg = 0
		}else{
			if subtime <= int64(1800) {
				agg = 0
			}else {
				agg = 6 //1min
			}
		}
	}else if subtime <= 6*hour {
		if num <=limit {
			agg = 6
		}else{
			agg = 30 //5min
		}
	}else if subtime <= 24*hour {
		if num <=limit {
			agg = 30
		}else{
			agg = 60 //10min
		}
	}else if subtime <= 3*24*hour {
		if num <=limit {
			agg = 60
		}else{
			agg = 360 //60min
		}
	}else if subtime <= 7*24*hour {  //graph data
		if num <=limit {
			agg = 180 //30min
		}else{
			agg = 720 //120min
		}
	}else if subtime <= 14*24*hour {
		if num <=limit {
			agg = 360 //60min
		}else{
			agg = 2160 //6hour
		}
	}else if subtime <= 90*24*hour {
		if num <=limit {
			agg = 2160 //6hour
		}else{
			agg = 8640 //1day
		}
	}else if subtime <= 180*24*hour {
		agg = 8640
	}
	if step > 10 && subtime <= 3*24*hour {
		agg = agg/(step/10)
	}
	return agg
}

func Aggregate(data [][]float64, num int, method string) [][]float64 {
	var result [][]float64
	var tmpV []float64
	var tmpT float64
	tmpI := 0
	numH := num/2
	for i,v := range data {
		tmpI += 1
		tmpV = append(tmpV, v[1])
		if (i+1)%num == 0 && i != 0 {
			tmpI = 0
			result = append(result, []float64{tmpT, calc(tmpV, method)})
			tmpV = []float64{}
		}
		if tmpI%numH==0 {
			tmpT = v[0]
		}
	}
	return result
}

func calc(data []float64, method string) float64 {
	var result float64
	if method == "avg" {
		var sum float64
		for _,v := range data {
			sum = sum + v
		}
		result = sum/float64(len(data))
	}else if method == "max"{
		result = data[0]
		for _,v := range data {
			if v > result {
				result = v
			}
		}
	}else if method == "min" {
		result = data[0]
		for _,v := range data {
			if v < result {
				result = v
			}
		}
	}else if method == "p95" {
		sort.Float64s(data)
		if len(data) == 1 {
			result = data[0]
		}else if len(data) <=5 {
			result = data[len(data)-1]
		}else{
			index := len(data)*95/100
			result = data[index-1]
		}
	}
	result,_ = strconv.ParseFloat(fmt.Sprintf("%.3f", result), 64)
	return result
}