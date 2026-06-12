#!/bin/bash
set -e -x

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

if npm -v >/dev/null 2>&1; then
    cd "$1/monitor-ui"
    if ! install_with_mirrors "${REGISTRIES[@]}"; then
        echo "所有 npm 镜像安装失败，终止构建" >&2
        exit 1
    fi
    npm run build
    npm run plugin
else
    docker run --rm -v "$1":/app/open-monitor --name node-build node:12.13.1 /bin/bash /app/open-monitor/build/build-ui-docker.sh
fi