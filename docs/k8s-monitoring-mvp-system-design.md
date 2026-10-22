# Kubernetes 监控 MVP 详细系统设计

> **WeCube 交付约束**：复用本项目 `build/register.xml` 的插件接口、系统参数、资源依赖及菜单授权机制。编排接口遵循平台 `requestId/inputs` 与 `resultCode/resultMessage/results.outputs` 契约，逐条保留 `guid/callbackParameter`；Token 使用平台敏感参数加密和现有 cipher 库。数据库通过插件包 `init.sql` 的版本标记段安装升级，不新增独立注册、认证或部署协议。namespace 仍需人工登记，再由 Monitor 自动发现其资源。设计目标不代表当前代码已全部完成。

本文用于开发与联调，定义模块、数据流、接口和恢复机制；范围决策、授权 YAML 和容量明细以 [评审方案](k8s-monitoring-plan.md) 为准。

## 1. 目标与边界

本设计面向首个“能用、够用”的 MVP，目标是通过开发和内部联调，完成多个 namespace 的 CPU、内存、网络/块设备 I/O、日志关键字监控、告警、Web 配置、SMTP 通知和 WeCube 编排闭环。

### 1.1 MVP 必须支持

- 每个 Monitor 实例绑定一个运行环境（例如 SIT 或 UAT）；人工登记该环境允许监控的多个 namespace，namespace 不自动发现。
- 每个 namespace 独立保存 Kubernetes 对象 API 凭据和监控配置；指标采集凭据按集群独立保存，不把 cAdvisor Token 复制到各 namespace。
- 登记成功后，Monitor 才对该 namespace 执行 StatefulSet、Pod、container 的发现和同步。
- CPU、内存以及首批容器网络/块设备 I/O 指标；默认 30 秒采样目标（可通过 `MONITOR_K8S_METRICS_INTERVAL_SECONDS` 调整为 15～300 秒）；30 天 TSDB 保留，旧数据直接删除。
- 每个集群/环境安全隔离边界一套日志 Agent DaemonSet，每个目标节点一份 Agent，可读取多个已登记 namespace 的白名单日志目录。
- 日志 literal/regex 关键字、文件缺失、读取错误和 Agent 健康监控。
- 日志规则可选 `exclude: [{mode, pattern}]`，整行命中主规则且未命中任意排除条件才计数；默认不排除，Web 可编辑。规则由 Server 保存、Agent 拉取，版本在保存事务中内部递增，自动发现和跨节点迁移保留人工排除配置。详见 [Agent 规则说明](../monitor-agent/k8s-log-agent/README.md#可选排除条件)。
- Web 配置、当前/历史告警、确认、恢复和 SMTP/Postfix 邮件。
- WeCube 批量接口；沿用现有 JSON `inputs/outputs/callbackParameter` 契约。

### 1.2 明确不做

首个 MVP 的 Kubernetes API 只读取人工登记 namespace 的 StatefulSet、Pod 和容器；Service、EndpointSlice、PVC、Events、节点对象和 NAS 深度指标留作后续版本。节点级 cAdvisor 仅用于读取容器指标，不把节点管理纳入监控范围。

应用部署、发布回滚、日志集中存储、全文检索、Kubernetes 节点管理和节点级告警、NAS 后端容量监控、复杂 PromQL、多条件规则、告警静默/抑制/升级、短信/IM、指标 MySQL 归档、Monitor 同环境多实例高可用和多实例锁。cAdvisor 仅用于读取已登记工作负载的容器指标；每个环境默认单个 Monitor 实例运行。

## 2. 总体架构

```text
用户/WeCube
    │ 配置 namespace、规则、通知
    ▼
Monitor Server
    ├─ Namespace Registry（人工登记）
    ├─ Kubernetes Discovery（仅登记 namespace）
    ├─ Kubernetes Metric Collector（kubelet/cAdvisor；metrics.k8s.io 回退）
    ├─ Rule Evaluator / Alarm Service
    ├─ Agent Config Publisher
    ├─ Web / WeCube API
    └─ SMTP Worker
          │ 查询/写入
          ▼
      Monitor TSDB（Prometheus 兼容，30 天）

部署平台
    └─ 一套 DaemonSet/集群与环境安全边界
          └─ 每个目标节点一个 Log Agent
               └─ 只读 hostPath → /metrics
```

Monitor 实例启动时绑定一个环境标识（SIT/UAT 等），只处理该实例人工登记且启用的 namespace。Monitor 的 Kubernetes 指标采集器通过 API Server 代理抓取 cAdvisor，按登记 namespace、Pod、业务容器白名单过滤后暴露给自带 Prometheus 保存到同一 TSDB；未来扩展 cAdvisor 已暴露的指标时增加映射和规则；未暴露的 NAS 延迟等指标需要另接数据源，不能承诺只改映射即可实现。`metrics.k8s.io` 作为没有 cAdvisor 权限时的 CPU/内存回退。Agent `/metrics` 也写入同一 TSDB。部署平台需提供稳定的 Agent `/metrics` 抓取入口；MVP 不假设 Monitor 可以读取节点对象或自动发现 Agent Pod。

## 3. 运行对象和人工登记

### 3.1 NamespaceRegistry

用户通过 Web 或 WeCube 提交：`environment`、`clusterId`、`namespace`、API server、CA 引用、对象 API Token 引用、启用状态和配置版本。`clusterId + namespace` 唯一；每个部署的 Monitor 实例固定属于一个环境，environment 不接受客户端任意覆盖。SIT/UAT 分别使用独立凭据、MySQL 数据库和 Prometheus 数据目录。保存 Token 的密文/引用，不回显。cAdvisor 集群 Token 在集群级凭据表中单独保存。

状态：`DRAFT → VALIDATING → ACTIVE`，失败为 `DEGRADED`；禁用为 `DISABLED`。enabled 与健康状态分开：ACTIVE/DEGRADED 均执行有限退避重试；首次验证成功前不开放业务采集。禁用/删除先撤销本地指标白名单、清除缓存并停止规则，再异步撤销 Agent 目标，跟踪 observed version；旧 Agent 样本仍须在入库前拒绝。历史按保留策略保存，默认业务查询只允许当前登记范围。

namespace 不读取集群级 Namespace 列表，也不根据 Pod 或发布事件自动创建。SIT 的 Monitor 只处理 SIT 实例登记的 namespace，UAT 同理；即使两个环境位于同一集群，也必须分别登记，未登记 namespace 的指标直接丢弃并记录过滤计数。

### 3.2 DiscoveryService

对每个已启用且验证过的 namespace 独立运行同步任务（DEGRADED 保持重试）。Monitor 启动或进程重启后，先从 MySQL 恢复 namespace、规则、游标和最后观察状态，再对每个 ACTIVE/DEGRADED namespace 执行一次全量 `list` 重扫，使用返回的当前 `resourceVersion` 重建对象索引、运行实例、节点抓取目标和日志目标，完成重扫后才建立 watch。重扫结果会标记已不存在的对象为非活动并撤销其采集目标；因此不依赖进程停止期间未收到的旧 watch 事件。

1. `list` StatefulSet 和 Pod，并读取容器、状态、资源 request/limit 所需字段；（Service、EndpointSlice、PVC、Events 不在 MVP 同步）
2. 建立 StatefulSet → Pod → container 关系；
3. 保存 UID、resourceVersion、ordinal、phase、ready、restart、request/limit 和容器信息；
4. 启动 namespace 范围内的 watch；
5. 处理 410、429、5xx、网络断开和权限错误；
6. 每 5 分钟全量校准一次，清理已删除对象的活动状态但保留历史；对象更新、删除、重建和 nodeName 变化都触发指标目标与日志配置重算。watch 断连、410、进程重启或全量校准发现差异时，按“重新 list → 原子替换当前索引 → 建立新 watch”处理。

Kubernetes 网络调用不放在数据库事务中；对象字段使用白名单，不保存完整 annotations、env 或 Secret。

### 3.3 采集启动与对象生命周期

namespace 注册是唯一的人工采集入口：注册接口保存环境、集群、namespace 和凭据引用，验证通过且启用后才启动该 namespace 的同步任务。Pod、StatefulSet 和容器不要求逐个调用接口添加。DiscoveryService 首次 `list` 后建立 namespace 范围的 watch，并周期性全量校准；发现的对象自动进入指标索引、cAdvisor 过滤范围和日志目标。

Monitor 业务对象按名称识别，Kubernetes UID 只作为运行实例维度；`resourceVersion` 只用于 watch 续接：

```text
逻辑监控对象键：
cluster_id + namespace + workload_type + workload_name + pod_name + container_name

运行实例维度（不作为业务对象主键）：
workload_uid + pod_uid + container_id
```

对 StatefulSet 应用，同一 Monitor 数据边界内、同一集群和 namespace 作用域内名称键相同就视为同一个逻辑监控对象，CPU/内存趋势、资源规则和 Web 展示保持连续。SIT、UAT 使用独立 Monitor 部署及独立存储边界，因此不需要把部署实例 ID 再放入业务键；`cluster_id` 用于区分 Monitor 可能接入的多个 Kubernetes 集群。`workload_uid`、`pod_uid` 和 `container_id` 只用于识别 replace、驱逐、容器重启、Counter 重置、日志游标切换和审计。

同名 StatefulSet 或 Pod 删除重建时，逻辑对象默认连续；原始 Counter 不能跨运行实例直接相加，需按新 `container_id`/Counter 起点重新计算速率。日志游标按文件身份和运行实例安全切换。若业务确实需要把同名重建视为新的监控对象，必须停用旧配置或引入显式业务版本/实例标识；不能仅依赖 UID 变化覆盖本方案的名称语义。

处理规则如下：

- Pod/StatefulSet 原对象更新时，UID 保持不变。Monitor 更新规格、状态、容器、节点和规则关联，继续使用同一对象的告警与历史。
- 如果 replace/发布流程实际执行 delete + create，新对象会获得新 UID；名称键不变时继续关联同一个逻辑监控对象，同时更新运行实例维度、重新计算 Counter 和切换日志目标。
- StatefulSet 滚动更新通常保留 StatefulSet UID，但被替换的 Pod 会获得新 UID。StatefulSet 级聚合、槽位趋势和逻辑对象告警保持连续；原始 Counter 样本和日志游标按运行实例 UID 分开，逻辑查询再按名称键聚合。
- Pod 驱逐或删除后由控制器补建时，同名新 Pod 通过 watch 或重启后的全量重扫自动发现并继续关联逻辑监控对象；新 Pod 可能位于另一节点，Counter 按新运行实例重新计算，日志目标切换到新节点。

Pod/容器对象事件触发两个重算：观察到 `spec.nodeName` 变化或新 Pod UID 出现时（主机切换通常是后者），MetricCollector 重新生成去重后的节点抓取目标；日志配置生成器重新生成 node/Pod/目录映射并发布新版本。旧节点目标和旧 Pod 日志目标被撤销，新节点上的 Agent 负责接管。Monitor 重启后的全量重扫执行同样的重算，因此停止期间发生的 Pod 重建、扩缩容、节点切换和容器变化会自动补回当前状态；停止窗口内未采到的指标和旧节点不可读的日志仍需报告数据缺口。前提是部署平台已经在新节点运行同一环境的 Agent，并且新节点 hostPath、UID/GID/ACL 和配置发布通道可用。

## 4. 指标采集与 TSDB

### 4.1 MetricSource 接口

```go
type MetricSource interface {
    Query(ctx context.Context, target MetricTarget, query MetricQuery) ([]MetricSample, error)
}
type MetricTarget struct {
    MonitorInstanceID string
    Environment string
    ClusterID string
    Namespace string
}
type MetricSample struct {
    Name string
    Value float64
    Timestamp time.Time
    Labels map[string]string
    Quality string // GOOD, STALE, UNKNOWN
}
```

MVP 逻辑指标（CPU/内存/网络/块设备 I/O 来自 cAdvisor；request/limit、restart、ready 和副本数来自 Kubernetes 对象 exporter；采集器过滤后的原始/派生序列写入 TSDB，逻辑查询不重复写一份同名序列）：

```text
k8s_pod_cpu_cores
k8s_pod_memory_working_set_bytes
k8s_pod_cpu_request_cores
k8s_pod_memory_limit_bytes
k8s_pod_restart_total
k8s_pod_ready
k8s_statefulset_ready_replicas
k8s_pod_network_receive_bytes_total
k8s_pod_network_transmit_bytes_total
k8s_pod_block_read_bytes_total
k8s_pod_block_write_bytes_total
```

MetricCollector 按默认 30 秒周期（可通过 `MONITOR_K8S_METRICS_INTERVAL_SECONDS` 调整）通过 API Server `/api/v1/nodes/{node}/proxy/metrics/cadvisor` 抓取，使用独立集群级 Token；通过已登记 Pod 的 `spec.nodeName` 生成节点目标。采集器先按环境、namespace、Pod 名称/UID 映射和业务容器过滤，再将允许的样本暴露给 Monitor Prometheus。cAdvisor 标签未提供可验证 Pod UID 或发生同名重建无法消歧时，样本不得关联到新 Pod；采集器应等待对象索引完成并报告映射异常。cAdvisor CPU counter 使用 `rate()`，内存和 I/O 字节等按指标类型处理。`metrics.k8s.io` 适配器只在回退时读取当前 CPU/内存，不承担扩展指标。

### 4.2 内部采样链路

```text
MetricCollector（按节点抓取并按环境/namespace 过滤）
    → MetricCache（最近一次采样和 quality）
    → monitor-k8s-exporter /metrics
    → Monitor Prometheus（Kubernetes 两个 scrape job 默认 30s；跟随 `MONITOR_K8S_METRICS_INTERVAL_SECONDS`）
    → RuleEvaluator / Web query
```

最近成功值只用于显示最后已知状态，失败或过期值不得继续以新鲜业务样本暴露给 Prometheus；同时将 quality 置为 `UNKNOWN`；连续 3 个采样周期失败后产生数据源告警。恢复后记录恢复时间。Monitor 自带 Prometheus 对 `k8s_monitor` 和 `k8s_log_agent` 的默认抓取周期均为 30 秒，并与 cAdvisor 轮询共用 `MONITOR_K8S_METRICS_INTERVAL_SECONDS`（允许 15～300 秒）；修改参数后需重启 Monitor，启动脚本会更新持久化的 Prometheus 配置。Agent 文件扫描和配置拉取仍为独立的 10 秒周期。TSDB 启动参数使用 30 天 retention，旧块直接删除；不写入 MySQL。

TSDB 存储要求：正式部署使用独立本地盘或受支持的块存储 PVC，不使用 NFS/NAS PVC。当前集群的 `topolvm` StorageClass（`topolvm.cybozu.com`、XFS、`WaitForFirstConsumer`、支持扩容）是优先候选，但上线前必须确认目标节点的 device class、可用容量、调度约束和节点故障恢复策略，并确认 `Delete` 回收策略符合生产数据保留要求。原因是 Prometheus TSDB/WAL 需要稳定的本地文件系统追加写、`fsync`、`mmap`、rename、锁和崩溃恢复语义；网络文件系统即使容量、带宽和 IOPS 达标，也可能在网络或服务端异常时造成 WAL/Head/compaction 阻塞、写入失败、数据损坏或恢复不完整。当前 NAS 仅允许开发或临时联调，不能作为满足 30 天可靠保留目标的生产存储；若平台只能提供 NAS，必须走书面风险豁免并接受指标历史丢失或损坏风险。

Collector 的单次请求设置超时、大小和并发上限；保留原始 Counter 类型及 container_id、device/interface 维度，在 PromQL 中先 rate 再聚合，不能先合并可能独立重置的 Counter。CPU 和网络/块设备 I/O 速率窗口默认为抓取周期的 4 倍（默认 120 秒），至少需要两个有效样本；日志匹配和 OOM 计数窗口取 `max(300s, 4×抓取周期)`，日志健康新鲜度为抓取周期的 3 倍。索引中的 workload UID、ordinal、Pod UID、container 名称、containerID 与 cgroup id 对照，不仅凭同名 Pod 关联。采样失败立即停止输出对应业务样本，质量另行输出；同时检查节点抓取健康与业务样本新鲜度，不能只看 exporter up。

### 4.3 Token、采样时效与 cAdvisor 边界

每个登记 namespace 使用独立 ServiceAccount + Role + RoleBinding 读取 core `pods`、apps `statefulsets`；只有启用回退时才增加 metrics.k8s.io `pods` 的 get/list。cAdvisor 使用单独集群级身份读取 `nodes/proxy`，不与 namespace Token 混用。完整授权 YAML 和签发命令见 [评审方案第 6 节](k8s-monitoring-plan.md#6-kubernetes-提供方交付要求)。RBAC 的 apiGroups 不包含版本号。容器数据读取自 Pod，而非独立 containers API。

本次用户使用 admin 验证 Metrics API 和所测节点 cAdvisor 端点可用；不代表新建 Monitor Token 有节点代理权限。对象 API Token 的验收使用其实际身份读取已授权 namespace，并验证其他 namespace 拒绝访问；cAdvisor Token 则验收节点代理访问、采集器的环境/namespace 过滤和未登记样本不入库。无需反复检测原业务 Token。

采集周期是轮询/抓取目标，不保证源按相同周期更新。仅对 Metrics API 回退：保留源 timestamp/window，CPU quantity 转 cores、memory 转 bytes，CPU 是源窗口内平均用量；不能对该 CPU Gauge 再使用 rate()。缓存同时输出源时间和数据质量，按源更新周期确定 stale 阈值；未知更新周期先通过重复读取 timestamp 测量。Pod UID 从对象索引关联，源时间早于新 Pod 创建时间的数据不得关联给新实例。request/limit、restart、ready 从 Kubernetes 对象取得。空值不是零，initContainer 不进入业务容器统计。

主路径 cAdvisor 直接抓取需独立集群级 `nodes/proxy get` 身份；从已登记 Pod 的 nodeName 生成按集群/节点去重的抓取目标，不额外需要 nodes list/watch。GET proxy 不能按后缀路径限制为仅指标，也不是 namespace 隔离的只读授权；因此必须在采集器中先做环境/namespace/Pod 名称与 UID 映射/容器过滤，过滤前数据不得进入 TSDB、规则或接口响应。该过滤不是 RBAC 替代，而是满足 SIT/UAT 人工范围边界的应用层控制。

cAdvisor 按指标族过滤：CPU、内存和块设备 I/O 仅保留 Pod spec 中的普通业务容器，排除根 cgroup、Pod 汇总和 initContainer。网络通常属于 Pod 共享网络命名空间，可能只有空 container 或 sandbox 样本；网络应按 Pod UID/interface 选取唯一来源，不使用业务容器非空过滤，不对每个容器重复归属。hostNetwork Pod 无法可靠区分业务流量时标为不支持。对网络和块设备 I/O 指标必须先核对目标集群实际暴露的 metric family、标签和重启语义，再决定映射；端点中存在其他指标不代表这些指标已经可用。CPU Counter 使用 rate() 后先按容器身份核对再聚合；内存直接读取 working set。样本 id/容器重启处理和 CPU/memory 覆盖需专门验证，不能仅凭端点有返回就认定完成。提供方不需要额外提供 Prometheus/VictoriaMetrics。

## 5. 日志 Agent 系统设计

### 5.1 部署拓扑和职责

每个集群/环境安全隔离边界一套 DaemonSet，在每个目标节点运行一个 Agent Pod，不按 namespace 重复部署。Agent 配置可包含多个 namespace；配置中的每个目标必须同时通过 namespace 白名单和 hostPath 白名单校验。发生 hostPath、凭据或运维隔离要求时才拆分 DaemonSet。

Monitor 负责镜像、配置 schema、指标协议、健康检查、资源建议和兼容版本；部署平台负责 DaemonSet 安装、升级、回滚、ConfigMap/Secret 挂载和节点运维。Monitor 不调用 Kubernetes 创建 DaemonSet。若同一集群同时承载 SIT 和 UAT，部署平台按环境安全边界提供独立 Agent 配置和抓取入口，避免两个 Monitor 互相读取对方指标。

Agent 启动和周期刷新配置时，主动向 Monitor 上报自身可达地址、监听端口、节点名和 Agent Pod UID；Monitor 自动维护该节点的抓取目标并在地址变化后刷新 Prometheus file_sd。管理员手工修改地址时设置 `manual_override`，后续自动上报只更新存活时间和 Pod UID，不覆盖手工地址；Web/运维更新接口将 `manualOverride=false` 即可取消手工覆盖并恢复自动更新。这样部署 Agent 时只需配置 Monitor 地址和认证信息。

Monitor 同时为自动生成的普通日志目标创建同目录 `*_metric.log` 目标，格式为 `otlp_metric`，不附带关键字规则。Agent 从最新完整 OTLP 记录提取白名单内应用指标并暴露为 `k8s_app_*`，通过同一个 Prometheus job 保存，使用同一 Pod/容器和 `instance_id` 关联。普通日志扫描排除该文件；应用指标没有新鲜样本时独立标记不可用。类型、时间、容量和属性维度约束见 [应用指标白名单](k8s-app-metric-whitelist.md)。

### 5.2 Agent 内部模块

```text
ConfigManager
  ├─ SchemaValidator
  ├─ VersionStore
  └─ AtomicReloader
        ↓
TargetResolver（namespace/workload/pod/container → 白名单路径）
        ↓
FileDiscovery（glob，禁止越界符号链接）
        ↓
TailManager
  ├─ FileIdentity（device/inode/path）
  ├─ RotationHandler（rename/copytruncate/truncate）
  └─ CursorStore（独立可写目录）
        ↓
Matcher（literal / regex）
        ↓
WindowAggregator（计数、最近匹配、读取状态）
        ↓
MetricsEndpoint + HealthManager
```

每个文件一个受限读取任务；匹配在单文件顺序执行，配置 reload 使用读写锁交换 immutable rule set。队列达到上限时停止继续读取并将 `agent_read_lag_seconds`、`agent_queue_full_total` 置为异常，不能静默丢弃。

### 5.3 配置协议

交付给 Agent 的配置使用英文 JSON/YAML，版本化并带 `configVersion`；传输方式见第 6 节。Monitor 解析对象索引和经确认的日志路径模板后生成具体 node/Pod/目录映射，Agent 不自行查询 Kubernetes，也不使用任意 label 猜测实例。窗口/阈值归 Monitor 规则，Agent 仅计数。示例：

```yaml
configVersion: 12
clusterId: cluster-a
namespaceTargets:
  - namespace: sit-a
    rootAllowlist: [/var/log/apps/sit-a/order-service-0]
    targets:
      - workload: order-service
        workloadUid: workload-uid
        node: node-a
        pod: order-service-0
        podUid: pod-uid
        targetId: target-a
        container: app
        files:
          - glob: "*.log"
            encoding: UTF-8
            startPosition: end
            missingAfterSeconds: 60
        rules:
          - ruleId: order-error
            mode: literal
            pattern: "ERROR"
            ruleVersion: 1
```

Monitor 将 CMDB `app_instance.deploy_path` 解析为 `{deploy_path}/logs` 作为目标目录；例如部署目录 `/var/data/SIT_APP_INSTANCE` 对应 `/var/data/SIT_APP_INSTANCE/logs`。Agent 校验：路径必须是 `rootAllowlist` 子路径、禁止 `..` 和目录外 symlink；glob 不得为空；regex 必须可编译；文件数、regex 长度、规则 ID、label 长度有上限；同一 `ruleId` 在同一目标内唯一。校验失败不替换当前生效配置，并返回错误详情。

### 5.4 文件状态机

```text
DISCOVERING → TRACKING → ROTATED → TRACKING
      │           │          │
      ├─ missing ─┘          └─ re-scan
      └─ permission/error → ERROR → retry
```

- 初次读取默认从文件尾部开始，避免一次性处理历史日志；规则可显式要求从头读取。
- 以 inode/device 和路径组合识别文件；rename 后继续跟踪旧 inode，并扫描新文件。
- copytruncate 检测文件大小回退，游标归零后继续读取；截断与检测之间仍可能丢失数据，不能承诺零漏读。rename 后的新文件从头开始，首次接管文件默认从尾开始，两者不可混用。
- 每个 checkpoint 原子保存文件身份、已完成完整行的 offset、ruleVersion 和 Counter 状态：临时文件写入、fsync、rename，并同步父目录。仅将已提交 checkpoint 的计数对外暴露；重启重读 checkpoint 后数据，避免计数与游标不一致。并发抓取读取不可变快照。
- 文件删除或暂时不可见不立即恢复业务告警，超过 `missingAfter` 才产生文件缺失告警。
- Pod 被删除后停止输出业务序列，保留游标 24 小时后清理；不同 Pod UID、规则版本或不同节点不得复用旧游标。节点丢失本地日志不可恢复，应暴露数据缺口。
- 多文件规则计数聚合，missing 代表期望 glob 无可读文件；单文件错误独立计入 error，不能因另一文件正常而静默忽略。长行、部分行、编码失败和文件扫描深度必须有限制。

### 5.5 Agent 指标和健康

```text
k8s_log_match_total{cluster,namespace,workload,pod_uid,container,rule}
k8s_log_file_missing{cluster,namespace,workload,pod_uid,container,rule}
k8s_log_read_error_total{cluster,namespace,workload,pod_uid,container,rule}
k8s_log_last_match_timestamp{cluster,namespace,workload,pod_uid,container,rule}
k8s_log_last_read_timestamp{cluster,namespace,workload,pod_uid,container,rule}
k8s_log_agent_up{cluster,node}
k8s_log_agent_config_version{cluster,node}
k8s_log_agent_read_lag_seconds{cluster,node}
k8s_log_agent_queue_full_total{cluster,node}
```

每个新规则先输出零基线，规则版本变化启用新序列；首个样本前的增量无法由 Prometheus increase 精确还原，需标记预热期。归属不明确时使用 workloadUid + targetId，pod_uid 为空，不伪造 Pod；标签补充 environment、targetId、ruleVersion。`match_total` 为 Counter，规则使用 `increase(k8s_log_match_total[window])`；`file_missing` 为 Gauge；时间戳为 Gauge。禁止文件名、完整路径、关键字正文作为 label。Agent 暴露 `/metrics`、`/healthz`、`/readyz`，不返回日志正文。

### 5.6 安全和资源

Agent 使用非特权、无 hostNetwork/hostPID/hostIPC、禁止提权和多余 capabilities；hostPath 只读，游标目录单独可写。通过 UID/GID/ACL 确认可读性，不依赖 root。设置 CPU、内存、文件句柄和最大文件/规则/队列限制。异常记录英文堆栈并退出，由 Kubernetes 重启；配置和游标保证重启后恢复。

## 6. Agent 配置发布

1. 用户在 Web/WeCube 提交规则；
2. Monitor 校验并写入 desired config/version；
3. Monitor 生成 namespace 合并后的 Agent 配置和 checksum；
4. Monitor 通过部署平台约定的 ConfigMap 发布接口或配置下载接口交付；
5. 部署平台更新 DaemonSet 挂载内容；
6. Agent watch 文件并 reload；
7. Agent `/readyz`、config version 和指标被 Monitor 查询确认；
8. 失败时 Agent 保留旧版本，Monitor 状态为 `DEGRADED`。

MVP 必须在开发第一天确定第 4 步采用一种方式：部署平台 API、共享 ConfigMap 或 Agent 拉取 Monitor 配置。Monitor 不假设自己拥有创建 ConfigMap 的权限，也不在 MVP 同时实现多种发布通道。

## 7. 规则和告警

### 7.1 规则

资源规则支持 CPU cores、memory bytes、网络/块设备 I/O bytes/s、request/limit 百分比；日志规则支持窗口内匹配次数。第一版每条规则只有一个指标、比较符、阈值、持续时间和通知策略。

资源、I/O 和日志评估周期跟随采集周期，默认 30 秒；I/O 速率查询至少需要两个有效 Counter 样本，首次样本处于预热状态。进入 Firing 前连续满足条件达到 `forSeconds`；连续 3 个正常周期恢复。数据 quality 为 UNKNOWN 时进入 Unknown，不自动恢复业务告警；数据源恢复后继续评估。MVP 不实现静默、抑制和升级。

实际评估按 namespace 在同一轮复用相同指标的 Prometheus 查询结果，下一轮重新查询，查询超时跟随当轮 context。资源发现或 watch 失败时，保留对象身份用于恢复，但停止导出旧的对象状态及资源样本，告警进入 Unknown；重新发现成功后恢复。Phase 仍作为带 `phase` 标签的指标提供，不直接用于单值阈值告警；对象健康告警使用 `ready` 或 `container_ready`。

### 7.2 唯一键和状态

```text
cluster_id + namespace + workload_type + workload_name + pod_name + container_name + rule_id + normalized_labels + incident_generation
```

数据库唯一约束保证单实例重复评估不会重复创建。评估状态、处置状态和通知状态分开保存。Monitor 默认单实例，不实现多实例锁；通知任务使用 MySQL 持久化状态恢复；SMTP 仍存在重复发送窗口。当前状态单独一张表按对象/规则键唯一，历史按 generation 唯一；事务内锁当前状态行并插入事件/outbox，确保并发评估不会创建两个事件。

### 7.3 SMTP

复用现有 SMTP 发送器的连接配置；触发/恢复邮件先持久化告警记录，再写入 outbox。发送适配器明确返回结果，成功才标记 `sent`，失败按指数退避重试，默认最多 5 次，超过后标记 `failed` 并保留错误。SMTP 已接受后进程在状态更新前崩溃时可能重复发送，按至少一次投递设计，不能承诺 exactly-once。现有发送器的 TLS 校验和认证行为需要在 MVP 中按部署环境适配，不能直接沿用跳过证书校验的配置。SMTP 密码使用凭据引用或加密值，不回显。MVP 支持 Postfix 无认证中继、AUTH 和 TLS。

通知派发使用独立于当轮告警评估的超时预算，评估失败不阻塞已提交的 outbox。租约覆盖发送预算及单次 SMTP 操作的缓冲时间，取消后停止发送剩余事件；直接 SMTP 链路的连接、TLS 握手及 SMTP 命令共用 10 秒期限。通知对象身份包含 cluster、namespace、workload type/name、pod 和 container，避免跨集群或容器串告警；超长身份使用固定长度哈希，完整对象信息保留在告警内容中。

## 8. Web 与 WeCube 接口

Web 页面：namespace 注册、资源规则、日志规则、Agent 状态、SMTP 配置、当前/历史告警和告警确认。页面使用 i18n，API 字段使用英文 camelCase。

WeCube 沿用仓库现有 JSON 批量结构（参考 `monitor-server/api/v1/agent/kubernetes.go`）。MVP 的业务接口统一使用 JSON；只有实际 WeCube 通道明确要求 XML 时，才增加不改变业务字段的 envelope adapter：

```json
{
  "requestId": "request-id",
  "inputs": [{"guid": "rule-id", "callbackParameter": "callback", "namespace": "sit-a"}]
}
```

响应示例：

```json
{
  "results": {"outputs": [{"guid": "rule-id", "callbackParameter": "callback", "errorCode": "0", "errorMessage": ""}]},
  "resultCode": "0",
  "resultMessage": ""
}
```

上例仅说明 envelope；新增字段和版本需同步 DTO、Web 及 build/register.xml。现有注册描述中的 XML 是接口声明，不代表运行时 XML。现有 handler 会记录完整请求且非 add 分支进入 delete，复用前必须脱敏并严格枚举 action。

MVP 接口（相对 `/monitor/api/v1`；读查询无业务写副作用）：

| 路径 | action | 作用 |
|---|---|---|
| `/agent/export/kubernetes/cluster/:action` | add/update/delete | 改造现有入口；update 必须新增显式分支，禁止走 delete 分支 |
| `/k8s/monitor/config` | apply/validate | 批量写资源、日志、通知配置 |
| `/k8s/monitor/config/status` | query | desired/observed version、Agent 和错误 |
| `/k8s/alarm/list` | query | 当前/历史告警分页查询 |
| `/k8s/alarm/ack` | update | 幂等确认 |
| `/k8s/mail/test` | execute | 测试邮件 |
| `/k8s/monitor/status` | query | namespace、指标、Agent、规则和 SMTP 状态 |
| `/k8s/alert-rules` | GET/POST/PUT/DELETE | Web 规则列表、新增、按版本更新和删除 |
| `/k8s/objects` | GET | 已启用 namespace 的对象清单及关联告警 |
| `/k8s/metrics/catalog` | GET | 查询指标白名单、单位、作用域和采集周期 |
| `/k8s/metrics/history` | POST | 按对象名称查询历史，支持资源、日志及应用指标；不限定当前 Pod UID |

写操作使用 `requestId + guid + version` 幂等；已达到目标状态返回成功。

## 9. 数据模型和一致性

核心表：`k8s_namespace_source`、`k8s_workload`、`k8s_pod`、`k8s_container`、`k8s_metric_rule`、`k8s_log_rule`、`k8s_agent_config`、`k8s_alarm`、`k8s_notify_task`。复用边界见第 11 节。namespace 唯一键为 `cluster_id + namespace`；逻辑监控对象唯一键使用 `cluster_id + namespace + workload_type + workload_name + pod_name + container_name`；运行实例表单独保存 `workload_uid + pod_uid + container_id` 及起止时间，并记录实例起止时间；规则使用 `cluster_id + namespace + guid`；告警活动唯一键使用逻辑监控对象键 + `rule_id + normalized_labels`；历史告警使用 `incident_generation` 区分恢复后的新事件；通知任务使用 `alarm_id + transition + channel`。

配置写入顺序：校验 → 同一 MySQL 事务写 desired version 和 outbox → Worker 提交外部配置 → 外部确认 → 条件更新 observed version。崩溃后扫描 pending/超时任务重试；按目标+版本+checksum 幂等，旧版本不得覆盖新版本。外部失败保留已提交 desired 和上一生效配置，不假装跨系统回滚。告警状态变更和通知 outbox 也必须同事务。SMTP 重试可能重复发送，稳定 Message-ID 仅辅助追踪。

## 10. 容量与部署基准

安装后的最小闭环是：登记 Cluster、登记 Namespace、部署 DaemonSet；其余 Pod/StatefulSet、指标目标、CMDB 日志目标、默认 ERROR/FATAL/PANIC 日志规则和基础告警由 Monitor 自动发现与协调。SMTP 和接收人仍由运维按需配置；未配置 SMTP 时告警状态仍保存，通知不会发送。

容量公式、序列估算、资源申请表和 Prometheus 存储约束统一维护于 [评审方案第 8.1 节](k8s-monitoring-plan.md#81-mvp-初始容量结论评审基准)，不在本文另列一套数值。实施时按当前环境全部已登记 namespace 汇总，并在联调后以活跃序列数、磁盘增长和日志吞吐校准。

## 11. 仓库复用检查与实现约束

| 仓库依据 | 可复用范围与必须修改项 |
|---|---|
| `build/conf/kubernetes_prometheus.tpl` | 可参考 cAdvisor 路径；原模板依赖节点发现并跳过 TLS 校验，不能直接套用本方案权限/过滤 |
| `monitor-server/api/v1/agent/kubernetes.go`、`build/register.xml` | 复用 JSON 批量 envelope；新增 update 分支、凭据引用、错误码和脱敏，保持旧 add/delete 兼容 |
| `monitor-agent/node_exporter/collector/monitor_log_keyword_linux.go` | tail/regex 可参考；当前 Gauge、末条原文和大队列不能满足新 Counter/checkpoint/无原文契约 |
| `monitor-server/services/other/mail.go` | 现有发送函数无返回错误，固定端口和 TLS 行为需改造；outbox 不能只靠日志判成功 |
| `build/start.sh`、`monitor-server/services/db/dashboard.go` | 已有 30d retention 配置；新指标查询/任务需隔离旧 archive_mysql 链路，不能全局删除旧业务归档 |

按评审清单落实：外部 SDK/网络错误使用实际包装类型及 errors.Is/errors.As 判断 timeout、取消、认证和重试；日志记录脱敏堆栈而非原始响应体。数据库唯一约束与版本条件更新是并发兜底，MySQL 8 迁移须幂等。API server 仅允许配置的 HTTPS 目标，CA 校验、禁止非预期重定向、使用 URL builder/PathEscape，namespace/name 先按 Kubernetes 格式校验；PromQL 仅允许参数化模板，禁止任意查询拼接。路径按白名单解析且打开时防 symlink 越界，凭据不得进入日志、label 或响应。

SIT/UAT 各部署一套独立控制面，不是同一控制面的 HA 多副本。每环境一个活动进程；单进程任务有互斥/有界并发，数据库唯一约束防重入，重启恢复 outbox 和 checkpoint。同环境 HA 尚未实现，部署须禁止重叠活动实例；若启用 HA，必须先增加数据库租约和 fencing，不能把现有单机锁当成多机保护。

## 12. 实施前置条件

集群身份、指标覆盖、日志契约及部署条件见 [评审方案第 10 节](k8s-monitoring-plan.md#10-待提供信息)。开发启动前还需确认：

- 第 6 节的 Agent 配置发布通道及各节点独立抓取目标；
- SMTP/Postfix 认证、端口和 TLS 参数；
- WeCube DTO、注册描述和部署版本的兼容范围。

以上条件尚未全部验证，不将设计文档审查等同于实施条件已经满足。

## 13. 开发工作量与验收

以下是约 14 个工作日的粗估，不是两周自然日交付承诺；以权限、Agent 配置通道、日志映射和存储条件确定后的联调结果调整。

| 阶段 | 结果 |
|---|---|
| 1～2 天 | 设计冻结、namespace/规则/Agent 配置 DTO、WeCube 接口 |
| 3～4 天 | 多 namespace 手工登记、对象发现、指标适配和内部 exporter |
| 5～8 天 | Agent 模块、文件状态机、游标、reload、健康和 DaemonSet 联调 |
| 9～10 天 | 规则评估、告警状态、去重、恢复 |
| 11～12 天 | Web、SMTP、通知任务 |
| 13～14 天 | WeCube、多 namespace 联调、异常演练和文档交付 |

验收：至少两个手工登记 namespace；每个 namespace 独立对象 API 凭据和规则，cAdvisor 使用独立集群凭据；StatefulSet/Pod/容器可发现；CPU/内存及网络/块设备 I/O 有当前值和趋势；并明确块设备 I/O 不替代 NAS/NFS 后端指标；日志关键字可触发/恢复；轮转和 Agent 重启不持续漏报；Web 和 WeCube 可配置；SMTP 触发/恢复邮件可见；对象 Token 失败仅影响对应 namespace；cAdvisor Token 失败影响该实例该集群全部相关节点并统一告警；Monitor 重启后配置、告警和 Agent 状态可恢复。

## 默认日志告警与统一对象入口

Namespace 登记在同一数据库事务中创建 `default-log-error`（`log_matches > 0`），重复登记不覆盖人工修改。Agent 的 ERROR/FATAL/PANIC 是匹配计数规则；Monitor 使用 `increase(k8s_log_match_total[window])` 评估窗口内新增数，其中 `window` 默认为 300 秒，并在抓取周期变长时取 `max(300s, 4×抓取周期)`，不以累计总数直接告警。默认连续 3 个有效正常评估周期恢复；首次尚无足够样本、抓取失败、目标读取异常或数据过期时进入 unknown，不误判恢复。SMTP 与接收策略仍需人工配置。

日志与资源统一使用 `cluster_id / namespace / workload_type / workload_name / pod / pod_uid / container` 标签。日志匹配指标额外保留 `instance_id`（CMDB guid）；UID 隔离运行实例，名称组合维持业务对象身份。Agent 新增 `k8s_log_target_healthy`，当前扫描存在日志文件且所有匹配文件读取成功为 1，否则为 0。

Web 在管理员配置中提供“监控对象”页面，Namespace 行也可直接跳转。支持 cluster/namespace、工作负载、Pod 名称和 instanceId 筛选，统一查看资源、日志、应用指标曲线及当前关联告警状态；支持预设和最长 30 天的自定义时间范围。查询在 Prometheus 端按对象过滤，保留不同 Pod UID 的历史序列，缺失样本留空；自动刷新跟随配置的采集间隔，默认 30 秒。详细操作与指标语义见 [对象指标查看说明](k8s-object-metrics.md)。初始日志读取位置是文件末尾，首次 Prometheus 样本之前的匹配不保证由增量查询捕获；不承诺逐条日志的精确一次告警。
