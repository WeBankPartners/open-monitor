package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/api"
	"github.com/WeBankPartners/open-monitor/monitor-server/api/v1/alarm"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	ds "github.com/WeBankPartners/open-monitor/monitor-server/services/datasource"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/k8s"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/other"
	"go.uber.org/zap"
	"time"
)

// @title Monitor Server API
// @version 1.0
// @description 监控门户后台服务.
// @termsOfService http://swagger.io/terms/

// @host monitor.test.com
// @BasePath /v1
func main() {
	cfgFile := flag.String("c", "conf/default.json", "config file")
	//exportAgent := flag.Bool("export_agent", false, "true or false to choice export agent")
	flag.Parse()
	m.InitConfig(*cfgFile)
	if err := log.InitLogger(); err != nil {
		fmt.Printf("Server  init loggers failed, err: %v\n", err)
		return
	}
	defer log.SyncLoggers()
	if err := db.InitDatabase(); err != nil {
		panic("database initialization failed")
	}
	ds.InitPrometheusDatasource()
	if m.Config().KubernetesMonitor.Enabled {
		index := k8s.NewIndex()
		exporter := k8s.NewExporter(index)
		k8s.ConfigureLogTargetResolver(k8s.CMDBLogTargetResolver{Client: k8s.CMDBClient{BaseURL: m.CoreUrl, Token: m.GetCoreToken}})
		exporter.SetFreshness(m.Config().KubernetesMonitor.MetricsInterval() * 2)
		k8s.ConfigureRuntime(index, exporter)
		go k8s.RunMetrics(context.Background(), m.Config().KubernetesMonitor, db.KubernetesMonitoringStore(), index, exporter, func(err error) { log.Error(nil, log.LOGGER_APP, "Kubernetes metrics failed", zap.Error(err)) })
		ctx := context.Background()
		go k8s.RunDiscovery(ctx, m.Config().KubernetesMonitor, db.KubernetesMonitoringStore(), index, func(err error) {
			log.Error(nil, log.LOGGER_APP, "Kubernetes discovery failed", zap.Error(err))
		})
		reader := k8s.PrometheusMetricReader{
			Interval: m.Config().KubernetesMonitor.MetricsInterval(),
			Query: func(ctx context.Context, query string) (map[string]interface{}, error) {
				return ds.QueryPrometheusInstantRawContext(ctx, query, "")
			},
		}
		sender := k8s.NotificationSenderFunc(func(_ context.Context, notification k8s.Notification) error {
			if notification.StrategyID == 0 {
				return nil
			}
			status := "firing"
			if notification.Event == k8s.AlertEventRecovered {
				status = "ok"
			}
			raw, err := json.Marshal(notification.Payload)
			if err != nil {
				return err
			}
			var body struct {
				Key   k8s.AlertKey `json:"key"`
				Value *float64     `json:"value"`
			}
			if err = json.Unmarshal(raw, &body); err != nil {
				return err
			}
			value := 0.0
			if body.Value != nil {
				value = *body.Value
			}
			now := time.Now()
			alarmObj := &m.AlarmHandleObj{AlarmTable: m.AlarmTable{StrategyId: notification.StrategyID, Endpoint: body.Key.Endpoint(), Status: status, SMetric: body.Key.RuleID, StartValue: value, EndValue: value, Start: now, End: now, Content: fmt.Sprintf("Kubernetes alert: cluster=%s namespace=%s workload=%s/%s pod=%s container=%s", body.Key.ClusterID, body.Key.Namespace, body.Key.WorkloadType, body.Key.WorkloadName, body.Key.PodName, body.Key.Container)}, NotifyEnable: 1}
			return db.NotifyAlarmWithError(alarmObj)
		})
		alertStore := db.KubernetesAlertStore()
		go k8s.RunAlerts(ctx, m.Config().KubernetesMonitor, db.KubernetesMonitoringStore(), k8s.IndexAlertSnapshotSource{Index: index}, alertStore, alertStore, sender, reader, func(err error) {
			log.Error(nil, log.LOGGER_APP, "Kubernetes alert failed", zap.Error(err))
		})
	}
	if m.Config().Http.Session.Enable == "true" {
		middleware.InitSession()
	}
	if m.Config().Alert.Enable {
		other.InitSmtpMail()
	}
	go api.InitClusterApi()
	go db.InitPrometheusConfig()
	go db.InitSysParameter()
	go db.InitServiceGroup()
	go db.StartCronJob()
	go db.StartCheckCron()
	go db.StartLogKeywordMonitorCronJob()
	go db.StartDbKeywordMonitorCronJob()
	go alarm.StartAlarmEngineCron()
	go db.SyncDbMetric(true)
	go db.StartCallCronJob()
	go db.StartNotifyPingExport()
	go api.InitDependenceParam()
	go db.StartInitAlarmUniqueTags()
	go db.SyncMetricComparison()

	api.InitHttpServer()
}
