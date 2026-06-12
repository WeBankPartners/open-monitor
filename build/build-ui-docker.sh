#!/bin/bash
set -e -x

cd "$(dirname "$0")"/../monitor-ui

install_with_mirrors() {
    local registry
    for registry in "$@"; do
        echo "尝试使用 npm registry: $registry"
        if npm install --registry "$registry" --unsafe-perm --force; then
            echo "npm install 成功，registry: $registry"
            return 0
        fi
        echo "npm install 失败，registry: $registry，尝试下一个镜像"
    done
    return 1
}

REGISTRIES=(
    https://mirrors.cloud.tencent.com/npm/
    https://registry.npmmirror.com
    https://registry.npmjs.org/
)

install_with_mirrors "${REGISTRIES[@]}"
npm run build
npm run plugin