#!/bin/bash
set -e -x
cd $(dirname $0)/../monitor-server
export CGO_ENABLED=1
cd ../monitor-agent/agent_manager
go build -buildvcs=false  -ldflags "-linkmode external -extldflags -static -s"
cd ../archive_mysql_tool
go build -buildvcs=false  -ldflags "-linkmode external -extldflags -static -s"
cd ../ping_exporter
go build -buildvcs=false  -ldflags "-linkmode external -extldflags -static -s"
cd ../node_exporter
go build -buildvcs=false  -ldflags "-linkmode external -extldflags -static -s" -o monitor_exporter
cd ../transgateway
go build -buildvcs=false  -ldflags "-linkmode external -extldflags -static -s"
cd ../db_data_exporter
go build -buildvcs=false  -ldflags "-linkmode external -extldflags -static -s"
cd ../daemon_proc
go build -o daemon_proc
cd ../metric_comparison_exporter
go build -buildvcs=false  -ldflags "-linkmode external -extldflags -static -s"