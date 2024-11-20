FROM  ccr.ccs.tencentyun.com/webankpartners/wecube-prometheus:v1.1
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

RUN mkdir -p $BASE_HOME $PROMETHEUS_HOME $PROMETHEUS_HOME/rules $PROMETHEUS_HOME/token $ALERTMANAGER_HOME $MONITOR_HOME $MONITOR_HOME/conf $AGENT_MANAGER_HOME $PING_EXPORTER $AGENT_MANAGER_DEPLOY $TRANS_GATEWAY $ARCHIVE_TOOL $DB_DATA_EXPORTER $DAEMON_PROC $METRIC_COMPARISON_EXPORTER $METRIC_COMPARISON_EXPORTER/config

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

RUN chmod +x $PROMETHEUS_HOME/prometheus $PROMETHEUS_HOME/promtool $ALERTMANAGER_HOME/alertmanager $AGENT_MANAGER_HOME/agent_manager $TRANS_GATEWAY/transgateway $MONITOR_HOME/monitor-server $BASE_HOME/*.sh $PING_EXPORTER/ping_exporter $ARCHIVE_TOOL/archive_mysql_tool $DB_DATA_EXPORTER/db_data_exporter $DAEMON_PROC/daemon_proc $METRIC_COMPARISON_EXPORTER/metric_comparison

WORKDIR $BASE_HOME
RUN addgroup -S apps -g 6000 && adduser -S app -u 6001 -G apps
RUN chown -R app:apps $BASE_HOME && chmod -R 755 $BASE_HOME
USER app
ENTRYPOINT ["/bin/sh", "start.sh"]