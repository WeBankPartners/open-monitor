# 多阶段构建：第一阶段提取基础镜像中的二进制文件
FROM ccr.ccs.tencentyun.com/webankpartners/wecube-prometheus:v1.2 AS extractor

ENV BASE_HOME=/app/monitor
ENV PROMETHEUS_HOME=$BASE_HOME/prometheus
ENV ALERTMANAGER_HOME=$BASE_HOME/alertmanager
ENV AGENT_MANAGER_HOME=$BASE_HOME/agent_manager

# 提取基础镜像中的二进制文件到临时位置
RUN mkdir -p /tmp/extract/prometheus /tmp/extract/alertmanager /tmp/extract/agent_manager && \
    if [ -f $PROMETHEUS_HOME/prometheus ]; then cp $PROMETHEUS_HOME/prometheus /tmp/extract/prometheus/; fi && \
    if [ -f $PROMETHEUS_HOME/promtool ]; then cp $PROMETHEUS_HOME/promtool /tmp/extract/prometheus/; fi && \
    if [ -f $ALERTMANAGER_HOME/alertmanager ]; then cp $ALERTMANAGER_HOME/alertmanager /tmp/extract/alertmanager/; fi && \
    if [ -d "$PROMETHEUS_HOME" ] && [ "$(ls -A $PROMETHEUS_HOME 2>/dev/null)" ]; then \
        cp -rf $PROMETHEUS_HOME/* /tmp/extract/prometheus/ 2>/dev/null || true; \
    fi && \
    if [ -d "$ALERTMANAGER_HOME" ] && [ "$(ls -A $ALERTMANAGER_HOME 2>/dev/null)" ]; then \
        cp -rf $ALERTMANAGER_HOME/* /tmp/extract/alertmanager/ 2>/dev/null || true; \
    fi && \
    if [ -d "$AGENT_MANAGER_HOME" ] && [ "$(ls -A $AGENT_MANAGER_HOME 2>/dev/null)" ]; then \
        cp -rf $AGENT_MANAGER_HOME/* /tmp/extract/agent_manager/ 2>/dev/null || true; \
    fi && \
    # 确保目录不为空，避免 COPY --from 失败
    touch /tmp/extract/prometheus/.keep /tmp/extract/alertmanager/.keep /tmp/extract/agent_manager/.keep

# 第二阶段：最终镜像，只包含需要的文件，不包含基础镜像的冗余层
FROM ccr.ccs.tencentyun.com/webankpartners/wecube-prometheus:v1.2
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

# 为 prometheus、alertmanager 和 agent_manager 创建临时目录
# 临时目录用于存储所有文件（包括基础镜像中的二进制文件），避免与 PV 挂载冲突
ENV PROMETHEUS_TMP=$BASE_HOME/prometheus_tmp
ENV ALERTMANAGER_TMP=$BASE_HOME/alertmanager_tmp
ENV AGENT_MANAGER_TMP=$BASE_HOME/agent_manager_tmp

RUN mkdir -p $BASE_HOME $PROMETHEUS_TMP $PROMETHEUS_TMP/rules $PROMETHEUS_TMP/token $ALERTMANAGER_TMP $MONITOR_HOME $MONITOR_HOME/conf $AGENT_MANAGER_TMP $PING_EXPORTER $AGENT_MANAGER_DEPLOY $TRANS_GATEWAY $ARCHIVE_TOOL $DB_DATA_EXPORTER $DAEMON_PROC $METRIC_COMPARISON_EXPORTER $METRIC_COMPARISON_EXPORTER/config

COPY build/start.sh $BASE_HOME/
COPY build/stop.sh $BASE_HOME/
# 复制 prometheus 文件到临时目录
COPY build/conf/prometheus.yml $PROMETHEUS_TMP/
COPY build/conf/kubernetes_prometheus.tpl $PROMETHEUS_TMP/
COPY build/conf/snmp_prometheus.tpl $PROMETHEUS_TMP/
COPY build/conf/remote_write_prometheus.tpl $PROMETHEUS_TMP/
COPY build/conf/prometheus.yml $PROMETHEUS_TMP/prometheus_tpl.yml
COPY build/conf/sd_file $PROMETHEUS_TMP/sd_file
COPY build/conf/base.yml $PROMETHEUS_TMP/
# 复制 alertmanager 文件到临时目录
COPY build/conf/alertmanager.yml $ALERTMANAGER_TMP/
# 复制 agent_manager 文件到临时目录
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

# 从第一阶段复制提取的二进制文件到临时目录（只复制一次，不包含基础镜像的冗余）
# 先复制到临时位置，然后移动到目标目录（处理源可能不存在的情况）
COPY --from=extractor /tmp/extract /tmp/extract_stage2
RUN if [ -d "/tmp/extract_stage2/prometheus" ] && [ "$(ls -A /tmp/extract_stage2/prometheus 2>/dev/null)" ]; then \
        cp -rf /tmp/extract_stage2/prometheus/* $PROMETHEUS_TMP/ 2>/dev/null || true; \
    fi && \
    if [ -d "/tmp/extract_stage2/alertmanager" ] && [ "$(ls -A /tmp/extract_stage2/alertmanager 2>/dev/null)" ]; then \
        cp -rf /tmp/extract_stage2/alertmanager/* $ALERTMANAGER_TMP/ 2>/dev/null || true; \
    fi && \
    if [ -d "/tmp/extract_stage2/agent_manager" ] && [ "$(ls -A /tmp/extract_stage2/agent_manager 2>/dev/null)" ]; then \
        cp -rf /tmp/extract_stage2/agent_manager/* $AGENT_MANAGER_TMP/ 2>/dev/null || true; \
    fi && \
    rm -rf /tmp/extract_stage2

# 删除基础镜像中的原始目录，避免重复（所有内容已在临时目录中）
RUN rm -rf $PROMETHEUS_HOME $ALERTMANAGER_HOME $AGENT_MANAGER_HOME && \
    mkdir -p $PROMETHEUS_HOME $ALERTMANAGER_HOME $AGENT_MANAGER_HOME

# 设置执行权限
RUN chmod +x $PROMETHEUS_TMP/prometheus $PROMETHEUS_TMP/promtool $ALERTMANAGER_TMP/alertmanager 2>/dev/null || true && \
    chmod +x $AGENT_MANAGER_TMP/agent_manager $TRANS_GATEWAY/transgateway $MONITOR_HOME/monitor-server $BASE_HOME/*.sh $PING_EXPORTER/ping_exporter $ARCHIVE_TOOL/archive_mysql_tool $DB_DATA_EXPORTER/db_data_exporter $DAEMON_PROC/daemon_proc $METRIC_COMPARISON_EXPORTER/metric_comparison

# 安全基线：禁止容器内以 root 运行进程，创建专用用户并切换
WORKDIR $BASE_HOME
RUN addgroup -S apps -g 6000 && adduser -S app -u 6001 -G apps
RUN chown -R app:apps $BASE_HOME $AGENT_MANAGER_DEPLOY && chmod -R 755 $BASE_HOME $AGENT_MANAGER_DEPLOY
USER app
ENTRYPOINT ["/bin/sh", "start.sh"]
