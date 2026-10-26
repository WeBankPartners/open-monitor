# Kubernetes 日志 Agent

独立 Linux Go 应用。Monitor 保存日志目标和规则，Agent 主动注册、拉取配置、扫描节点日志，通过标准 Prometheus `/metrics` 暴露内存中的结果。Prometheus 负责抓取并写入 TSDB。

首次部署、占位符替换与验收见 [DaemonSet 部署指南](../../docs/k8s-log-agent-deployment.md)；本文说明构建、启动参数和 Agent 的运行边界。

## 构建和测试

在本目录运行：

```bash
make build
make test
make image IMAGE=<registry>/monitor-k8s-log-agent:<tag>
```

部署模板检查依赖 `../../build/tests/requirements.txt`，在 Python 虚拟环境安装后执行 `make deploy-template-check`。该检查覆盖 YAML 引用、PVC 子目录隔离和初始化脚本重复执行，不替代真实 NFS 联调。

镜像和 DaemonSet 默认使用 `10001:10001`，安装 `ca-certificates`、`tzdata`，`TZ` 默认 `Asia/Shanghai`。APK 默认使用阿里云源；构建时可传 `ALPINE_MIRROR=dl-cdn.alpinelinux.org` 切换源域名，基础镜像仓库仍由构建环境配置。

## 启动参数

| ENV | 要求或默认值 |
| --- | --- |
| `LOG_AGENT_MONITOR_URL` | 必填，HTTP/HTTPS Monitor 基础路径，通常以 `/monitor` 结尾 |
| `LOG_AGENT_CLUSTER_ID` | 必填，与 Monitor 登记的 Cluster ID 一致 |
| `LOG_AGENT_NODE_NAME` | 必填，DaemonSet 通过 Downward API 注入 |
| `LOG_AGENT_ADVERTISE_ADDRESS` | 必填，Monitor/Prometheus 可访问的地址，模板注入 Pod IP |
| `LOG_AGENT_AUTH_TOKEN` | 必填，与 Monitor 的 `MONITOR_K8S_LOG_AGENT_TOKEN` 一致；不是 Kubernetes Token |
| `LOG_AGENT_ROOT` | 必填，允许读取的绝对日志根目录 |
| `LOG_AGENT_STATE_DIR` | 必填，checkpoint 绝对目录，模板为 `/state` |
| `LOG_AGENT_HTTP_ADDR` | 默认 `:9105` |
| `LOG_AGENT_POLL_INTERVAL` | 默认 `10s`，至少 `1s`；控制注册、配置轮询和扫描循环 |
| `TZ` | 镜像默认 `Asia/Shanghai` |

ENV 变更需重启 Pod。日志目标及规则不放入 ENV 或 ConfigMap，由 Monitor 动态下发。Prometheus 的抓取间隔独立于 Agent 循环：Monitor 内置两个 Kubernetes job 默认 `30s`，共用 `MONITOR_K8S_METRICS_INTERVAL_SECONDS`。

## 配置与采集流程

1. Agent 向 `<Monitor基础路径>/internal/k8s/log-agent/register` 上报 `clusterId/node/address/port`，并携带 Bearer Token。
2. 通过 `GET <Monitor基础路径>/internal/k8s/log-agent/config` 拉取完整配置；查询参数为 URL 编码后的 `clusterId/node/version`，禁止重定向。
3. 应用配置后扫描文件。HTTP `/metrics` 只读取已生成的内存结果，不在抓取时现场解析日志。

Monitor 按 Pod 的 `instanceId` 查询 CMDB `app_instance` guid，将 `{deploy_path}/logs` 下的普通 `*.log` 和应用 `*_metric.log` 目标分配给该节点。默认普通规则匹配日志头中的 `ERROR/FATAL/PANIC` 级别，支持方括号格式和时间戳后直接跟级别的格式（包括毫秒前使用冒号），正文中的同名字段和异常续行不单独计数。普通扫描显式跳过 `_metric.log`，避免应用指标内容触发关键字告警。

