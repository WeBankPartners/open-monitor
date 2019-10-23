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

if [ $# -nq 2 ]
then
  echo "usage: deploy_remote.sh config remote_host"
  exit 1 
else
  export config=$1
  export remote_host=$2
fi

source ${config}

cd $(dirname $0)

sed "s~{{MONITOR_DATABASE_IMAGE_NAME}}~$database_image_name~g" docker-compose.tpl > docker-compose.yml
sed -i "s~{{MYSQL_ROOT_PASSWORD}}~$database_init_password~g" docker-compose.yml
sed -i "s~{{MONITOR_IMAGE_NAME}}~$monitor_image_name~g" docker-compose.yml
sed -i "s~{{MONITOR_SERVER_PORT}}~$monitor_server_port~g" docker-compose.yml

docker-compose -f docker-compose.yml -H ${remote_host} up -d

 