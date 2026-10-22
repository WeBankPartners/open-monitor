# Kubernetes log Agent

独立 Linux Go 应用；启动参数全部使用 ENV，动态规则主动拉取 Monitor。

## 构建和测试

在本目录运行 `make build`、`make test`；镜像使用 `make image IMAGE=<image>`。
校验部署模板时，先在 Python 虚拟环境中安装 `python3 -m pip install -r ../../build/tests/requirements.txt`，再运行 `make deploy-template-check`。该检查验证合并 YAML 的配置引用、PVC 节点隔离和初始化脚本的重复执行；不替代实际 NFS 联调。
镜像与 DaemonSet 默认使用 `10001:10001`。运行时安装 `ca-certificates` 和 `tzdata`；`TZ` 默认 `Asia/Shanghai`，可在 DaemonSet 中改为其他 IANA 时区。指标时间戳使用 Unix 时间，不因时区变更偏移；普通关键字日志不按业务时间计算；OTLP 指标优先使用 timeUnixNano，缺省时按 TZ 解析行首时间。

APK 默认使用阿里云镜像源；可通过 `make image IMAGE=<image> ALPINE_MIRROR=dl-cdn.alpinelinux.org` 切回官方源。该参数仅替换 APK 源域名，保留 Alpine 版本路径，不影响基础镜像拉取；基础镜像加速需配置构建环境的镜像仓库。
`deploy/daemonset.yaml` 是包含 PVC、ConfigMap 和 DaemonSet 三个文档的占位符模板；一次 `kubectl apply -f` 会创建 1Gi RWX NFS PVC `monitor-k8s-log-agent-state`、ConfigMap `monitor-k8s-log-agent-config` 和 DaemonSet `monitor-k8s-log-agent`。Agent 镜像默认名为 `monitor-k8s-log-agent`。日志目录需预先允许 UID/GID `10001:10001` 读取；initContainer 会在 PVC 中创建 `k8s-log-agent/<namespace>/<node>` 子目录。模板内 Agent Token 默认值必须与 Monitor 的 `MONITOR_K8S_LOG_AGENT_TOKEN` 一致；这是公开的共享默认凭据，生产环境应每个环境单独生成并建议改用 Secret 注入。initContainer 不使用 root，不执行递归 `chown`，也不修改业务日志目录权限。完整替换说明见 [DaemonSet 部署指南](../../docs/k8s-log-agent-deployment.md)。

## ENV

必填：`LOG_AGENT_ROOT`、`LOG_AGENT_STATE_DIR`（绝对目录）、`LOG_AGENT_MONITOR_URL`（HTTP/HTTPS Monitor 基础路径）、`LOG_AGENT_CLUSTER_ID`、`LOG_AGENT_NODE_NAME`、`LOG_AGENT_AUTH_TOKEN`、`LOG_AGENT_ADVERTISE_ADDRESS`。Agent 启动及周期性向 Monitor 注册实际地址和监听端口。生产环境优先使用 HTTPS；HTTP 仅适用于受信任的集群内网络。
可选：`LOG_AGENT_HTTP_ADDR`（默认 `:9105`）、`LOG_AGENT_POLL_INTERVAL`（默认 `10s`，至少 `1s`）。凭据由部署时注入，不提交真实值。

Agent 请求 `<Monitor基础路径>/internal/k8s/log-agent/config`，携带 Bearer 认证及 URL 编码的 `clusterId/node/version`。服务端按 cluster/node 返回完整配置；撤销目标时发布更高版本的空 targets 配置。禁止重定向。

有效配置取得后定时扫描。拉取失败保留旧配置，超过一分钟无有效响应时暂停采集、readiness 与 metrics 返回 503，避免无限期执行已可能撤销的目标。健康检查 `/healthz` 表示进程 HTTP 服务可用，不等同于采集正常。

当前输出 `k8s_log_match_total`；计数在内存中立即推进，并尽力保存到 checkpoint。支持 literal/Go regex、完整行、重启半行续读、文件身份和同一路径轮转后新文件从头读取。成功扫描目标目录后，Agent 会回收不再匹配的旧文件游标，并把历史计数折叠到每条规则的累计记录，因此日志轮转不会让 checkpoint 随轮转次数持续增长；目录或文件读取异常时暂缓该目标的回收。初次接管默认由目标的 `start` 指定 beginning/end。NFS checkpoint 不可用时 Agent 降级为内存游标继续采集，暴露 `k8s_log_agent_state_available`、`k8s_log_agent_state_resets_total`、`k8s_log_agent_state_persist_errors_total`；恢复写入时保存当前内存状态，不回读旧 checkpoint。进程在 checkpoint 恢复前重启可能造成计数重置、日志漏计或重复计数。PVC 挂载失败或 initContainer 无法创建子目录属于 Pod 启动前故障，Agent 无法绕过。

