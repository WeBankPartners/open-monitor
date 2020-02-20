#!/bin/bash

echo "start run"
sed -i "s~{{MONITOR_SERVER_PORT}}~$MONITOR_SERVER_PORT~g" alertmanager/alertmanager.yml
sed -i "s~{{MONITOR_SERVER_PORT}}~$MONITOR_SERVER_PORT~g" monitor/conf/default.json
sed -i "s~{{MONITOR_DB_HOST}}~$MONITOR_DB_HOST~g" monitor/conf/default.json
sed -i "s~{{MONITOR_DB_PORT}}~$MONITOR_DB_PORT~g" monitor/conf/default.json
sed -i "s~{{MONITOR_DB_USER}}~$MONITOR_DB_USER~g" monitor/conf/default.json
sed -i "s~{{MONITOR_DB_PWD}}~$MONITOR_DB_PWD~g" monitor/conf/default.json
if [ $MONITOR_SESSION_ENABLE ]
then
sed -i "s~{{MONITOR_SESSION_ENABLE}}~true~g" monitor/conf/default.json
else
sed -i "s~{{MONITOR_SESSION_ENABLE}}~false~g" monitor/conf/default.json
fi

cd consul
mkdir -p logs
mkdir -p data
nohup ./consul agent -server -ui -http-port 8500 -client 0.0.0.0 -bind 127.0.0.1 -data-dir data -bootstrap-expect 1 > logs/consul.log 2>&1 &
cd ../alertmanager/
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
cd ../transgateway/
mkdir -p logs
mkdir -p data
nohup ./transgateway -d data -m http://127.0.0.1:8080 > logs/app.log 2>&1 &
cd ../monitor/
mkdir -p logs
./monitor-server --export_agent=true


