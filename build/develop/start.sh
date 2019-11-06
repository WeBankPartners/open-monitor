#!/bin/bash
set -ex
if ! docker --version &> /dev/null
then
    echo "must have docker installed"
    exit 1
fi

if ! docker-compose --version &> /dev/null
then
    echo  "must have docker-compose installed"
    exit 1
fi

MONITOR_BASE_PATH=/app/docker/monitor

mkdir -p $MONITOR_BASE_PATH
mkdir -p $MONITOR_BASE_PATH/prometheus/conf
mkdir -p $MONITOR_BASE_PATH/prometheus/rules
mkdir -p $MONITOR_BASE_PATH/alertmanager/conf
mkdir -p $MONITOR_BASE_PATH/monitor-db
mkdir -p $MONITOR_BASE_PATH/prometheus/data
mkdir -p $MONITOR_BASE_PATH/consul/data
mkdir -p $MONITOR_BASE_PATH/alertmanager/data

cp prometheus.yml $MONITOR_BASE_PATH/prometheus/conf
cp alertmanager.yml $MONITOR_BASE_PATH/alertmanager/conf

#get local ip
EXTERNAL_IP=`ifconfig |grep inet|grep -v inet6|grep -v '.0.1 '|head -1|awk '{print $2}'`

osmsg=`uname -a`
if [[ $osmsg =~ "Darwin" ]]
then
    cp docker-compose.tpl docker-compose.yml
    cp monitor.json.tpl monitor.json
    sed -i "" "s~{{MONITOR_BASE_PATH}}~$MONITOR_BASE_PATH~g" docker-compose.yml
    sed -i "" "s~{{MONITOR_BASE_PATH}}~$MONITOR_BASE_PATH~g" monitor.json
    sed -i "" "s~{{EXTERNAL_IP}}~$EXTERNAL_IP~g" $MONITOR_BASE_PATH/alertmanager/conf/alertmanager.yml
else
    sed -i "s~{{MONITOR_BASE_PATH}}~$MONITOR_BASE_PATH~g" docker-compose.tpl > docker-compose.yml
    sed -i "s~{{MONITOR_BASE_PATH}}~$MONITOR_BASE_PATH~g" monitor.json.tpl > monitor.json
    sed -i "s~{{EXTERNAL_IP}}~$EXTERNAL_IP~g" $MONITOR_BASE_PATH/alertmanager/conf/alertmanager.yml
fi

docker-compose  -f docker-compose.yml  up -d