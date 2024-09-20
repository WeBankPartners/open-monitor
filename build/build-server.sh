#!/bin/bash
set -e -x
cd $(dirname $0)/../monitor-server
#go build -ldflags "-linkmode external -extldflags -static -s"
cd ../monitor-agent/agent_manager
go build -ldflags "-linkmode external -extldflags -static -s"
cd ../archive_mysql_tool
go build -ldflags "-linkmode external -extldflags -static -s"
cd ../ping_exporter
go build -ldflags "-linkmode external -extldflags -static -s"
cd ../node_exporter
go build -ldflags "-linkmode external -extldflags -static -s" -o monitor_exporter
cd ../transgateway
go build -ldflags "-linkmode external -extldflags -static -s"
cd ../db_data_exporter
go build -ldflags "-linkmode external -extldflags -static -s"
cd ../daemon_proc
go build -o daemon_proc
cd ../metric_comparison_exporter
go build -ldflags "-linkmode external -extldflags -static -s"