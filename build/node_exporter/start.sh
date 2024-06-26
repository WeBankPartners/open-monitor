#!/bin/bash

tmppath=$0
exporter_type='host'
bin_name='node_exporter_new'
package_path=${tmppath/\/start\.sh/}
chmod +x $package_path/$bin_name $package_path/control
mkdir -p /usr/local/monitor/$exporter_type
mkdir -p /usr/local/monitor/$exporter_type/data
cleanprocess='yes'
if [ -f /usr/local/monitor/host/VERSION ]
then
  if [ "`cat /usr/local/monitor/host/VERSION | grep v1.11.3.5`" == "v1.11.3.5" ]
  then
    cleanprocess='no'
  fi
fi
rm -f /usr/local/monitor/host/VERSION
/bin/cp -f $package_path/* /usr/local/monitor/$exporter_type/
cd /usr/local/monitor/$exporter_type/
rm -f start.sh
./control stop
if [ "$cleanprocess" == "yes" ]
then
  rm -f /usr/local/monitor/host/data/process_cache.data
fi
./control start
for i in `seq 1 60`
do
  if [ "`netstat -lntp|grep ':9100'|wc -l`" = "1" ]
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
* * * * * cd /usr/local/monitor/$exporter_type;./control start > /dev/null 2>&1" >> /var/spool/cron/root;
fi