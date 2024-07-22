#!/bin/bash

tmppath=$0
exporter_type='host'
bin_name='node_exporter_new'
package_path=${tmppath/\/start\.sh/}
deploy_path=$1
if [ -z $deploy_path ];then
  deploy_path='/usr/local/monitor'
fi
exporter_port=$2
if [ -z $exporter_port ];then
  exporter_port='9100'
fi
sed -i "s~{{exporter_port}}~$exporter_port~g" $package_path/control
chmod +x $package_path/$bin_name $package_path/control
mkdir -p $deploy_path/$exporter_type
mkdir -p $deploy_path/$exporter_type/data
cleanprocess='yes'
if [ -f $deploy_path/host/VERSION ]
then
  if [ "`cat $deploy_path/host/VERSION | grep v1.11.3.5`" == "v1.11.3.5" ]
  then
    cleanprocess='no'
  fi
fi
rm -f $deploy_path/host/VERSION
/bin/cp -f $package_path/* $deploy_path/$exporter_type/
cd $deploy_path/$exporter_type/
rm -f start.sh
./control stop
if [ "$cleanprocess" == "yes" ]
then
  rm -f $deploy_path/host/data/process_cache.data
fi
./control start
for i in `seq 1 60`
do
  if [ "`netstat -lntp|grep \":${exporter_port}\"|wc -l`" = "1" ]
  then
    break
  else
    sleep 1
  fi
done

cronfile="/var/spool/cron/root"
if [[ ! -f $cronfile ]];
then
    result=0
else
    result=`grep -c 'monitor-agent-host' ${cronfile}`
fi

if [[ $result -eq 0 ]];
then
    echo "#Ansible: monitor-agent-host
* * * * * cd $deploy_path/$exporter_type;./control start > /dev/null 2>&1" >> /var/spool/cron/root;
fi