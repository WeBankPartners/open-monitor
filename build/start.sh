#!/bin/bash

echo "start run"
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
sed -i "s~{{MONITOR_ARCHIVE_ENABLE}}~$MONITOR_ARCHIVE_ENABLE~g" monitor/conf/default.json
sed -i "s~{{MONITOR_ARCHIVE_MYSQL_HOST}}~$MONITOR_ARCHIVE_MYSQL_HOST~g" monitor/conf/default.json
sed -i "s~{{MONITOR_ARCHIVE_MYSQL_PORT}}~$MONITOR_ARCHIVE_MYSQL_PORT~g" monitor/conf/default.json
sed -i "s~{{MONITOR_ARCHIVE_MYSQL_USER}}~$MONITOR_ARCHIVE_MYSQL_USER~g" monitor/conf/default.json
sed -i "s~{{MONITOR_ARCHIVE_MYSQL_PWD}}~$MONITOR_ARCHIVE_MYSQL_PWD~g" monitor/conf/default.json
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
    echo "${v//=/ }" >> /etc/hosts
  done
fi

archive_day="30d"
if [ "$MONITOR_PROMETHEUS_ARCHIVE_DAY" -gt 0 ] 2>/dev/null;
then
  archive_day="${MONITOR_PROMETHEUS_ARCHIVE_DAY}d"
fi

cd agent_manager
mkdir -p logs
tar zxf exporters.tar.gz
#nohup ./agent_manager > logs/app.log 2>&1 &
cd ../daemon_proc
nohup ./daemon_proc > app.log 2>&1 &
cd ../alertmanager
mkdir -p logs
#nohup ./alertmanager --config.file=alertmanager.yml --web.listen-address=":9093"  --cluster.listen-address=":9094" > logs/alertmanager.log 2>&1 &
cd ../prometheus/
mkdir -p rules
mkdir -p logs
/bin/cp -f base.yml rules/
nohup ./prometheus --config.file=prometheus.yml --web.enable-lifecycle --storage.tsdb.retention.time=${archive_day} > logs/prometheus.log 2>&1 &
cd ../ping_exporter/
mkdir -p logs
#nohup ./ping_exporter > logs/app.log 2>&1 &
cd ../transgateway/
mkdir -p logs
mkdir -p data
#nohup ./transgateway -d data -m http://127.0.0.1:8080 > logs/app.log 2>&1 &
cd ../archive_mysql_tool
mkdir -p logs
#nohup ./archive_mysql_tool > logs/app.log 2>&1 &
cd ../db_data_exporter
mkdir -p logs
#nohup ./db_data_exporter > logs/app.log 2>&1 &
cd ../metric_comparison_exporter
mkdir -p logs
cd ../monitor/
mkdir -p logs
sleep 2
Exit_actions (){
  kill `ps aux|grep -E "prometheus"|grep -v "grep"|awk '{print $1}'` `ps aux|grep -E "transgateway"|grep -v "grep"|awk '{print $1}'`
  wait $!
}
trap Exit_actions INT TERM EXIT
nohup ./monitor-server > logs/app.log 2>&1 &
wait $!