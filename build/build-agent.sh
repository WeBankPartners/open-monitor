#!/bin/bash
set -e -x
mv monitor-agent/node_exporter/node_exporter_new build/conf/node_exporter/
cd build/conf
tar zcf node_exporter_v2.1.tar.gz node_exporter
mv node_exporter_v2.1.tar.gz ../../