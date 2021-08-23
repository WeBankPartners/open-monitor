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

mkdir -p /app/docker/prometheus
mkdir -p /app/docker/prometheus/rules
mkdir -p /app/docker/alertmanager
mkdir -p /app/docker/monitor/conf
mkdir -p /app/docker/monitor/logs
mkdir -p /data/docker/monitor-db
mkdir -p /data/docker/prometheus
mkdir -p /data/docker/consul
mkdir -p /data/docker/alertmanager
cp conf/prometheus.yml /app/docker/prometheus
cp conf/alertmanager.yml /app/docker/alertmanager
cp conf/monitor.json /app/docker/monitor/conf/default.json

source monitor.cfg

sed "s~{{MONITOR_DATABASE_IMAGE_NAME}}~$database_image_name~g" docker-compose.tpl > docker-compose.yml
sed -i "s~{{MYSQL_ROOT_PASSWORD}}~$database_init_password~g" docker-compose.yml
sed -i "s~{{MONITOR_IMAGE_NAME}}~$monitor_image_name~g" docker-compose.yml
sed -i "s~{{MONITOR_SERVER_PORT}}~$monitor_server_port~g" docker-compose.yml

sed -i "s~{{MYSQL_ROOT_PASSWORD}}~$database_init_password~g" /app/docker/monitor/conf/default.json
sed -i "s~{{MONITOR_SERVER_PORT}}~$monitor_server_port~g" /app/docker/monitor/conf/default.json

sed -i "s~{{MONITOR_SERVER_PORT}}~$monitor_server_port~g" /app/docker/alertmanager/alertmanager.yml

docker-compose  -f docker-compose.yml  up -d

 