### 可选排除条件

规则由 Monitor Server 保存，Agent 主动拉取。Web 的 Agent 日志目标配置中点击“日志匹配与排除规则”，选择目标后编辑主规则与排除条件；也可在目标 JSON 中填写可选的 `exclude`：

```json
{
  "id": "default-error",
  "mode": "regex",
  "pattern": "(?i)\\b(ERROR|FATAL|PANIC)\\b",
  "exclude": [
    { "mode": "literal", "pattern": "EXPECTED_ERROR" },
    { "mode": "regex", "pattern": "(?i)healthcheck.*timeout" }
  ]
}
```

每个完整日志行命中主规则，且没有命中任何排除条件，才增加该规则计数。排除对整行判断，仅影响所属规则；被排除行照常推进读取游标，其他规则仍可计数。`exclude` 不填或为 `[]` 时不排除，自动生成的默认规则不带排除。普通文本区分大小写，正则使用 Go regexp 语法，可用 `(?i)` 忽略大小写，不支持前后查找断言。每条规则最多 100 个排除条件，每个非空 pattern 最多 4096 字节；Server 和 Agent 均校验模式、长度及正则合法性。

人工保存时规则版本由 Server 在已有集群配置事务锁内管理：新增规则为 1，匹配或排除条件变化才递增版本，相同配置重试保持版本不变。调用方无需填写规则 version，Agent 收到的配置仍包含 version。自动发现、Pod 重建和节点切换保留人工配置的排除条件。新规则版本使用新计数序列，在日志目录、glob 和 Pod UID 未变化时继承已提交读取位置（包括 Agent 重启后），不重读历史，也不跳过尚未读取的行；规则变更前已累计的计数和历史告警不会被追溯重算。排除内容不进入指标标签，增加排除条件不会增加每个规则的活跃序列数。

以 `*_metric.log` 为默认 glob 的 `otlp_metric` 目标读取文件最后一条完整 OTLP JSON 记录，将白名单内的应用指标转换为 `k8s_app_*`。该目标使用最近 5 分钟的新鲜样本，单目标最多 4096 条序列，单条记录最多 1 MiB；没有匹配文件时静默跳过且不输出应用指标，文件存在但读取或解析失败时输出可用性和错误计数，不会把它误报为普通关键字日志异常。指标目标不进入普通日志游标和匹配计数。

## 未完成及边界

第5步的 Agent 配置拉取、认证、版本化配置存储、基础日志扫描、Snapshot 目标重绑定和 Prometheus file_sd 接入已经实现。规则和目录由 Monitor 根据 Pod instanceId 与 CMDB deploy_path 自动生成：日志目录为 `{deploy_path}/logs`，并支持 Web 调整；实际日志目录映射、Agent 地址和 DaemonSet 仍需环境联调确认。当前读取失败只影响对应目标，正常目标仍可提交 checkpoint。

普通日志当前仅支持最大 64 KiB 完整行，指标日志单条上限为 1 MiB，超限均报错，不静默跳过。copytruncate 检测存在不可避免的漏读窗口；节点日志丢失不能恢复。匹配计数不是精确一次告警交付保证。请勿把日志正文、路径或匹配内容作为指标标签。

当前指标还包括 `k8s_log_last_match_timestamp`、`k8s_log_file_missing`、`k8s_log_read_error_total`、`k8s_log_last_read_timestamp`、`k8s_log_agent_up` 和 `k8s_log_agent_config_version`。配置版本变更会清理已撤销目标和规则的本地游标，避免旧 Pod/规则继续复用状态。

指标对象标签统一为 `cluster_id`、`namespace`、`workload_type`、`workload_name`、`pod`、`pod_uid`、`container`，匹配指标保留 `instance_id`。`k8s_log_target_healthy` 表示当前目标扫描是否成功，读取异常或文件缺失时为 0。Monitor 自动创建的日志告警按最近 5 分钟新增匹配数评估，采集缺失不会作为恢复。

应用指标完整白名单、排除项和扩展方法见 [应用指标说明](../../docs/k8s-app-metric-whitelist.md)。
