FROM  ccr.ccs.tencentyun.com/webankpartners/wecube-prometheus-arm64:v1.3
LABEL maintainer = "Webank CTB Team"

ENV JAVA_HOME=/opt/jdk
ENV PATH=$PATH:/opt/jdk/bin
ENV BASE_HOME=/app/monitor
ENV PROMETHEUS_HOME=$BASE_HOME/prometheus
ENV ALERTMANAGER_HOME=$BASE_HOME/alertmanager
ENV MONITOR_HOME=$BASE_HOME/monitor
ENV AGENT_MANAGER_HOME=$BASE_HOME/agent_manager
ENV AGENT_MANAGER_DEPLOY=/app/deploy
ENV TRANS_GATEWAY=$BASE_HOME/transgateway
ENV PING_EXPORTER=$BASE_HOME/ping_exporter
ENV ARCHIVE_TOOL=$BASE_HOME/archive_mysql_tool
ENV DB_DATA_EXPORTER=$BASE_HOME/db_data_exporter
ENV DAEMON_PROC=$BASE_HOME/daemon_proc
ENV METRIC_COMPARISON_EXPORTER=$BASE_HOME/metric_comparison_exporter

COPY build/start.sh $BASE_HOME/
COPY build/stop.sh $BASE_HOME/
COPY build/conf/prometheus.yml $PROMETHEUS_HOME/
COPY build/conf/kubernetes_prometheus.tpl $PROMETHEUS_HOME/
COPY build/conf/snmp_prometheus.tpl $PROMETHEUS_HOME/
COPY build/conf/remote_write_prometheus.tpl $PROMETHEUS_HOME/
COPY build/conf/prometheus.yml $PROMETHEUS_HOME/prometheus_tpl.yml
COPY build/conf/sd_file $PROMETHEUS_HOME/sd_file
COPY build/conf/alertmanager.yml $ALERTMANAGER_HOME/
COPY build/conf/base.yml $PROMETHEUS_HOME/
COPY monitor-server/monitor-server $MONITOR_HOME/
COPY build/conf/monitor.json $MONITOR_HOME/conf/default.json
COPY monitor-server/conf/i18n $MONITOR_HOME/conf/i18n
COPY monitor-ui/dist $MONITOR_HOME/public
COPY monitor-agent/agent_manager/agent_manager $AGENT_MANAGER_HOME/
COPY monitor-agent/agent_manager/exporters.tar.gz $AGENT_MANAGER_HOME/
COPY build/conf/agent_manager.json $AGENT_MANAGER_HOME/conf.json
COPY monitor-agent/transgateway/transgateway $TRANS_GATEWAY/
COPY monitor-agent/ping_exporter/ping_exporter $PING_EXPORTER/
COPY build/conf/ping_exporter.json $PING_EXPORTER/cfg.json
COPY monitor-agent/archive_mysql_tool/archive_mysql_tool $ARCHIVE_TOOL/
COPY build/conf/archive_mysql_tool.json $ARCHIVE_TOOL/default.json
COPY build/conf/core-site.xml $ARCHIVE_TOOL/conf/core-site.xml
COPY build/conf/hdfs-site.xml $ARCHIVE_TOOL/conf/hdfs-site.xml
COPY build/conf/krb5.conf $ARCHIVE_TOOL/conf/krb5.conf
COPY monitor-agent/db_data_exporter/db_data_exporter $DB_DATA_EXPORTER/
COPY monitor-agent/daemon_proc/daemon_proc $DAEMON_PROC/
COPY monitor-agent/daemon_proc/config.json $DAEMON_PROC/
COPY monitor-agent/metric_comparison_exporter/metric_comparison $METRIC_COMPARISON_EXPORTER/
COPY monitor-server/conf/menu-api-map.json $MONITOR_HOME/conf/

WORKDIR $BASE_HOME

ENTRYPOINT ["/bin/sh", "start.sh"]