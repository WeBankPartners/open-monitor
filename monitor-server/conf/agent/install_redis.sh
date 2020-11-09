#!/usr/bin/env bash

server_address='{{server_address}}'
exporter_type='{{exporter_type}}'
bin_name='{{bin_name}}'
mkdir -p /usr/local/monitor
cd /usr/local/monitor
if [[ -d $exporter_type/data ]]
then
    cp -r $exporter_type/data tmp_data
fi
if [[ -d $exporter_type ]]
then
    rm -rf $exporter_type/*
else
    mkdir -p $exporter_type
fi
curl -O http://$server_address/monitor/exporter/exporter_$exporter_type.tar.gz
tar zxf exporter_$exporter_type.tar.gz $exporter_type/
rm -f exporter_$exporter_type.tar.gz
if [[ -d tmp_data ]]
then
    mv tmp_data $exporter_type/data
fi
cd $exporter_type
mkdir -p data
if [[ -f data/app.pid ]]
then
    kill -9 `cat data/app.pid`
fi
if [[ -f data/app.log.old ]]
then
   rm -f data/app.log.old
fi
if [[ -f data/app.log ]]
then
   mv data/app.log data/app.log.old
fi
nohup ./$bin_name --redis.addr="redis://{{redis_host}}:{{redis_port}}" --redis.password="{{redis_pwd}}" > data/app.log 2>&1 &
pidof $bin_name > data/app.pid