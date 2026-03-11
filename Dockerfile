FROM  ccr.ccs.tencentyun.com/webankpartners/wecube-prometheus:v1.4
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

# 定义临时目录变量（基础镜像中已重命名为 _tmp，新建的 agent_manager 也用 _tmp）
ENV PROMETHEUS_TMP=$BASE_HOME/prometheus_tmp
ENV ALERTMANAGER_TMP=$BASE_HOME/alertmanager_tmp
ENV AGENT_MANAGER_TMP=$BASE_HOME/agent_manager_tmp

# 创建临时目录（基础镜像中 prometheus_tmp 和 alertmanager_tmp 已存在，这里确保 agent_manager_tmp 存在）
# 同时创建空的原始目录供 PV 挂载
RUN mkdir -p $BASE_HOME $PROMETHEUS_TMP $PROMETHEUS_TMP/rules $PROMETHEUS_TMP/token \
    $ALERTMANAGER_TMP $MONITOR_HOME $MONITOR_HOME/conf $AGENT_MANAGER_TMP \
    $PING_EXPORTER $AGENT_MANAGER_DEPLOY $TRANS_GATEWAY $ARCHIVE_TOOL \
    $DB_DATA_EXPORTER $DAEMON_PROC $METRIC_COMPARISON_EXPORTER \
    $METRIC_COMPARISON_EXPORTER/config && \
    # 创建空的原始目录供 PV 挂载
    mkdir -p $PROMETHEUS_HOME $ALERTMANAGER_HOME $AGENT_MANAGER_HOME

COPY build/start.sh $BASE_HOME/
COPY build/stop.sh $BASE_HOME/
# 复制 prometheus 相关文件到临时目录
COPY build/conf/prometheus.yml $PROMETHEUS_TMP/
COPY build/conf/kubernetes_prometheus.tpl $PROMETHEUS_TMP/
COPY build/conf/snmp_prometheus.tpl $PROMETHEUS_TMP/
COPY build/conf/remote_write_prometheus.tpl $PROMETHEUS_TMP/
COPY build/conf/prometheus.yml $PROMETHEUS_TMP/prometheus_tpl.yml
COPY build/conf/sd_file $PROMETHEUS_TMP/sd_file
COPY build/conf/base.yml $PROMETHEUS_TMP/
# 复制 alertmanager 相关文件到临时目录
COPY build/conf/alertmanager.yml $ALERTMANAGER_TMP/
# 复制 agent_manager 相关文件到临时目录
COPY monitor-agent/agent_manager/agent_manager $AGENT_MANAGER_TMP/
COPY monitor-agent/agent_manager/exporters.tar.gz $AGENT_MANAGER_TMP/
COPY build/conf/agent_manager.json $AGENT_MANAGER_TMP/conf.json
# 复制其他文件（不受持久化卷影响）
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

# 设置执行权限（对临时目录中的文件）
RUN chmod +x $PROMETHEUS_TMP/prometheus $PROMETHEUS_TMP/promtool $ALERTMANAGER_TMP/alertmanager $AGENT_MANAGER_TMP/agent_manager $TRANS_GATEWAY/transgateway $MONITOR_HOME/monitor-server $BASE_HOME/*.sh $PING_EXPORTER/ping_exporter $ARCHIVE_TOOL/archive_mysql_tool $DB_DATA_EXPORTER/db_data_exporter $DAEMON_PROC/daemon_proc $METRIC_COMPARISON_EXPORTER/metric_comparison

# 基础镜像 v1.4 中已创建 app:apps 用户并对 /app/monitor 设置了权限（包括 prometheus_tmp 和 alertmanager_tmp）
# 应用镜像新增的文件和目录需要设置为 app:apps，以便 gosu app 可以正常访问
# 注意：不包含 PROMETHEUS_TMP 和 ALERTMANAGER_TMP，因为基础镜像中已设置
# consul 目录如果存在，也需要设置权限
RUN chown -R app:apps $MONITOR_HOME $AGENT_MANAGER_TMP $PING_EXPORTER \
    $AGENT_MANAGER_DEPLOY $TRANS_GATEWAY $ARCHIVE_TOOL $DB_DATA_EXPORTER \
    $DAEMON_PROC $METRIC_COMPARISON_EXPORTER $PROMETHEUS_HOME \
    $ALERTMANAGER_HOME $AGENT_MANAGER_HOME $BASE_HOME/*.sh && \
    # 如果 consul 目录存在，也设置权限（可能来自基础镜像）
    ([ -d "$BASE_HOME/consul" ] && chown -R app:apps $BASE_HOME/consul || true)

WORKDIR $BASE_HOME

# 容器以 root 启动，start.sh 中使用 gosu app 运行服务
ENTRYPOINT ["/bin/sh", "start.sh"]
