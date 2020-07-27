FROM platten/alpine-oracle-jre8-docker
LABEL maintainer = "Webank CTB Team"

ENV BASE_HOME=/app/monitor
ENV PROMETHEUS_HOME=$BASE_HOME/prometheus
ENV ALERTMANAGER_HOME=$BASE_HOME/alertmanager
ENV CONSUL_HOME=$BASE_HOME/consul
ENV MONITOR_HOME=$BASE_HOME/monitor
ENV AGENT_MONITOR_HOME=$BASE_HOME/agent_manager
ENV TRANS_GATEWAY=$BASE_HOME/transgateway
ENV PING_EXPORTER=$BASE_HOME/ping_exporter
ENV ARCHIVE_TOOL=$BASE_HOME/archive_mysql_tool

ENV MONITOR_SERVER_PORT=8080
ENV MONITOR_DB_HOST=127.0.0.1
ENV MONITOR_DB_PORT=3306
ENV MONITOR_DB_USER=root
ENV MONITOR_DB_PWD=mysql

RUN mkdir -p $BASE_HOME $PROMETHEUS_HOME $PROMETHEUS_HOME/rules $ALERTMANAGER_HOME $CONSUL_HOME $MONITOR_HOME $AGENT_MONITOR_HOME $PING_EXPORTER /app/deploy $TRANS_GATEWAY $ARCHIVE_TOOL

ADD monitor/prometheus $PROMETHEUS_HOME/
ADD monitor/alertmanager $ALERTMANAGER_HOME/
ADD monitor/monitor $MONITOR_HOME/
ADD monitor/start.sh $BASE_HOME/
ADD monitor/stop.sh $BASE_HOME/
ADD monitor/agent_manager $AGENT_MONITOR_HOME/
ADD monitor/transgateway $TRANS_GATEWAY/
ADD monitor/ping_exporter $PING_EXPORTER/
ADD monitor/archive_mysql_tool $ARCHIVE_TOOL/

RUN chmod +x $PROMETHEUS_HOME/prometheus
RUN chmod +x $PROMETHEUS_HOME/promtool
RUN chmod +x $PROMETHEUS_HOME/tsdb
RUN chmod +x $ALERTMANAGER_HOME/alertmanager
RUN chmod +x $AGENT_MONITOR_HOME/agent_manager
RUN chmod +x $TRANS_GATEWAY/transgateway
RUN chmod +x $AGENT_MONITOR_HOME/exporters/*/*
RUN chmod +x $MONITOR_HOME/monitor-server
RUN chmod +x $BASE_HOME/*.sh
RUN chmod +x $PING_EXPORTER/ping_exporter
RUN chmod +x $ARCHIVE_TOOL/archive_mysql_tool

WORKDIR $BASE_HOME

ENTRYPOINT ["/bin/sh", "start.sh"]