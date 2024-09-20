#!/bin/bash
set -e -x
mv monitor-agent/node_exporter/monitor_exporter build/node_exporter/
/bin/cp -f monitor-agent/node_exporter/VERSION build/node_exporter/
cd build
chmod 644 node_exporter/VERSION
tar zcf node_exporter.tar.gz node_exporter
mv node_exporter.tar.gz ../