#!/bin/bash
set -x
npm -v
if [ $? -eq 0 ]
then
    cd $1/monitor-ui
    # npm install --registry https://registry.npmmirror.com --unsafe-perm --force
    npm install --force
    npm run build
    npm run plugin
else
    docker run --rm -v $1:/app/open-monitor --name node-build node:12.13.1 /bin/bash /app/open-monitor/build/build-ui-docker.sh
fi