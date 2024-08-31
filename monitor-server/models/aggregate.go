package models

import (
	"fmt"
	"sort"
	"strconv"
)

func Aggregate(data [][]float64, step int64, method string) [][]float64 {
	if len(data) == 0 || step <= 10 {
		return data
	}
	step = step * 1000
	var result [][]float64
	var tmpV []float64
	var start, end float64
	start = data[0][0] - float64(int64(data[0][0])%step)
	end = start + float64(step)
	for _, v := range data {
		if v[0] >= end {
			result = append(result, []float64{end, CalcData(tmpV, method)})
			start = end
			end = start + float64(step)
			if v[0] > end {
				for v[0] > end {
					start = end
					end = start + float64(step)
				}
			}
			tmpV = []float64{}
		}
		tmpV = append(tmpV, v[1])
	}
	if len(tmpV) > 0 {
		result = append(result, []float64{end, CalcData(tmpV, method)})
	}
	return result
}

func CalcData(data []float64, method string) float64 {
	var result float64
	if method == "avg" {
		var sum float64
		for _, v := range data {
			sum = sum + v
		}
		result = sum / float64(len(data))
	} else if method == "avg_nonzero" {
		var sum, pointNum float64
		for _, v := range data {
			if v != 0 {
				sum = sum + v
				pointNum += 1
			}
		}
		if pointNum > 0 {
			result = sum / pointNum
		}
	} else if method == "max" {
		result = data[0]
		for _, v := range data {
			if v > result {
				result = v
			}
		}
	} else if method == "min" {
		result = data[0]
		for _, v := range data {
			if v < result {
				result = v
			}
		}
	} else if method == "p95" {
		sort.Float64s(data)
		if len(data) == 1 {
			result = data[0]
		} else if len(data) <= 5 {
			result = data[len(data)-1]
		} else {
			index := len(data) * 95 / 100
			result = data[index-1]
		}
	} else if method == "sum" {
		for _, v := range data {
			result = result + v
		}
	}
	result, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", result), 64)
	return result
}
