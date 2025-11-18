#!/bin/bash
set -e -x

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
PROJECT_ROOT=$(cd "${SCRIPT_DIR}/.." && pwd)

# 固定输出 ARM64 版本
export GOOS=linux
export GOARCH=arm64
unset GOARM
export CGO_ENABLED=${CGO_ENABLED:-1}
echo ">> Building monitor-server for ${GOOS}/${GOARCH}"

cd "${PROJECT_ROOT}/monitor-server"
go build -ldflags "-linkmode external -extldflags -static -s"