配置采用 `{"config":{"version":1,"targets":[...]}}` 响应，撤销目标通过更高版本配置移除目标；配置、规则版本由 Server 维护。具体协议见 [系统设计](../../docs/k8s-monitoring-system-design.md#5-agent-配置采集与状态恢复)。

拉取失败暂时保留旧配置；超过一分钟没有有效配置响应时，在采集循环的检查点暂停扫描，`/readyz` 和 `/metrics` 返回 503。`/healthz` 仅表示 HTTP 进程可用。注册、拉取和扫描共用循环，慢请求或阻塞文件操作可能拉长实际周期，不能把一分钟当作严格超时保证。

## 可选排除条件

Web 的 Agent 日志目标配置提供“日志匹配与排除规则”，也支持在目标 JSON 中填写可选 `exclude`。以下是人工编辑的规则片段，保存后 Server 补充内部版本：

```json
{
  "id": "default-error",
  "mode": "regex",
  "pattern": "(?i)^\\s*(?:(?:\\[[^\\]\\r\\n]*\\]){0,2}\\[(?:ERROR|FATAL|PANIC)\\]|\\d{4}[-/]\\d{2}[-/]\\d{2}[ T]\\d{2}:\\d{2}:\\d{2}(?:[.,:]\\d+)?[\\t ]+(?:ERROR|FATAL|PANIC)\\b)",
  "exclude": [
    { "mode": "literal", "pattern": "EXPECTED_ERROR" },
    { "mode": "regex", "pattern": "(?i)healthcheck.*timeout" }
  ]
}
```

完整行命中主规则且未命中任何排除条件才计数。排除针对整行、只影响所属规则；被排除行仍推进游标，其他规则可正常计数。`exclude` 缺省或为空数组时不排除，自动默认规则不带排除。

literal 区分大小写；regex 使用 Go regexp 语法，可用 `(?i)` 忽略大小写，不支持前后查找断言。每条规则最多 100 个排除条件，每个非空 pattern 最多 4096 字节，Server 与 Agent 都校验。

Server 在集群配置事务锁内管理规则版本：新增为 1，匹配或排除内容变化才递增，相同配置重试不变。自动发现、Pod 重建和节点切换保留人工规则。新版本使用新计数序列；目录、glob 和 Pod UID 未变时继承已提交游标，不重读历史，已有计数和告警不追溯重算。排除文本不进入指标标签。

已有环境升级 Monitor Server 和 Web 后，可在“K8s监控配置 → 监控源 → 日志 Agent 管理 → 日志目标（JSON）”点击“重置默认日志规则”，检查后点击“保存”。一次处理当前 Agent 下所有普通日志目标的 `default-error` 规则（缺少时补充），清空该默认规则的排除项，保留其他自定义规则和应用指标目标。重置仅修改编辑内容，取消不会下发；重复重置并保存相同内容不会再次递增版本。无需执行 SQL 或升级 Agent 镜像。

## 文件与 checkpoint

- 扫描不递归，单目录最多检查 10000 个条目。普通日志只处理完整行，单行最多 64 KiB；OTLP 单记录最多 1 MiB，超限报告错误。
- 游标记录文件设备号和 inode；首次接管由 `start` 控制，默认目标为 `end`。相同路径换 inode 时新文件从头读取，检测到 truncate 时归零。若轮转旧文件不再匹配 glob，其尚未读取的尾部不能保证补读；copytruncate 也有漏读窗口。
- 成功扫描目标后回收不再匹配的旧文件游标，将计数折叠到目标/规则累计记录，避免 checkpoint 按轮转次数持续增长；读取异常时暂缓该目标回收。配置撤销时清理对应目标、规则和旧 Pod 游标。
- checkpoint 通过文件锁、临时文件、文件同步和原子重命名提交。文件锁防止同目录双写；已有进程持锁时，新进程拒绝启动。模板的 RWX PVC 按 `k8s-log-agent/<namespace>/<node>` 分目录，同节点重建沿用原状态。
- 扫描先更新内存，再尝试保存。state 损坏或读写返回错误时可降级为空状态或内存游标，后续重试写当前状态，不回读过期 checkpoint；新建默认目标从日志末尾接管。NFS 调用若阻塞，仍可能阻塞采集和 HTTP 请求，不能保证存储故障时持续服务。
- checkpoint 尚未恢复时重启可能重置计数并漏计或重复计数，影响窗口可能覆盖整个未持久化期间，不能固定为一个扫描周期。PVC 挂载或 initContainer 失败属于启动前故障，Agent 无法自行绕过。

这不是原始日志存储或精确一次交付方案。容量取决于当前目标、活跃文件和规则数；游标回收避免随轮转无限增长，但不等于任意负载下 1Gi 都足够。

## 指标与健康

| 指标组 | 主要输出 |
| --- | --- |
| Agent | `k8s_log_agent_up`、`k8s_log_agent_config_version` |
| checkpoint | `k8s_log_agent_state_available`、`k8s_log_agent_state_resets_total`、`k8s_log_agent_state_persist_errors_total` |
| 普通日志 | `k8s_log_match_total`、`k8s_log_last_match_timestamp`、`k8s_log_target_healthy`、`k8s_log_file_missing`、`k8s_log_read_error_total`、`k8s_log_last_read_timestamp` |
| 应用指标 | 白名单内的 `k8s_app_*` 及可用性、读取错误、源样本时间指标 |

匹配指标包含 `cluster_id/namespace/workload_type/workload_name/pod/pod_uid/container/instance_id`，其中 `instance_id` 是 CMDB guid。Agent 级状态按 `cluster_id/node` 标记；文件健康等指标按各自用途携带目标或 Pod 标签，并非每条序列都有全部业务标签。不要将日志正文、路径或匹配内容作为标签。

应用指标读取文件尾部约 2 MiB 中最后一条完整 OTLP JSON 记录，最多 4096 条样本，要求最近 5 分钟内且不领先当前时间超过 5 秒。没有 `*_metric.log` 时静默跳过，不输出该目标的应用和诊断指标；文件存在但读取、解析失败或过期时报告不可用。应用指标不进入普通日志游标和匹配计数。完整目录、单位与扩展方法见 [应用指标白名单](../../docs/k8s-app-metric-whitelist.md)。

Unix 指标时间不因 `TZ` 偏移。普通关键字匹配不按业务时间计算；应用记录优先使用 `timeUnixNano`，缺省时按 `TZ` 解析行首时间。
