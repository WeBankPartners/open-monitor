#!/bin/bash
set -e -x
cd $(dirname $0)/../monitor-ui
# npm install
npm run build
npm run plugin