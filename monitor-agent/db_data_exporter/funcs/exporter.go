package funcs

import (
	"bytes"
	"fmt"
)

func GetExportMetric() []byte {
	var buff bytes.Buffer
	buff.WriteString("# HELP ping check 0 -> alive, 1 -> dead, 2 -> problem. \n")
	resultLock.RLock()
	for _,v := range resultList {
		buff.WriteString(fmt.Sprintf("%s{name=\"%s\",guid=\"%s\",address=\"%s:%s\"} %d \n", metricString, v.Name, v.Endpoint, v.Server, v.Port, v.Value))
	}
	resultLock.RUnlock()
	return buff.Bytes()
}
