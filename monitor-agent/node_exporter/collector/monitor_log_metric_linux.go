package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"sync"
)

const (
	log_metricCollectorName   = "log_metric_monitor"
	log_metricMonitorFilePath = "data/log_metric_monitor_cache.data"
)

var (
	logMetricMonitorJobs       []*logMetricMonitorObj
	loMetricMonitorLock       = new(sync.RWMutex)
	logMetricMonitorMetrics    []*logMetricRuleMetricObj
	logMetricMonitorMetricLock = new(sync.RWMutex)
	monitorLogger                 log.Logger
)

type logMetricMonitorCollector struct {
	logMetricMonitor *prometheus.Desc
	logger          log.Logger
}

func InitMonitorLogger(logger log.Logger) {
	monitorLogger = logger
}

func init() {
	registerCollector(log_metricCollectorName, defaultEnabled, LogMetricMonitorCollector)
}

func LogMetricMonitorCollector(logger log.Logger) (Collector, error) {
	return &logMetricMonitorCollector{
		logMetricMonitor: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, log_metricCollectorName, "value"),
			"Show log_metric data from log file.",
			[]string{"key", "tags", "path", "agg"}, nil,
		),
		logger: logger,
	}, nil
}

func (c *logMetricMonitorCollector) Update(ch chan<- prometheus.Metric) error {
	logMetricMonitorMetricLock.RLock()
	for _, v := range logMetricMonitorMetrics {
		ch <- prometheus.MustNewConstMetric(c.logMetricMonitor,
			prometheus.GaugeValue,
			v.Value, v.Metric, v.TagsString, v.Path, v.Agg)
	}
	logMetricMonitorMetricLock.RUnlock()
	return nil
}
