#!/usr/bin/env bash

exporter_type='host'
package_name='exporter_host_v1.tar.gz'
bin_name='node_exporter_new'
mkdir -p /usr/local/monitor
cd /usr/local/monitor
if [[ -d $exporter_type/data ]]
then
    cp -r $exporter_type/data host_tmp
fi
if [[ -d $exporter_type ]]
then
    rm -rf $exporter_type/*
else
    mkdir -p $exporter_type
fi
curl -O {{server_address}}/wecube-agent/$package_name
tar zxf $package_name $exporter_type/
rm -f exporter_$exporter_type.tar.gz
if [[ -d host_tmp ]]
then
    mv host_tmp $exporter_type/data
    rm -rf host_tmp
fi
cd $exporter_type
mkdir -p data
if [[ -f data/app.pid ]]
then
    kill -9 `cat data/app.pid`
else
    kill `pidof node_exporter_new`
fi
if [[ -f data/app.log.old ]]
then
   rm -f data/app.log.old
fi
if [[ -f data/app.log ]]
then
   mv data/app.log data/app.log.old
fi
chmod +x $bin_name
nohup ./$bin_name > data/app.log 2>&1 &
pidof $bin_name > data/app.pid
if [[ -d data/host_tmp ]]
then
    rm -rf data/host_tmp
fi