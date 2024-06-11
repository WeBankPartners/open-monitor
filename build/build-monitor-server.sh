#!/bin/bash
set -e -x
cd $(dirname $0)/../monitor-server
go build -ldflags "-linkmode external -extldflags -static -s"