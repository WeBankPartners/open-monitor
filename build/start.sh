#!/bin/bash

echo "start run"

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

# ========== 第零步：确保 consul 目录权限正确（如果存在） ==========
# consul 目录可能来自基础镜像,需要确保权限正确
if [ -d "/app/monitor/consul" ]; then
  sudo chown -R app:apps /app/monitor/consul 2>/dev/null || true
fi

# ========== 第一步：从 _tmp 目录复制到原目录（PV 挂载前准备） ==========
echo "Copying files from tmp directories to target directories..."

# 复制 prometheus_tmp 到 prometheus
if [ -d "/app/monitor/prometheus_tmp" ] && [ "$(ls -A /app/monitor/prometheus_tmp 2>/dev/null)" ]; then
  echo "Copying prometheus files from prometheus_tmp to prometheus..."
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
  echo "Copying alertmanager files from alertmanager_tmp to alertmanager..."
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
  echo "Copying agent_manager files from agent_manager_tmp to agent_manager..."
  mkdir -p /app/monitor/agent_manager
  cp -rf /app/monitor/agent_manager_tmp/* /app/monitor/agent_manager/
  # 确保所有文件属于 app:apps，以便 app 用户可以正常访问
  sudo chown -R app:apps /app/monitor/agent_manager
  # 删除 tmp 目录（如果权限不对，使用 sudo）
  sudo rm -rf /app/monitor/agent_manager_tmp 2>/dev/null || rm -rf /app/monitor/agent_manager_tmp
fi

# ========== 第二步：执行原有的 sed 替换逻辑 ==========
# 确保所有配置文件属于 app:apps，以便 sed 可以正常修改（文件复制时已设置，这里再次确保）
sudo chown -R app:apps alertmanager/alertmanager.yml monitor/conf/default.json archive_mysql_tool/default.json agent_manager/conf.json 2>/dev/null || true

laststr=`echo ${MONITOR_HOST_IP}|awk -F '' '{print $NF}'`
subnum='3'
if [ $laststr ]
then
subnum=`expr $laststr % 5`
fi
sed -i "s~{{MONITOR_SERVER_PORT}}~$MONITOR_SERVER_PORT~g" alertmanager/alertmanager.yml
sed -i "s~{{MONITOR_ALERT_WAIT}}~${subnum}~g" alertmanager/alertmanager.yml
sed -i "s~{{MONITOR_SERVER_PORT}}~$MONITOR_SERVER_PORT~g" monitor/conf/default.json
sed -i "s~{{MONITOR_DB_HOST}}~$MONITOR_DB_HOST~g" monitor/conf/default.json
sed -i "s~{{MONITOR_DB_PORT}}~$MONITOR_DB_PORT~g" monitor/conf/default.json
sed -i "s~{{MONITOR_DB_USER}}~$MONITOR_DB_USER~g" monitor/conf/default.json
sed -i "s~{{MONITOR_DB_PWD}}~$MONITOR_DB_PWD~g" monitor/conf/default.json
sed -i "s~{{MONITOR_DB_SCHEMA}}~$MONITOR_DB_SCHEMA~g" monitor/conf/default.json
sed -i "s~{{MONITOR_ARCHIVE_ENABLE}}~$MONITOR_ARCHIVE_ENABLE~g" monitor/conf/default.json
sed -i "s~{{MONITOR_READ_ARCHIVE_ENABLE}}~$MONITOR_READ_ARCHIVE_ENABLE~g" monitor/conf/default.json
sed -i "s~{{MONITOR_ARCHIVE_MYSQL_HOST}}~$MONITOR_ARCHIVE_MYSQL_HOST~g" monitor/conf/default.json
sed -i "s~{{MONITOR_ARCHIVE_MYSQL_PORT}}~$MONITOR_ARCHIVE_MYSQL_PORT~g" monitor/conf/default.json
sed -i "s~{{MONITOR_ARCHIVE_MYSQL_USER}}~$MONITOR_ARCHIVE_MYSQL_USER~g" monitor/conf/default.json
sed -i "s~{{MONITOR_ARCHIVE_MYSQL_PWD}}~$MONITOR_ARCHIVE_MYSQL_PWD~g" monitor/conf/default.json

# 为数字类型的环境变量设置默认值，避免空值导致 JSON 格式错误
# 这些字段在配置文件中没有引号，是数字类型，如果为空会导致 JSON 解析失败
if [ -z "$MONITOR_ARCHIVE_READ_MAX_OPEN" ]; then
  MONITOR_ARCHIVE_READ_MAX_OPEN="20"
  echo "WARNING: MONITOR_ARCHIVE_READ_MAX_OPEN is empty, using default value: 20"
fi
if [ -z "$MONITOR_ARCHIVE_READ_MAX_IDLE" ]; then
  MONITOR_ARCHIVE_READ_MAX_IDLE="10"
  echo "WARNING: MONITOR_ARCHIVE_READ_MAX_IDLE is empty, using default value: 10"
fi
if [ -z "$MONITOR_ARCHIVE_READ_TIMEOUT" ]; then
  MONITOR_ARCHIVE_READ_TIMEOUT="60"
  echo "WARNING: MONITOR_ARCHIVE_READ_TIMEOUT is empty, using default value: 60"
fi
sed -i "s~{{MONITOR_ARCHIVE_READ_MAX_OPEN}}~$MONITOR_ARCHIVE_READ_MAX_OPEN~g" monitor/conf/default.json
sed -i "s~{{MONITOR_ARCHIVE_READ_MAX_IDLE}}~$MONITOR_ARCHIVE_READ_MAX_IDLE~g" monitor/conf/default.json
sed -i "s~{{MONITOR_ARCHIVE_READ_TIMEOUT}}~$MONITOR_ARCHIVE_READ_TIMEOUT~g" monitor/conf/default.json
sed -i "s~{{MONITOR_ALARM_ALIVE_MAX_DAY}}~$MONITOR_ALARM_ALIVE_MAX_DAY~g" monitor/conf/default.json
sed -i "s~{{MONITOR_LOG_LEVEL}}~$MONITOR_LOG_LEVEL~g" monitor/conf/default.json
sed -i "s~{{MONITOR_DB_HOST}}~$MONITOR_DB_HOST~g" archive_mysql_tool/default.json
sed -i "s~{{MONITOR_DB_PORT}}~$MONITOR_DB_PORT~g" archive_mysql_tool/default.json
sed -i "s~{{MONITOR_DB_USER}}~$MONITOR_DB_USER~g" archive_mysql_tool/default.json
sed -i "s~{{MONITOR_DB_PWD}}~$MONITOR_DB_PWD~g" archive_mysql_tool/default.json
sed -i "s~{{MONITOR_ARCHIVE_ENABLE}}~$MONITOR_ARCHIVE_ENABLE~g" archive_mysql_tool/default.json
sed -i "s~{{MONITOR_ARCHIVE_MYSQL_HOST}}~$MONITOR_ARCHIVE_MYSQL_HOST~g" archive_mysql_tool/default.json
sed -i "s~{{MONITOR_ARCHIVE_MYSQL_PORT}}~$MONITOR_ARCHIVE_MYSQL_PORT~g" archive_mysql_tool/default.json
sed -i "s~{{MONITOR_ARCHIVE_MYSQL_USER}}~$MONITOR_ARCHIVE_MYSQL_USER~g" archive_mysql_tool/default.json
sed -i "s~{{MONITOR_ARCHIVE_MYSQL_PWD}}~$MONITOR_ARCHIVE_MYSQL_PWD~g" archive_mysql_tool/default.json
sed -i "s~{{MONITOR_ARCHIVE_MYSQL_MAX_OPEN}}~$MONITOR_ARCHIVE_MYSQL_MAX_OPEN~g" archive_mysql_tool/default.json
sed -i "s~{{MONITOR_ARCHIVE_MYSQL_MAX_IDLE}}~$MONITOR_ARCHIVE_MYSQL_MAX_IDLE~g" archive_mysql_tool/default.json
sed -i "s~{{MONITOR_ALARM_MAIL_ENABLE}}~$MONITOR_ALARM_MAIL_ENABLE~g" monitor/conf/default.json
sed -i "s~{{MONITOR_ALARM_CALLBACK_LEVEL_MIN}}~$MONITOR_ALARM_CALLBACK_LEVEL_MIN~g" monitor/conf/default.json
sed -i "s~{{MONITOR_ALARM_CALLBACK_LEVEL_MIN}}~$MONITOR_ALARM_CALLBACK_LEVEL_MIN~g" monitor/conf/default.json
sed -i "s~{{MONITOR_ARCHIVE_UNIT_SPEED}}~$MONITOR_ARCHIVE_UNIT_SPEED~g" archive_mysql_tool/default.json
sed -i "s~{{MONITOR_ARCHIVE_CONCURRENT_NUM}}~$MONITOR_ARCHIVE_CONCURRENT_NUM~g" archive_mysql_tool/default.json
sed -i "s~{{MONITOR_ARCHIVE_MAX_HTTP_OPEN}}~$MONITOR_ARCHIVE_MAX_HTTP_OPEN~g" archive_mysql_tool/default.json
sed -i "s~{{MONITOR_AGENT_MANAGER_REMOTE_MODE}}~$MONITOR_AGENT_MANAGER_REMOTE_MODE~g" agent_manager/conf.json
sed -i "s~{{ENCRYPT_SEED}}~$ENCRYPT_SEED~g" monitor/conf/default.json
sed -i "s~{{MONITOR_MENU_API_ENABLE}}~$MONITOR_MENU_API_ENABLE~g" monitor/conf/default.json


if [ $GATEWAY_URL ]
then
sed -i "s~{{CORE_ADDR}}~$GATEWAY_URL~g" monitor/conf/default.json
else
sed -i "s~{{CORE_ADDR}}~$CORE_ADDR~g" monitor/conf/default.json
fi
sed -i "s~{{MONITOR_SERVER_IP}}~$MONITOR_HOST_IP~g" monitor/conf/default.json
if [ $MONITOR_SESSION_ENABLE ]
then
sed -i "s~{{MONITOR_SESSION_ENABLE}}~true~g" monitor/conf/default.json
else
sed -i "s~{{MONITOR_SESSION_ENABLE}}~false~g" monitor/conf/default.json
fi
if [ $PLUGIN_MODE ]
then
sed -i "s~{{PLUGIN_MODE}}~yes~g" monitor/conf/default.json
else
sed -i "s~{{PLUGIN_MODE}}~no~g" monitor/conf/default.json
fi

if [ -n "$MONITOR_LOCAL_DNS_MAP" ]
then
  dns_map=${MONITOR_LOCAL_DNS_MAP}
  set ${dns_map//,/ }
  for v in "$@"
  do
    sudo sh -c "echo '${v//=/ }' >> /etc/hosts"
  done
fi

archive_day="30d"
if [ "$MONITOR_PROMETHEUS_ARCHIVE_DAY" -gt 0 ] 2>/dev/null;
then
  archive_day="${MONITOR_PROMETHEUS_ARCHIVE_DAY}d"
fi

cd agent_manager
sudo mkdir -p logs && sudo chown app:apps logs
tar zxf exporters.tar.gz
# 确保解压出来的文件属于 app:apps
sudo chown -R app:apps . 2>/dev/null || true
ensure_log_file logs/app.log
GODEBUG=netdns=go nohup ./agent_manager > logs/app.log 2>&1 &
cd ../daemon_proc
sudo mkdir -p logs && sudo chown app:apps logs
ensure_log_file logs/app.log
GODEBUG=netdns=go nohup ./daemon_proc > logs/app.log 2>&1 &
cd ../alertmanager
sudo mkdir -p logs && sudo chown app:apps logs
ensure_log_file logs/alertmanager.log
GODEBUG=netdns=go nohup ./alertmanager --config.file=alertmanager.yml --web.listen-address=":9093"  --cluster.listen-address=":9094" > logs/alertmanager.log 2>&1 &
cd ../prometheus/
sudo mkdir -p rules logs && sudo chown app:apps logs rules
# 如果 rules/base.yml 已存在，先删除（解决 PV 挂载时文件权限问题）
sudo rm -f rules/base.yml 2>/dev/null || rm -f rules/base.yml
if [ -f "base.yml" ]; then
  sudo cp -f base.yml rules/ && sudo chown app:apps rules/base.yml
fi
cd /app/monitor/prometheus && ensure_log_file logs/prometheus.log && GODEBUG=netdns=go nohup ./prometheus --config.file=prometheus.yml --web.enable-lifecycle --storage.tsdb.retention.time=${archive_day} > logs/prometheus.log 2>&1 &
cd ../ping_exporter/
sudo chown app:apps . 2>/dev/null || true
sudo mkdir -p logs && sudo chown app:apps logs
ensure_log_file logs/app.log
GODEBUG=netdns=go nohup ./ping_exporter > logs/app.log 2>&1 &
cd ../transgateway/
sudo chown app:apps . 2>/dev/null || true
sudo mkdir -p logs data && sudo chown app:apps logs data
ensure_log_file logs/app.log
GODEBUG=netdns=go nohup ./transgateway -d data -m http://127.0.0.1:8080 > logs/app.log 2>&1 &
cd ../archive_mysql_tool
sudo chown app:apps . 2>/dev/null || true
sudo mkdir -p logs && sudo chown app:apps logs
ensure_log_file logs/app.log
GODEBUG=netdns=go nohup ./archive_mysql_tool > logs/app.log 2>&1 &
cd ../db_data_exporter
sudo chown app:apps . 2>/dev/null || true
sudo mkdir -p logs && sudo chown app:apps logs
ensure_log_file logs/app.log
GODEBUG=netdns=go nohup ./db_data_exporter > logs/app.log 2>&1 &
cd ../metric_comparison_exporter
sudo chown app:apps . 2>/dev/null || true
sudo mkdir -p logs && sudo chown app:apps logs
cd ../monitor/
sudo chown app:apps . 2>/dev/null || true
sudo mkdir -p logs && sudo chown app:apps logs
ensure_log_file logs/app.log
sleep 2
Exit_actions (){
  kill `ps aux|grep -E "prometheus"|grep -v "grep"|awk '{print $1}'` `ps aux|grep -E "transgateway"|grep -v "grep"|awk '{print $1}'`
  wait $!
}
trap Exit_actions INT TERM EXIT
GODEBUG=netdns=go nohup ./monitor-server > logs/app.log 2>&1 &
wait $!
