#!/bin/bash
set -e -x
cd $(dirname $0)/../monitor-server
export CGO_ENABLED=1
go build -ldflags "-linkmode external -extldflags -static -s"