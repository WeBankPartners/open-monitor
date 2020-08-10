package transfer

import (
	"github.com/gin-gonic/gin"
	m "github.com/WeBankPartners/open-monitor/monitor-agent/transgateway/models"
	"fmt"
	"net/http"
)

func DisplayMetrics(c *gin.Context)  {
	var outputString string
	for _,v := range m.DataCache {
		if !v.Active {
			continue
		}
		v.Lock.RLock()
		for _,vv := range v.Metrics {
			if !vv.Active {
				continue
			}
			outputString += fmt.Sprintf("# TYPE %s gauge\n", vv.Metric)
			outputString += fmt.Sprintf("%s{system=\"%s\",interface=\"%s\"} %.3f \n", vv.Metric, v.Name, vv.InterfaceName, vv.Value)
		}
		v.Lock.RUnlock()
	}
	c.Header("Transfer-Encoding", "chunked")
	c.String(http.StatusOK, outputString)
}
