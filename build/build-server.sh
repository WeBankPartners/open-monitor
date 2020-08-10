#!/bin/bash
set -e -x
cd $(dirname $0)/../monitor-server
go build -ldflags "-linkmode external -extldflags -static -s"
cd ../monitor-agent/agent_manager
go build -ldflags "-linkmode external -extldflags -static -s"
cd ../archive_mysql_tool
go build -ldflags "-linkmode external -extldflags -static -s"
cd ../ping_exporter
go build -ldflags "-linkmode external -extldflags -static -s"
cd ../node_exporter
go build -ldflags "-linkmode external -extldflags -static -s" -o node_exporter_new
cd ../transgateway
go build -ldflags "-linkmode external -extldflags -static -s"