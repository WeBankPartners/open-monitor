#!/bin/bash

# 简单日志函数，带时间前缀
log() {
  echo "$(date '+%Y-%m-%d %H:%M:%S') [start.sh] $*"
}

log "start run"

# ========== 工具函数：确保日志文件可以创建 ==========
# 先尝试以 app 用户创建，如果失败则使用 sudo 创建并设置权限
ensure_log_file() {
  local log_file=$1
  local log_dir=$(dirname "$log_file")
  
  # 确保日志目录存在且权限正确
  if [ ! -d "$log_dir" ]; then
    sudo mkdir -p "$log_dir" && sudo chown app:apps "$log_dir"
  else
    # 检查目录权限，如果不对则修复
    if [ ! -w "$log_dir" ]; then
      sudo chown app:apps "$log_dir" 2>/dev/null || true
    fi
  fi
  
  # 尝试以 app 用户创建日志文件
  if touch "$log_file" 2>/dev/null; then
    # 成功创建，确保权限正确
    sudo chown app:apps "$log_file" 2>/dev/null || true
    return 0
  else
    # 创建失败，使用 sudo 创建并设置权限
    sudo touch "$log_file" && sudo chown app:apps "$log_file"
    return $?
  fi
}

# ========== 工具函数：带日志的占位符替换 ==========
# 用法：log_replace ENV_NAME CURRENT_VALUE 文件 占位符
log_replace() {
  local env_name="$1"
  local env_value="$2"
  local target_file="$3"
  local placeholder="$4"

  if [ -z "$env_value" ]; then
    log "WARN env ${env_name} is empty, replacing ${target_file} ${placeholder} with empty string"
  else
    log "INFO replace ${target_file} ${placeholder} with '${env_value}'"
  fi
  sed -i "s~${placeholder}~${env_value}~g" "${target_file}"
}

# ========== 第零步：确保 consul 目录权限正确（如果存在） ==========
# consul 目录可能来自基础镜像,需要确保权限正确
if [ -d "/app/monitor/consul" ]; then
  sudo chown -R app:apps /app/monitor/consul 2>/dev/null || true
fi

# ========== 第一步：从 _tmp 目录复制到原目录（PV 挂载前准备） ==========
log "STEP1 copy files from *_tmp directories to target directories"

