package funcs

import (
	"bytes"
	"fmt"
)

func GetExportMetric() []byte {
	var buff bytes.Buffer
	buff.WriteString("# HELP ping check 0 -> alive, 1 -> dead, 2 -> problem. \n")
	resultLock.RLock()
	for _, v := range resultList {
		tmpMetricDisplay := metricString
		if v.KeywordFlag {
			tmpMetricDisplay = dbKeywordMetric
		}
		buff.WriteString(fmt.Sprintf("%s{key=\"%s\",t_endpoint=\"%s\",address=\"%s:%s\",service_group=\"%s\"} %s \n", tmpMetricDisplay, v.Name, v.Endpoint, v.Server, v.Port, v.ServiceGroup, transFloatValueToString(v.Value)))
	}
	resultLock.RUnlock()
	return buff.Bytes()
}

func transFloatValueToString(input float64) string {
	outputString := fmt.Sprintf("%.6f", input)
	for i := 0; i < 6; i++ {
		if outputString[len(outputString)-1:] == "0" {
			outputString = outputString[:len(outputString)-1]
		} else {
			break
		}
	}
	if outputString[len(outputString)-1:] == "." {
		outputString = outputString[:len(outputString)-1]
	}
	return outputString
}
