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

# 使用 --chown=app:apps 在 COPY 时设置文件所有者，避免后续 chown -R 创建大层
# 基础镜像中已对整个 /app/monitor 设置了权限，新文件也需要属于 app:apps
COPY --chown=app:apps build/start.sh $BASE_HOME/
COPY --chown=app:apps build/stop.sh $BASE_HOME/
# 复制 prometheus 相关文件到临时目录
COPY --chown=app:apps build/conf/prometheus.yml $PROMETHEUS_TMP/
COPY --chown=app:apps build/conf/kubernetes_prometheus.tpl $PROMETHEUS_TMP/
COPY --chown=app:apps build/conf/snmp_prometheus.tpl $PROMETHEUS_TMP/
COPY --chown=app:apps build/conf/remote_write_prometheus.tpl $PROMETHEUS_TMP/
COPY --chown=app:apps build/conf/prometheus.yml $PROMETHEUS_TMP/prometheus_tpl.yml
COPY --chown=app:apps build/conf/sd_file $PROMETHEUS_TMP/sd_file
COPY --chown=app:apps build/conf/base.yml $PROMETHEUS_TMP/
# 复制 alertmanager 相关文件到临时目录
COPY --chown=app:apps build/conf/alertmanager.yml $ALERTMANAGER_TMP/
# 复制 agent_manager 相关文件到临时目录
COPY --chown=app:apps monitor-agent/agent_manager/agent_manager $AGENT_MANAGER_TMP/
COPY --chown=app:apps monitor-agent/agent_manager/exporters.tar.gz $AGENT_MANAGER_TMP/
COPY --chown=app:apps build/conf/agent_manager.json $AGENT_MANAGER_TMP/conf.json
# 复制其他文件（不受持久化卷影响）
COPY --chown=app:apps monitor-server/monitor-server $MONITOR_HOME/
COPY --chown=app:apps build/conf/monitor.json $MONITOR_HOME/conf/default.json
COPY --chown=app:apps monitor-server/conf/i18n $MONITOR_HOME/conf/i18n
COPY --chown=app:apps monitor-ui/dist $MONITOR_HOME/public
COPY --chown=app:apps monitor-agent/transgateway/transgateway $TRANS_GATEWAY/
COPY --chown=app:apps monitor-agent/ping_exporter/ping_exporter $PING_EXPORTER/
COPY --chown=app:apps build/conf/ping_exporter.json $PING_EXPORTER/cfg.json
COPY --chown=app:apps monitor-agent/archive_mysql_tool/archive_mysql_tool $ARCHIVE_TOOL/
COPY --chown=app:apps build/conf/archive_mysql_tool.json $ARCHIVE_TOOL/default.json
COPY --chown=app:apps build/conf/core-site.xml $ARCHIVE_TOOL/conf/core-site.xml
COPY --chown=app:apps build/conf/hdfs-site.xml $ARCHIVE_TOOL/conf/hdfs-site.xml
COPY --chown=app:apps build/conf/krb5.conf $ARCHIVE_TOOL/conf/krb5.conf
COPY --chown=app:apps monitor-agent/db_data_exporter/db_data_exporter $DB_DATA_EXPORTER/
COPY --chown=app:apps monitor-agent/daemon_proc/daemon_proc $DAEMON_PROC/
COPY --chown=app:apps monitor-agent/daemon_proc/config.json $DAEMON_PROC/
COPY --chown=app:apps monitor-agent/metric_comparison_exporter/metric_comparison $METRIC_COMPARISON_EXPORTER/
COPY --chown=app:apps monitor-server/conf/menu-api-map.json $MONITOR_HOME/conf/

# 设置执行权限（对临时目录中的文件）
RUN chmod +x $PROMETHEUS_TMP/prometheus $PROMETHEUS_TMP/promtool $ALERTMANAGER_TMP/alertmanager $AGENT_MANAGER_TMP/agent_manager $TRANS_GATEWAY/transgateway $MONITOR_HOME/monitor-server $BASE_HOME/*.sh $PING_EXPORTER/ping_exporter $ARCHIVE_TOOL/archive_mysql_tool $DB_DATA_EXPORTER/db_data_exporter $DAEMON_PROC/daemon_proc $METRIC_COMPARISON_EXPORTER/metric_comparison

# 对应用镜像新增的所有目录递归设置权限为 app:apps
# 基础镜像中的 prometheus_tmp 和 alertmanager_tmp 已属于 app:apps，无需再次设置
# 只对应用镜像新增的目录执行 chown -R，避免影响基础镜像已有目录，减少镜像大小影响
RUN chown -R app:apps $MONITOR_HOME $AGENT_MANAGER_TMP $PING_EXPORTER \
    $AGENT_MANAGER_DEPLOY $TRANS_GATEWAY $ARCHIVE_TOOL $DB_DATA_EXPORTER \
    $DAEMON_PROC $METRIC_COMPARISON_EXPORTER $PROMETHEUS_HOME \
    $ALERTMANAGER_HOME $AGENT_MANAGER_HOME

# 安全基线：禁止容器内以 root 运行进程，切换为非 root 用户运行
# 基础镜像 v1.4 中已创建用户并对整个 /app/monitor 设置了权限
# 应用镜像中所有文件已通过 COPY --chown=app:apps 设置了权限，并在上面递归设置了目录权限
WORKDIR $BASE_HOME
USER app
ENTRYPOINT ["/bin/sh", "start.sh"]