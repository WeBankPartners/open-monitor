#!/bin/bash

tmppath=$0
exporter_type='host'
bin_name='monitor_exporter'
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
sudo mkdir -p $deploy_path/$exporter_type
sudo mkdir -p $deploy_path/$exporter_type/data
cleanprocess='yes'
if [ -f $deploy_path/$exporter_type/VERSION ]
then
  if [ "`cat $deploy_path/$exporter_type/VERSION | grep v1.11.3.5`" == "v1.11.3.5" ]
  then
    cleanprocess='no'
  fi
fi
if [ -f $deploy_path/$exporter_type/control ]
then
  cd $deploy_path/$exporter_type/ && sudo ./control stop
  cd -
fi
sudo rm -f $deploy_path/$exporter_type/VERSION
sudo /bin/cp -f $package_path/* $deploy_path/$exporter_type/
cd $deploy_path/$exporter_type/
sudo rm -f start.sh
if [ "$cleanprocess" == "yes" ]
then
  sudo rm -f $deploy_path/host/data/process_cache.data
fi
sudo ./control start
for i in `seq 1 60`
do
  if [ "`sudo netstat -lntp|grep \":${exporter_port}\"|wc -l`" = "1" ]
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
    result=`sudo grep -c "monitor-agent-host ${deploy_path}" ${cronfile}`
fi

if [[ $result -eq 0 ]];
then
    sudo bash -c "echo \"#Ansible: monitor-agent-host ${deploy_path}
* * * * * cd $deploy_path/$exporter_type;./control start > /dev/null 2>&1\" >> /var/spool/cron/root;"
fi