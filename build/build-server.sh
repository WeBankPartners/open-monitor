#!/bin/bash
set -e -x

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
PROJECT_ROOT=$(cd "${SCRIPT_DIR}/.." && pwd)

# 固定输出 ARM64 版本
export GOOS=linux
export GOARCH=arm64
unset GOARM
export CGO_ENABLED=${CGO_ENABLED:-1}
echo ">> Building components for ${GOOS}/${GOARCH}"

cd "${PROJECT_ROOT}/monitor-server"
#go build -ldflags "-linkmode external -extldflags -static -s"

cd "${PROJECT_ROOT}/monitor-agent/agent_manager"
go build -ldflags "-linkmode external -extldflags -static -s"

cd "${PROJECT_ROOT}/monitor-agent/archive_mysql_tool"
go build -ldflags "-linkmode external -extldflags -static -s"

cd "${PROJECT_ROOT}/monitor-agent/ping_exporter"
go build -ldflags "-linkmode external -extldflags -static -s"

cd "${PROJECT_ROOT}/monitor-agent/node_exporter"
go build -ldflags "-linkmode external -extldflags -static -s" -o monitor_exporter

cd "${PROJECT_ROOT}/monitor-agent/transgateway"
go build -ldflags "-linkmode external -extldflags -static -s"

cd "${PROJECT_ROOT}/monitor-agent/db_data_exporter"
go build -ldflags "-linkmode external -extldflags -static -s"

cd "${PROJECT_ROOT}/monitor-agent/daemon_proc"
go build -o daemon_proc

cd "${PROJECT_ROOT}/monitor-agent/metric_comparison_exporter"
go build -ldflags "-linkmode external -extldflags -static -s"