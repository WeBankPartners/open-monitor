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
if [ $GATEWAY_URL ]
then
sed -i "s~{{CORE_ADDR}}~$GATEWAY_URL~g" monitor/conf/default.json
else
sed -i "s~{{CORE_ADDR}}~$CORE_ADDR~g" monitor/conf/default.json
fi
sed -i "s~{{MONITOR_SERVER_PEER}}~$MONITOR_HOST_IP~g" monitor/conf/default.json
if [ $MONITOR_SESSION_ENABLE ]
then
sed -i "s~{{MONITOR_SESSION_ENABLE}}~true~g" monitor/conf/default.json
else
sed -i "s~{{MONITOR_SESSION_ENABLE}}~false~g" monitor/conf/default.json
fi

cd alertmanager
mkdir -p logs
nohup ./alertmanager --config.file=alertmanager.yml --web.listen-address=":9093"  --cluster.listen-address=":9094" > logs/alertmanager.log 2>&1 &
cd ../prometheus/
mkdir -p rules
mkdir -p logs
/bin/cp -f base.yml rules/
nohup ./prometheus --config.file=prometheus.yml --web.enable-lifecycle --storage.tsdb.retention.time=30d > logs/prometheus.log 2>&1 &
cd ../agent_manager/
mkdir -p logs
nohup ./agent_manager > logs/app.log 2>&1 &
cd ../ping_exporter/
mkdir -p logs
nohup ./ping_exporter > logs/app.log 2>&1 &
cd ../transgateway/
mkdir -p logs
mkdir -p data
nohup ./transgateway -d data -m http://127.0.0.1:8080 > logs/app.log 2>&1 &
cd ../archive_mysql_tool
mkdir -p logs
nohup ./archive_mysql_tool > logs/app.log 2>&1 &
cd ../monitor/
mkdir -p logs
sleep 3
./monitor-server --export_agent=true


