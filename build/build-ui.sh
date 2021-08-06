#!/bin/bash
set -e -x
npm -v
if [ $? -eq 0 ]
then
    cd $1/monitor-ui
    npm install
    npm run build
    npm run plugin
else
    docker run --rm -v $1:/app/open-monitor --name node-build node:12.13.1 /bin/bash /app/open-monitor/build/build-ui-docker.sh
fi