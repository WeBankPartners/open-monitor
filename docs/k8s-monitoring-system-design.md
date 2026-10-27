# Kubernetes 监控 MVP 系统设计

本文记录当前代码实现，供开发、评审和联调使用。它描述静态代码行为，不代表已在目标 Kubernetes 环境完成联调。产品范围、权限申请和容量风险见[评审方案](k8s-monitoring-review.md)；Agent 首次部署见[部署指南](k8s-log-agent-deployment.md)。

本地资源/日志指标抓取与 Kubernetes 告警由 `MONITOR_K8S_LOCAL_COLLECTION_ENABLED` 控制，默认 `N`。有效状态要求原总开关同时开启；关闭时发现、CMDB 日志解析、Agent 配置和 exporter 保留，停止 cAdvisor、本地两个 job 的目标及告警发送。历史状态保留、pending 通知分批取消；详见[开关与升级验收](k8s-local-collection-switch.md)。以下资源采集与告警流程以有效本地采集开启为前提。

## 1. 架构与边界

每个环境部署一套 Monitor 和一套日志 Agent DaemonSet。用户在 Web 登记 cluster 与允许监控的 namespace；Monitor 自动发现 StatefulSet、Pod、容器，抓取资源指标，生成节点日志目标并评估告警。Prometheus 抓取 Monitor 与 Agent 的缓存指标并写入 TSDB；原始日志不进入 TSDB。

| 组件 | 当前职责 | 持久化/交互 |
| --- | --- | --- |
| Monitor K8s Worker | list/watch StatefulSet 和 Pod，定时重列举并替换对象快照 | MySQL 保存登记和快照；Watch 变化触发重新发现 |
| cAdvisor Collector | 按登记 Pod 所在节点经 API Server proxy 读取 cAdvisor，并映射到当前 Pod UID、容器 ID | 内存缓存最新采集结果，暴露 `/monitor/internal/k8s/metrics` |
| Log Agent DaemonSet | 每节点读取 Monitor 下发的日志目标，输出匹配计数及 OTLP 应用指标 | 本地 checkpoint；`/metrics` 提供内存样本 |
| Prometheus / TSDB | 抓取上述 endpoint、提供 PromQL 查询和保存时间序列 | Kubernetes 两个 scrape job 默认 30s，TSDB 默认保留 30 天 |
| Alert runtime | 评估规则、持久化告警状态与通知 outbox，SMTP 重试 | MySQL 事务与唯一约束；不保证邮件 exactly-once |
| Web | Cluster/Namespace、Agent 目标/端点、规则和监控对象页面 | 经 Monitor 已认证的 API 操作 |

Monitor 的 cAdvisor 采集、`k8s_monitor` 与 `k8s_log_agent` Prometheus scrape 周期共用 `MONITOR_K8S_METRICS_INTERVAL_SECONDS`，默认 30 秒，可设 15～300 秒；Monitor 重启后生效。Prometheus file_sd 目标刷新和 Agent 配置/扫描周期是独立的 10 秒。Agent poll interval 默认 10 秒，可由 `LOG_AGENT_POLL_INTERVAL` 调整。

## 2. 登记与对象发现

### Cluster 与 Namespace

