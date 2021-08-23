#!/bin/bash
set -e -x
mv monitor-agent/node_exporter/node_exporter_new build/node_exporter/
/bin/cp -f monitor-agent/node_exporter/VERSION build/node_exporter/
cd build
tar zcf node_exporter.tar.gz node_exporter
mv node_exporter.tar.gz ../