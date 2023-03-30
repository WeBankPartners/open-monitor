package transfer

import (
	"fmt"
	m "github.com/WeBankPartners/open-monitor/monitor-agent/transgateway/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func DisplayMetrics(c *gin.Context) {
	var outputString string
	for _, v := range m.DataCache {
		if !v.Active {
			continue
		}
		v.Lock.RLock()
		for _, vv := range v.Metrics {
			if !vv.Active {
				continue
			}
			outputString += fmt.Sprintf("# TYPE %s gauge\n", vv.Metric)
			outputString += fmt.Sprintf("%s{system=\"%s\",host=\"%s\",interface=\"%s\",object=\"%s\"} %.3f \n", vv.Metric, v.Name, vv.HostIp, vv.InterfaceName, vv.Object, vv.Value)
		}
		v.Lock.RUnlock()
	}
	c.Header("Transfer-Encoding", "chunked")
	c.String(http.StatusOK, outputString)
}