- Cluster 登记包含 Kubernetes API Server、可选 CA 和凭据；Cluster ID 由使用方确定。未提供 CA 时，客户端跳过 TLS 证书校验，因此只应在管理员接受该风险时使用。
- Namespace 登记不包含独立 API Server 或 CA。Namespace Token 可留空；此时运行时使用所属 Cluster Token，不会把 Cluster Token 复制进 Namespace 记录。填写 Namespace Token 时仅作为该 Namespace 的 Token 覆盖，API Server 与 CA 仍继承 Cluster。
- Web 与 WeCube 的登记都不要求调用方提交 version。Web 使用 PUT/DELETE；WeCube 使用批量 envelope 中的 `put/delete` action。Monitor 内部维护数据库 version；相同目标状态的重复 PUT 成功且不重复写，重复 DELETE 成功。
- Cluster 连接身份用于 cAdvisor 节点代理读取；namespace 连接身份用于该 namespace 的 StatefulSet、Pod 发现。Token 需有相应只读授权，详细范围见[权限评审](k8s-monitoring-review.md#3-kubernetes-权限)。

### 发现、重启和生命周期

启用且在 Monitor 允许列表内的登记项开始 Worker。Worker 启动先 list 并发布快照，再 watch；资源变化、watch 断连、resourceVersion 过期或 5 分钟周期重列举会重新发现并原子替换当前快照。Monitor 重启时从数据库恢复登记，再 list 补齐停机期间的资源变化。发现失败时暴露数据质量状态，不能把旧样本当作新鲜样本。

对象关联使用 cluster、namespace、workload type/name、pod name、container name 等可读名称；运行实例消歧使用 Workload UID、Pod UID 和 Container ID。STS/Pod 同名更新可延续业务对象视图，Pod UID 变化表示新运行实例，历史样本不会错误并入新实例。cAdvisor 样本只有在 UID 与当前快照容器映射成功时才会暴露。

## 3. 资源指标与存储

主采集路径是 Monitor 通过 Kubernetes API Server 的 `/api/v1/nodes/{node}/proxy/metrics/cadvisor` 查询目标节点 cAdvisor。代码没有 `metrics.k8s.io` 自动回退；它不是当前实现的数据源。集群身份需要获批 `nodes/proxy get`，namespace 发现身份需要 `pods` 与 `statefulsets` 的 `get/list/watch`。

输出包括：CPU 使用 counter、memory working set、CPU CFS throttling、CPU/memory request 与 limit、restart count、Pod/container Ready、Pod Phase、OOM 事件/最后终止 OOM、memory failcnt、网络收发 counter、块设备读写字节 counter。块设备 counter 可经 PromQL 算 bytes/s；当前没有 IOPS 或 I/O latency 数据。Request/limit/重启/Ready/Phase 来自 Kubernetes Pod API，不来自 cAdvisor。

Namespace/Pod 状态指标与本轮 cAdvisor 样本合并，由缓存 exporter 暴露。指标采集周期内，cAdvisor 响应按 node 缓存，多个 namespace 不重复请求同一节点。采集或发现失败时清空该 scope 的业务样本并将质量标为失败；源 runtime 样本时间戳容忍一分钟新鲜度；exporter 缓存超时为配置采集周期的两倍，二者是不同检查。资源曲线历史由 Prometheus TSDB 保存，数据库不归档样本。

TSDB 保留期默认 30 天。正式存储选型与容量应以目标环境实测序列、增长和故障恢复条件确认；Prometheus TSDB 与 Agent checkpoint 是不同的数据用途，不能互相替代。NFS checkpoint 只保存 Agent 游标，不是 TSDB 后端。

## 4. 自动日志关联

Monitor 发现带 `instanceId` 标签的业务 Pod/容器后，以该值作为 CMDB `app_instance.guid`，通过插件内已有 CMDB 查询客户端请求单条记录：`POST /wecmdb/api/v1/ci-data/query/app_instance`，过滤字段为 `guid`、操作为精确匹配。CMDB 返回的 `deploy_path` 经绝对路径校验，目标目录固定为 `{deploy_path}/logs`，glob 为 `*.log`，初始位置为末尾。Agent 必须将相同节点目录映射到相同容器路径，例如节点 `/var/data` 挂载为 Agent `/var/data`。

Monitor 按 node 汇总当前对象目标并持久化版本配置，Agent 地址通过 Agent 注册自动发现，也可在 Web 覆盖实际抓取 endpoint。自动目标随发现更新；人工日志规则与排除配置在自动 reconcile 时保留。普通扫描不递归并跳过 `_metric.log` 后缀文件。

同一目录另生成 `*_metric.log` 的 `otlp_metric` 目标，不带 ERROR 规则。Agent 从文件尾部的最后一条完整 OTLP JSON 记录转换白名单应用指标。文件不存在时静默无样本；读取、解析或样本过期时提供应用指标诊断状态。应用标签包含 `instance_id`（CMDB guid）和 Pod 信息，使日志指标可与同一监控对象关联。普通日志和应用指标目标均按 Pod/容器身份生成告警/曲线对象。

## 5. Agent 配置、采集与状态恢复

### Agent 注册和下发

Agent 启动后向 `POST /monitor/internal/k8s/log-agent/register` 发送 Bearer Token 认证 JSON：`clusterId`、`node`、`address`、`port`。Agent 不上报 Pod UID；Monitor 以 cluster/node 为注册键，保存上报地址并记录最近心跳，Web 可设置人工 endpoint override。

Agent 周期性请求 `GET /monitor/internal/k8s/log-agent/config?clusterId=...&node=...&version=...`，响应 envelope 为 `{ "config": { "version": 1, "targets": [...] } }`。配置尚未发布时返回 HTTP 204，Agent 保留当前配置并继续轮询；其他非 200 响应仍记录失败状态。实际 target 包含 id、namespace、workloadType/name、podName/podUid、instanceId、container、directory、glob、start、rules 和可选 format。规则包含 id、内部 version、literal/regex 主 pattern 和可选 exclude。路径限定到配置的 Agent root，打开文件时逐层拒绝符号链接越界。

CMDB 解析失败会记录 `Kubernetes log target resolution failed`，包含对象身份和失败原因，不记录 Token 或响应正文；失败对象被跳过，其他对象继续生成目标。配置未发布的 HTTP 204 仅表示等待，不代表采集目标已生成。

注册、拉取和扫描在 Agent 周期循环执行，默认 10 秒；Prometheus 拉取 `/metrics` 是另一条独立链路，读取 Agent 已准备的内存指标，不在 scrape 请求中解析文件。有效配置超过一分钟未成功更新时，Agent 在后续循环检查点暂停采集，readiness 与 metrics 返回 503；这不是精确的一分钟暂停 SLA。`/healthz` 只表示 HTTP 进程存活。

### 游标和故障语义

Agent 按设备号/inode 维护文件游标，默认新目标从文件尾开始；同路径换 inode 从头读，truncate 回退归零。仅处理完整行，普通行上限 64 KiB、OTLP 单记录上限 1 MiB。成功扫描后回收不匹配文件游标并汇总计数，避免轮转 checkpoint 无限增长；不再匹配 glob 的旧轮转文件尾部不保证补读，copytruncate 存在不可避免的漏读窗口。

Checkpoint 用 flock、临时文件、文件同步与原子重命名写入，避免同一目录的多个进程同时写。扫描先推进内存状态，再尝试持久化；state 损坏时从空状态恢复，写入失败时保留内存游标并重试，不回读旧文件。新建默认目标从日志尾部接管。NFS 故障时文件调用也可能阻塞，不能保证扫描永不中断；进程在内存状态恢复写盘前重启，可能漏计、重置计数或重复计数，丢失/重复窗口由未持久化时长决定，不能固定为一个 10 秒周期。

## 6. 指标与 Web 查询

`k8s_log_match_total` 的规则维度含 cluster_id、namespace、workload、pod、pod_uid、container、instance_id、target 和 rule。Agent/节点状态指标只带 cluster_id/node；健康、缺文件和读取错误指标按不同用途携带目标标签，并非所有 Agent 指标都包含完整对象标签。应用指标另有白名单约束的 `k8s_app_*` 序列。禁止将日志正文、文件路径和匹配文本作为指标标签。

Monitor `/monitor/api/v1/k8s/metrics/catalog` 提供当前 UI 查询白名单；`POST /monitor/api/v1/k8s/metrics/history` 查询历史；`GET /monitor/api/v1/k8s/objects` 和 `/alerts/states` 提供对象与告警视图。Web 在“监控对象”统一展示资源、日志、应用曲线和当前告警状态。字节类 counter 以 rate 语义显示；块设备读取粒度是 bytes，不提供 IOPS/latency。

图表沿用原有数量和布局。图例及悬浮提示只展示业务维度，完整标签仍参与曲线标识，Pod 重建和不同来源的曲线独立保留；同名显示项增加编号。提示框限制宽高，长文本换行，多曲线时内部滚动。单图可放大并展开查看完整标签，复用当前数据及时间范围，不另发历史查询。

普通日志默认规则匹配不区分大小写的 ERROR、FATAL、PANIC。用户可在 Web 为规则增加排除条件；只有主规则命中且没有任何排除项命中时才计数。规则变更使用新计数版本，不追溯重算历史日志或告警。

## 7. 告警与通知

告警规则由 Monitor 周期评估资源和日志计数，并检查样本质量与当前 Pod UID；状态为 `normal/pending/firing/recovered/unknown`，没有独立的 recovering 或 resolved 状态。相同规则和可读对象名称组合形成事件身份；Pod UID 限制样本来源，但同名 Pod 重建不会强制拆成新告警，状态机按名称延续判断。幂等键使用固定长度 SHA-256。规则状态与通知任务通过数据库事务/outbox 持久化，数据库主键/唯一键提供并发兜底；Dispatcher 通过 lease 和有限重试发送 SMTP。邮件系统无法提供与数据库共享事务，超时或回执丢失时可能重复投递。SMTP 服务器和接收策略仍需人工配置。

Namespace 首次登记自动补齐四条默认规则，重复登记不覆盖已存在规则：`default-log-error`（log_matches > 0）、`default-not-ready`（ready < 1）、`default-oom`（oom_events > 0）、`default-cpu-throttled`（cpu_throttled > 0.1 s/s，持续 300 秒）。CPU 限流指标是最近 `4 × 采集周期` 内的限流时间增长速率，不是限流周期占比或 CPU 使用率；默认阈值用于过滤轻微限流与短时突发，仍应按业务负载调整。默认恢复条件是连续 3 次有效正常评估；从 unknown 回来时先恢复有效性，再累计正常周期。日志/OOM 的增量窗口为 `max(300 秒, 4 × 采集周期)`。没有默认 CPU/内存百分比或 JVM/连接池阈值告警，用户可按业务阈值增补。

已有 CPU 限流规则不随升级覆盖，可手工执行[旧默认参数调整 SQL](development/k8s-default-cpu-throttled-once.sql)。脚本只更新仍为旧默认参数且没有对象筛选条件的同名规则，保留启用状态和通知策略；每批最多 500 条，可重复执行。SQL 不重置已有告警状态或删除历史事件；启用本地告警评估时，下一轮开始使用新参数，已触发告警按新阈值与原恢复流程处理。

## 8. API、编排与数据表

Kubernetes 页面统一位于一级菜单“Kubernetes监控”（`/k8sConfig`），下设监控源、监控对象、告警规则，页面标题及面包屑使用相同名称，默认进入监控源。原 `/adminConfig/k8sSources`、`/adminConfig/k8sObjects`、`/adminConfig/k8sAlertRules` 地址保留跳转并携带原查询参数。页面 API 权限归属 `MONITOR_K8S_CONFIG`；WeCube 发布时须同步更新 `register.xml` 的菜单注册及服务端 `menu-api-map.json`，超级管理员默认授权，自定义角色须授予新菜单权限。

告警规则页的集群和命名空间使用联动下拉框，仅展示已登记并启用的监控源；切换集群后选择其命名空间并重新查询，清空原范围结果，忽略旧请求的延迟响应。无可用命名空间时提示先登记监控源，列表加载失败时可重试。

新增的 K8s Web API 位于经登录认证的 `/monitor/api/v1/k8s` 下，资源登记提供 PUT/DELETE；内部 Agent 和 Prometheus endpoint 使用独立内部路由及各自的访问控制。WeCube 通过 `POST /cluster/plugin/put|delete` 和 `POST /namespace/plugin/put|delete` 发送平台批量 envelope，逐条返回 guid、callbackParameter 和错误状态。调用方不提交版本号。实际路由以 `monitor-server/api/api.go` 和 `build/register.xml` 为准，主要接口如下：

| 用途 | 方法与路径 |
| --- | --- |
| Cluster/Namespace 注册 | `PUT/DELETE /cluster/source`、`/namespace/source` |
| Cluster/Namespace 列表与 WeCube 编排 | `POST /cluster/list`、`/namespace/list`、`POST /cluster/plugin/{put,delete}`、`/namespace/plugin/{put,delete}` |
| Agent 配置与端点管理 | `GET/PUT/DELETE /log-agent/source`、`PUT/DELETE /log-agent/endpoint`、`DELETE /log-agent/endpoint/override`、`GET /log-agent/list` |
| 告警规则 | `GET/POST /alert-rules`、`PUT/DELETE /alert-rules/:ruleId` |
| 对象、历史、状态 | `GET /objects`、`POST /metrics/history`、`GET /metrics/catalog`、`GET /alerts/states` |
| 内部指标与 Agent 协议 | `/monitor/internal/k8s/metrics`、`/monitor/internal/k8s/log-agent/register`、`/monitor/internal/k8s/log-agent/config` |

新增持久化表为：`k8s_namespace_source`（登记与发现快照）、`k8s_metric_source`（Cluster 指标凭据）、`k8s_log_agent_config`（节点配置）、`k8s_log_agent_endpoint`（注册和覆盖端点）、`k8s_alert_rule`、`k8s_alert_state`、`k8s_alert_outbox`。指标样本在 Prometheus TSDB；没有独立 workload/pod/container 历史实体表。

Web 与 WeCube 均可管理 Cluster/Namespace。数据库迁移当前是 v3.8.0 开发包的 `#@v3.8.0-begin@;` 新增结构，不宣称已在任何已发布的插件版本中部署。

## 9. 未实现范围与联调边界

- 没有 metrics.k8s.io 自动回退；没有块设备 IOPS、I/O latency。
- 没有原始日志集中存储、精确一次匹配计数/SMTP 交付，也不能补读已离开 glob 的旧轮转文件。
- 日志采集依赖 instanceId、CMDB 可用、`deploy_path/logs` 路径同节点映射以及 Agent 可读权限；这些均需目标环境确认。
- 单环境按一个活动 Monitor 实例设计；当前不提供同环境多副本采集租约/HA 保证。
- 单元测试和模板静态检查不能证明真实 RBAC、CMDB 跨插件调用、NetworkPolicy、NAS 锁语义或 Prometheus Service/Pod 网络已通过部署联调。

旧 `kubernetes_cluster` 管理 API、Web 和数据库仍保留，与本 MVP 使用的 `kubernetes_metric_cluster` 是不同功能。旧 `k8s-kubelet-*`、`k8s-cadvisor-*` 抓取任务已停用：升级启动时清理 PVC 中的存量配置，后端不再重新生成。升级验证见 [Kubernetes 旧采集停用与升级验证](k8s-legacy-scrape-upgrade.md)。
