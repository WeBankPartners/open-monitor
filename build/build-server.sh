#!/bin/bash
set -e -x
cd $(dirname $0)/../monitor-server
LINKFLAGS="-linkmode external -extldflags -static -s"
go build -ldflags "-X main.VERSION=0.9 $LINKFLAGS"