# 复制 prometheus_tmp 到 prometheus
if [ -d "/app/monitor/prometheus_tmp" ] && [ "$(ls -A /app/monitor/prometheus_tmp 2>/dev/null)" ]; then
  log "copying prometheus files from prometheus_tmp to prometheus"
  mkdir -p /app/monitor/prometheus
  mkdir -p /app/monitor/prometheus/rules
  cp -rf /app/monitor/prometheus_tmp/* /app/monitor/prometheus/
  chmod +x /app/monitor/prometheus/prometheus /app/monitor/prometheus/promtool
  # 确保 rules 目录和 base.yml 的权限正确
  if [ -f "/app/monitor/prometheus/base.yml" ]; then
    # 如果 rules/base.yml 已存在，先删除（解决 PV 挂载时文件权限问题）
    sudo rm -f /app/monitor/prometheus/rules/base.yml 2>/dev/null || rm -f /app/monitor/prometheus/rules/base.yml
    sudo cp -f /app/monitor/prometheus/base.yml /app/monitor/prometheus/rules/base.yml && sudo chown app:apps /app/monitor/prometheus/rules/base.yml
  fi
  # 确保所有文件属于 app:apps，以便 app 用户可以正常访问
  sudo chown -R app:apps /app/monitor/prometheus
  # 删除 tmp 目录（如果权限不对，使用 sudo）
  sudo rm -rf /app/monitor/prometheus_tmp 2>/dev/null || rm -rf /app/monitor/prometheus_tmp
fi

# 复制 alertmanager_tmp 到 alertmanager
if [ -d "/app/monitor/alertmanager_tmp" ] && [ "$(ls -A /app/monitor/alertmanager_tmp 2>/dev/null)" ]; then
  log "copying alertmanager files from alertmanager_tmp to alertmanager"
  mkdir -p /app/monitor/alertmanager
  cp -rf /app/monitor/alertmanager_tmp/* /app/monitor/alertmanager/
  chmod +x /app/monitor/alertmanager/alertmanager
  # 确保所有文件属于 app:apps，以便 app 用户可以正常访问
  sudo chown -R app:apps /app/monitor/alertmanager
  # 删除 tmp 目录（如果权限不对，使用 sudo）
  sudo rm -rf /app/monitor/alertmanager_tmp 2>/dev/null || rm -rf /app/monitor/alertmanager_tmp
fi

# 复制 agent_manager_tmp 到 agent_manager
if [ -d "/app/monitor/agent_manager_tmp" ] && [ "$(ls -A /app/monitor/agent_manager_tmp 2>/dev/null)" ]; then
  log "copying agent_manager files from agent_manager_tmp to agent_manager"
  mkdir -p /app/monitor/agent_manager
  cp -rf /app/monitor/agent_manager_tmp/* /app/monitor/agent_manager/
  # 确保所有文件属于 app:apps，以便 app 用户可以正常访问
  sudo chown -R app:apps /app/monitor/agent_manager
  # 删除 tmp 目录（如果权限不对，使用 sudo）
  sudo rm -rf /app/monitor/agent_manager_tmp 2>/dev/null || rm -rf /app/monitor/agent_manager_tmp
fi

# ========== 第二步：执行配置占位符替换 ==========
log "STEP2 start replacing config placeholders"
# 确保所有配置文件属于 app:apps，以便 sed 可以正常修改（文件复制时已设置，这里再次确保）
sudo chown -R app:apps alertmanager/alertmanager.yml monitor/conf/default.json archive_mysql_tool/default.json agent_manager/conf.json 2>/dev/null || true

laststr=`echo ${MONITOR_HOST_IP}|awk -F '' '{print $NF}'`
subnum='3'
if [ -n "$laststr" ]
then
  subnum=`expr $laststr % 5`
fi

log_replace "MONITOR_SERVER_PORT" "$MONITOR_SERVER_PORT" alertmanager/alertmanager.yml "{{MONITOR_SERVER_PORT}}"
log "INFO replace alertmanager MONITOR_ALERT_WAIT with '${subnum}'"
sed -i "s~{{MONITOR_ALERT_WAIT}}~${subnum}~g" alertmanager/alertmanager.yml

log_replace "MONITOR_SERVER_PORT" "$MONITOR_SERVER_PORT" monitor/conf/default.json "{{MONITOR_SERVER_PORT}}"
log_replace "MONITOR_DB_HOST" "$MONITOR_DB_HOST" monitor/conf/default.json "{{MONITOR_DB_HOST}}"
log_replace "MONITOR_DB_PORT" "$MONITOR_DB_PORT" monitor/conf/default.json "{{MONITOR_DB_PORT}}"
log_replace "MONITOR_DB_USER" "$MONITOR_DB_USER" monitor/conf/default.json "{{MONITOR_DB_USER}}"
log_replace "MONITOR_DB_PWD" "$MONITOR_DB_PWD" monitor/conf/default.json "{{MONITOR_DB_PWD}}"
log_replace "MONITOR_DB_SCHEMA" "$MONITOR_DB_SCHEMA" monitor/conf/default.json "{{MONITOR_DB_SCHEMA}}"
log_replace "MONITOR_ARCHIVE_ENABLE" "$MONITOR_ARCHIVE_ENABLE" monitor/conf/default.json "{{MONITOR_ARCHIVE_ENABLE}}"
log_replace "MONITOR_READ_ARCHIVE_ENABLE" "$MONITOR_READ_ARCHIVE_ENABLE" monitor/conf/default.json "{{MONITOR_READ_ARCHIVE_ENABLE}}"
log_replace "MONITOR_ARCHIVE_MYSQL_HOST" "$MONITOR_ARCHIVE_MYSQL_HOST" monitor/conf/default.json "{{MONITOR_ARCHIVE_MYSQL_HOST}}"
log_replace "MONITOR_ARCHIVE_MYSQL_PORT" "$MONITOR_ARCHIVE_MYSQL_PORT" monitor/conf/default.json "{{MONITOR_ARCHIVE_MYSQL_PORT}}"
log_replace "MONITOR_ARCHIVE_MYSQL_USER" "$MONITOR_ARCHIVE_MYSQL_USER" monitor/conf/default.json "{{MONITOR_ARCHIVE_MYSQL_USER}}"
log_replace "MONITOR_ARCHIVE_MYSQL_PWD" "$MONITOR_ARCHIVE_MYSQL_PWD" monitor/conf/default.json "{{MONITOR_ARCHIVE_MYSQL_PWD}}"

# 为数字类型的环境变量设置默认值，避免空值导致 JSON 格式错误
# 这些字段在配置文件中没有引号，是数字类型，如果为空会导致 JSON 解析失败
if [ -z "$MONITOR_ARCHIVE_READ_MAX_OPEN" ]; then
  MONITOR_ARCHIVE_READ_MAX_OPEN="20"
  log "WARN MONITOR_ARCHIVE_READ_MAX_OPEN is empty, using default value: 20"
fi
if [ -z "$MONITOR_ARCHIVE_READ_MAX_IDLE" ]; then
  MONITOR_ARCHIVE_READ_MAX_IDLE="10"
  log "WARN MONITOR_ARCHIVE_READ_MAX_IDLE is empty, using default value: 10"
fi
if [ -z "$MONITOR_ARCHIVE_READ_TIMEOUT" ]; then
  MONITOR_ARCHIVE_READ_TIMEOUT="60"
  log "WARN MONITOR_ARCHIVE_READ_TIMEOUT is empty, using default value: 60"
fi

log_replace "MONITOR_ARCHIVE_READ_MAX_OPEN" "$MONITOR_ARCHIVE_READ_MAX_OPEN" monitor/conf/default.json "{{MONITOR_ARCHIVE_READ_MAX_OPEN}}"
log_replace "MONITOR_ARCHIVE_READ_MAX_IDLE" "$MONITOR_ARCHIVE_READ_MAX_IDLE" monitor/conf/default.json "{{MONITOR_ARCHIVE_READ_MAX_IDLE}}"
log_replace "MONITOR_ARCHIVE_READ_TIMEOUT" "$MONITOR_ARCHIVE_READ_TIMEOUT" monitor/conf/default.json "{{MONITOR_ARCHIVE_READ_TIMEOUT}}"
log_replace "MONITOR_ALARM_ALIVE_MAX_DAY" "$MONITOR_ALARM_ALIVE_MAX_DAY" monitor/conf/default.json "{{MONITOR_ALARM_ALIVE_MAX_DAY}}"
log_replace "MONITOR_LOG_LEVEL" "$MONITOR_LOG_LEVEL" monitor/conf/default.json "{{MONITOR_LOG_LEVEL}}"

log_replace "MONITOR_DB_HOST" "$MONITOR_DB_HOST" archive_mysql_tool/default.json "{{MONITOR_DB_HOST}}"
log_replace "MONITOR_DB_PORT" "$MONITOR_DB_PORT" archive_mysql_tool/default.json "{{MONITOR_DB_PORT}}"
log_replace "MONITOR_DB_USER" "$MONITOR_DB_USER" archive_mysql_tool/default.json "{{MONITOR_DB_USER}}"
log_replace "MONITOR_DB_PWD" "$MONITOR_DB_PWD" archive_mysql_tool/default.json "{{MONITOR_DB_PWD}}"
log_replace "MONITOR_ARCHIVE_ENABLE" "$MONITOR_ARCHIVE_ENABLE" archive_mysql_tool/default.json "{{MONITOR_ARCHIVE_ENABLE}}"
log_replace "MONITOR_ARCHIVE_MYSQL_HOST" "$MONITOR_ARCHIVE_MYSQL_HOST" archive_mysql_tool/default.json "{{MONITOR_ARCHIVE_MYSQL_HOST}}"
log_replace "MONITOR_ARCHIVE_MYSQL_PORT" "$MONITOR_ARCHIVE_MYSQL_PORT" archive_mysql_tool/default.json "{{MONITOR_ARCHIVE_MYSQL_PORT}}"
log_replace "MONITOR_ARCHIVE_MYSQL_USER" "$MONITOR_ARCHIVE_MYSQL_USER" archive_mysql_tool/default.json "{{MONITOR_ARCHIVE_MYSQL_USER}}"
log_replace "MONITOR_ARCHIVE_MYSQL_PWD" "$MONITOR_ARCHIVE_MYSQL_PWD" archive_mysql_tool/default.json "{{MONITOR_ARCHIVE_MYSQL_PWD}}"
log_replace "MONITOR_ARCHIVE_MYSQL_MAX_OPEN" "$MONITOR_ARCHIVE_MYSQL_MAX_OPEN" archive_mysql_tool/default.json "{{MONITOR_ARCHIVE_MYSQL_MAX_OPEN}}"
log_replace "MONITOR_ARCHIVE_MYSQL_MAX_IDLE" "$MONITOR_ARCHIVE_MYSQL_MAX_IDLE" archive_mysql_tool/default.json "{{MONITOR_ARCHIVE_MYSQL_MAX_IDLE}}"
log_replace "MONITOR_ALARM_MAIL_ENABLE" "$MONITOR_ALARM_MAIL_ENABLE" monitor/conf/default.json "{{MONITOR_ALARM_MAIL_ENABLE}}"
log_replace "MONITOR_ALARM_CALLBACK_LEVEL_MIN" "$MONITOR_ALARM_CALLBACK_LEVEL_MIN" monitor/conf/default.json "{{MONITOR_ALARM_CALLBACK_LEVEL_MIN}}"
log_replace "MONITOR_ARCHIVE_UNIT_SPEED" "$MONITOR_ARCHIVE_UNIT_SPEED" archive_mysql_tool/default.json "{{MONITOR_ARCHIVE_UNIT_SPEED}}"
log_replace "MONITOR_ARCHIVE_CONCURRENT_NUM" "$MONITOR_ARCHIVE_CONCURRENT_NUM" archive_mysql_tool/default.json "{{MONITOR_ARCHIVE_CONCURRENT_NUM}}"
log_replace "MONITOR_ARCHIVE_MAX_HTTP_OPEN" "$MONITOR_ARCHIVE_MAX_HTTP_OPEN" archive_mysql_tool/default.json "{{MONITOR_ARCHIVE_MAX_HTTP_OPEN}}"
log_replace "MONITOR_AGENT_MANAGER_REMOTE_MODE" "$MONITOR_AGENT_MANAGER_REMOTE_MODE" agent_manager/conf.json "{{MONITOR_AGENT_MANAGER_REMOTE_MODE}}"
log_replace "ENCRYPT_SEED" "$ENCRYPT_SEED" monitor/conf/default.json "{{ENCRYPT_SEED}}"
log_replace "MONITOR_MENU_API_ENABLE" "$MONITOR_MENU_API_ENABLE" monitor/conf/default.json "{{MONITOR_MENU_API_ENABLE}}"

if [ -n "$GATEWAY_URL" ]
then
  log_replace "GATEWAY_URL" "$GATEWAY_URL" monitor/conf/default.json "{{CORE_ADDR}}"
else
  log_replace "CORE_ADDR" "$CORE_ADDR" monitor/conf/default.json "{{CORE_ADDR}}"
fi

log_replace "MONITOR_HOST_IP" "$MONITOR_HOST_IP" monitor/conf/default.json "{{MONITOR_SERVER_IP}}"

if [ -n "$MONITOR_SESSION_ENABLE" ]
then
  log "INFO MONITOR_SESSION_ENABLE is set, enable session=true"
  sed -i "s~{{MONITOR_SESSION_ENABLE}}~true~g" monitor/conf/default.json
else
  log "INFO MONITOR_SESSION_ENABLE is empty, use session=false"
  sed -i "s~{{MONITOR_SESSION_ENABLE}}~false~g" monitor/conf/default.json
fi

if [ -n "$PLUGIN_MODE" ]
then
  log "INFO PLUGIN_MODE is set, is_plugin_mode=yes"
  sed -i "s~{{PLUGIN_MODE}}~yes~g" monitor/conf/default.json
else
  log "INFO PLUGIN_MODE is empty, is_plugin_mode=no"
  sed -i "s~{{PLUGIN_MODE}}~no~g" monitor/conf/default.json
fi

log "STEP2 config placeholders replaced"

if [ -n "$MONITOR_LOCAL_DNS_MAP" ]
then
  log "STEP2.1 MONITOR_LOCAL_DNS_MAP=${MONITOR_LOCAL_DNS_MAP}, start append to /etc/hosts"
  dns_map=${MONITOR_LOCAL_DNS_MAP}
  set ${dns_map//,/ }
  for v in "$@"
  do
    sudo sh -c "echo '${v//=/ }' >> /etc/hosts"
  done
else
  log "STEP2.1 MONITOR_LOCAL_DNS_MAP is empty, skip /etc/hosts update"
fi

log "STEP3 calculate prometheus archive_day"
archive_day="30d"
if [ "$MONITOR_PROMETHEUS_ARCHIVE_DAY" -gt 0 ] 2>/dev/null;
then
  archive_day="${MONITOR_PROMETHEUS_ARCHIVE_DAY}d"
fi

log "STEP4 only start monitor-server, other components are disabled for debugging"
## ===== 这里只保留 monitor-server 的启动，其它服务暂时不启动，用于排查问题 =====
#cd agent_manager
#sudo mkdir -p logs && sudo chown app:apps logs
#tar zxf exporters.tar.gz
## 确保解压出来的文件属于 app:apps
#sudo chown -R app:apps . 2>/dev/null || true
#ensure_log_file logs/app.log
#nohup ./agent_manager > logs/app.log 2>&1 &
#cd ../daemon_proc
#sudo mkdir -p logs && sudo chown app:apps logs
#ensure_log_file logs/app.log
#nohup ./daemon_proc > logs/app.log 2>&1 &
#cd ../alertmanager
#sudo mkdir -p logs && sudo chown app:apps logs
#ensure_log_file logs/alertmanager.log
#nohup ./alertmanager --config.file=alertmanager.yml --web.listen-address=":9093"  --cluster.listen-address=":9094" > logs/alertmanager.log 2>&1 &
#cd ../prometheus/
#sudo mkdir -p rules logs && sudo chown app:apps logs rules
## 如果 rules/base.yml 已存在，先删除（解决 PV 挂载时文件权限问题）
#sudo rm -f rules/base.yml 2>/dev/null || rm -f rules/base.yml
#if [ -f "base.yml" ]; then
#  sudo cp -f base.yml rules/ && sudo chown app:apps rules/base.yml
#fi
#cd /app/monitor/prometheus && ensure_log_file logs/prometheus.log && nohup ./prometheus --config.file=prometheus.yml --web.enable-lifecycle --storage.tsdb.retention.time=${archive_day} > logs/prometheus.log 2>&1 &
#cd ../ping_exporter/
#sudo chown app:apps . 2>/dev/null || true
#sudo mkdir -p logs && sudo chown app:apps logs
#ensure_log_file logs/app.log
#nohup ./ping_exporter > logs/app.log 2>&1 &
#cd ../transgateway/
#sudo chown app:apps . 2>/dev/null || true
#sudo mkdir -p logs data && sudo chown app:apps logs data
#ensure_log_file logs/app.log
#nohup ./transgateway -d data -m http://127.0.0.1:8080 > logs/app.log 2>&1 &
#cd ../archive_mysql_tool
#sudo chown app:apps . 2>/dev/null || true
#sudo mkdir -p logs && sudo chown app:apps logs
#ensure_log_file logs/app.log
#nohup ./archive_mysql_tool > logs/app.log 2>&1 &
#cd ../db_data_exporter
#sudo chown app:apps . 2>/dev/null || true
#sudo mkdir -p logs && sudo chown app:apps logs
#ensure_log_file logs/app.log
#nohup ./db_data_exporter > logs/app.log 2>&1 &
#cd ../metric_comparison_exporter
#sudo chown app:apps . 2>/dev/null || true
#sudo mkdir -p logs && sudo chown app:apps logs
# 由于当前工作目录是 /app/monitor，这里直接进入 monitor 目录
cd monitor/
sudo chown app:apps . 2>/dev/null || true
sudo mkdir -p logs && sudo chown app:apps logs
ensure_log_file logs/app.log
sleep 2
Exit_actions (){
  # 这里只启动了 monitor-server，先简单等待其退出
  wait $!
}
trap Exit_actions INT TERM EXIT
log "STEP5 start monitor-server"
GODEBUG=netdns=go nohup ./monitor-server > logs/app.log 2>&1 &
log "STEP5 monitor-server started, waiting for process to exit"
wait $!
