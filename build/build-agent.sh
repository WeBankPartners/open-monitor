#!/bin/bash
set -e -x
mv monitor-agent/node_exporter/node_exporter_new build/conf/node_exporter/
/bin/cp -f monitor-agent/node_exporter/VERSION build/conf/node_exporter/
cd build/conf
tar zcf node_exporter.tar.gz node_exporter
mv node_exporter.tar.gz ../../