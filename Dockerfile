FROM  ccr.ccs.tencentyun.com/webankpartners/wecube-prometheus:v1.0
LABEL maintainer = "Webank CTB Team"

ENV BASE_HOME=/app/monitor
ENV PROMETHEUS_HOME=$BASE_HOME/prometheus
ENV ALERTMANAGER_HOME=$BASE_HOME/alertmanager
ENV MONITOR_HOME=$BASE_HOME/monitor
ENV AGENT_MANAGER_HOME=$BASE_HOME/agent_manager
ENV AGENT_MANAGER_DEPLOY=/app/deploy
ENV TRANS_GATEWAY=$BASE_HOME/transgateway
ENV PING_EXPORTER=$BASE_HOME/ping_exporter
ENV ARCHIVE_TOOL=$BASE_HOME/archive_mysql_tool

RUN mkdir -p $BASE_HOME $PROMETHEUS_HOME $PROMETHEUS_HOME/rules $ALERTMANAGER_HOME $MONITOR_HOME $MONITOR_HOME/conf $AGENT_MANAGER_HOME $PING_EXPORTER $AGENT_MANAGER_DEPLOY $TRANS_GATEWAY $ARCHIVE_TOOL

COPY build/start.sh $BASE_HOME/
COPY build/stop.sh $BASE_HOME/
COPY build/conf/prometheus.yml $PROMETHEUS_HOME/
COPY build/conf/sd_file $PROMETHEUS_HOME/sd_file
COPY build/conf/alertmanager.yml $ALERTMANAGER_HOME/
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

RUN chmod +x $PROMETHEUS_HOME/prometheus
RUN chmod +x $PROMETHEUS_HOME/promtool
RUN chmod +x $PROMETHEUS_HOME/tsdb
RUN chmod +x $ALERTMANAGER_HOME/alertmanager
RUN chmod +x $AGENT_MANAGER_HOME/agent_manager
RUN chmod +x $TRANS_GATEWAY/transgateway
RUN chmod +x $MONITOR_HOME/monitor-server
RUN chmod +x $BASE_HOME/*.sh
RUN chmod +x $PING_EXPORTER/ping_exporter
RUN chmod +x $ARCHIVE_TOOL/archive_mysql_tool

WORKDIR $BASE_HOME

ENTRYPOINT ["/bin/sh", "start.sh"]