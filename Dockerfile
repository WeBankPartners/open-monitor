FROM  ccr.ccs.tencentyun.com/webankpartners/wecube-prometheus:v1.2
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

# Create tmp directories for prometheus, alertmanager, and agent_manager
ENV PROMETHEUS_TMP=$BASE_HOME/prometheus_tmp
ENV ALERTMANAGER_TMP=$BASE_HOME/alertmanager_tmp
ENV AGENT_MANAGER_TMP=$BASE_HOME/agent_manager_tmp

RUN mkdir -p $BASE_HOME $PROMETHEUS_TMP $PROMETHEUS_TMP/rules $PROMETHEUS_TMP/token $ALERTMANAGER_TMP $MONITOR_HOME $MONITOR_HOME/conf $AGENT_MANAGER_TMP $PING_EXPORTER $AGENT_MANAGER_DEPLOY $TRANS_GATEWAY $ARCHIVE_TOOL $DB_DATA_EXPORTER $DAEMON_PROC $METRIC_COMPARISON_EXPORTER $METRIC_COMPARISON_EXPORTER/config

COPY build/start.sh $BASE_HOME/
COPY build/stop.sh $BASE_HOME/
# Copy prometheus files to tmp directory
COPY build/conf/prometheus.yml $PROMETHEUS_TMP/
COPY build/conf/kubernetes_prometheus.tpl $PROMETHEUS_TMP/
COPY build/conf/snmp_prometheus.tpl $PROMETHEUS_TMP/
COPY build/conf/remote_write_prometheus.tpl $PROMETHEUS_TMP/
COPY build/conf/prometheus.yml $PROMETHEUS_TMP/prometheus_tpl.yml
COPY build/conf/sd_file $PROMETHEUS_TMP/sd_file
COPY build/conf/base.yml $PROMETHEUS_TMP/
# Copy alertmanager files to tmp directory
COPY build/conf/alertmanager.yml $ALERTMANAGER_TMP/
# Copy agent_manager files to tmp directory
COPY monitor-agent/agent_manager/agent_manager $AGENT_MANAGER_TMP/
COPY monitor-agent/agent_manager/exporters.tar.gz $AGENT_MANAGER_TMP/
COPY build/conf/agent_manager.json $AGENT_MANAGER_TMP/conf.json
# Copy other files (not affected by persistent volumes)
COPY monitor-server/monitor-server $MONITOR_HOME/
COPY build/conf/monitor.json $MONITOR_HOME/conf/default.json
COPY monitor-server/conf/i18n $MONITOR_HOME/conf/i18n
COPY monitor-ui/dist $MONITOR_HOME/public
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

# Copy binaries from base image to tmp directories (base image has prometheus and alertmanager binaries)
# Check if binaries exist in base image directories and copy them to tmp directories
RUN if [ -f $PROMETHEUS_HOME/prometheus ]; then cp $PROMETHEUS_HOME/prometheus $PROMETHEUS_TMP/; fi && \
    if [ -f $PROMETHEUS_HOME/promtool ]; then cp $PROMETHEUS_HOME/promtool $PROMETHEUS_TMP/; fi && \
    if [ -f $ALERTMANAGER_HOME/alertmanager ]; then cp $ALERTMANAGER_HOME/alertmanager $ALERTMANAGER_TMP/; fi

# Set execute permissions (only for files that exist)
RUN [ -f $PROMETHEUS_TMP/prometheus ] && chmod +x $PROMETHEUS_TMP/prometheus || true && \
    [ -f $PROMETHEUS_TMP/promtool ] && chmod +x $PROMETHEUS_TMP/promtool || true && \
    [ -f $ALERTMANAGER_TMP/alertmanager ] && chmod +x $ALERTMANAGER_TMP/alertmanager || true && \
    chmod +x $AGENT_MANAGER_TMP/agent_manager $TRANS_GATEWAY/transgateway $MONITOR_HOME/monitor-server $BASE_HOME/*.sh $PING_EXPORTER/ping_exporter $ARCHIVE_TOOL/archive_mysql_tool $DB_DATA_EXPORTER/db_data_exporter $DAEMON_PROC/daemon_proc $METRIC_COMPARISON_EXPORTER/metric_comparison

WORKDIR $BASE_HOME

ENTRYPOINT ["/bin/sh", "start.sh"]