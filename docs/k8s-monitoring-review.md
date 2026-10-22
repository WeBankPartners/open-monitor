# Kubernetes 监控 MVP 评审方案

## 结论

MVP 可在安装 Monitor 后，通过部署一套日志 Agent DaemonSet、登记 Cluster/Namespace 开始工作。业务 Pod/STS、容器、CPU/内存/网络/块设备字节指标和日志目标自动发现；资源与日志指标进入 Monitor Prometheus/TSDB，告警状态自动评估。无需逐个登记 Pod、节点或 Agent 地址，也无需先配置 SMTP 才能查看告警状态；SMTP 通知需人工配置。

Web 与 WeCube 批量编排接口都可登记 Cluster/Namespace；接口契约见[当前系统设计](k8s-monitoring-system-design.md)。这只确认代码和自动化测试范围，不代表目标环境部署已通过。实际 Kubernetes 权限、CMDB 查询、节点日志挂载、NFS 和网络仍须联调确认。Agent 操作见[部署指南](k8s-log-agent-deployment.md)。

## 1. MVP 范围

**包含：** namespace 范围 StatefulSet/Pod/容器发现；cAdvisor CPU、内存、CPU 限流、网络和块设备读写字节；Pod API request/limit、restart、Ready、Phase、OOM 状态；默认日志关键字计数和人工排除；默认 `_metric.log` OTLP 白名单应用指标；对象曲线；告警状态机、outbox 与 SMTP 通知。

**明确不包含：** `metrics.k8s.io` 回退、块设备 IOPS/latency、原始日志保存、Agent 多副本高可用、跨插件资源自动创建，以及 SMTP 自动配置。I/O bytes rate 不能推断 IOPS 或延迟。

最少人工初始化：

1. 在 Monitor 所在环境应用数据库结构，部署 Monitor；配置 Cluster API Server/凭据并登记 Cluster、Namespace。
2. 将 Agent 镜像推送至环境仓库，替换部署模板中的环境值，apply 一份 DaemonSet/PVC/ConfigMap YAML；节点日志路径与权限必须满足模板前提。
3. 验证自动发现、Agent 在线、Prometheus Targets 和对象曲线。策略、Token、CMDB、路径或存储有问题时先解决前置条件。

后续才需要人工调整的部分：日志规则排除、Agent 对接地址覆盖、自定义告警规则、SMTP 与通知对象。

## 2. 运行前提与可接受风险

- 每个 SIT/UAT 等隔离环境部署一套 Monitor 与一套 DaemonSet；用户手工登记该环境允许处理的 cluster/namespace。Monitor 过滤配置不是 Kubernetes RBAC 的替代品。
- Cluster 保存 API Server、可选 CA 和集群凭据；Namespace 可填写 Token 覆盖，空值继承 Cluster Token，且不复制 Cluster Token 到 Namespace 记录。Namespace 没有独立 API Server/CA。无 CA 时会跳过 TLS 验证，这是已接受的配置选项，存在中间人风险，应优先提供 CA。
- `instanceId` 必须对应 CMDB `app_instance.guid`。Monitor 查询 `deploy_path` 并固定使用 `{deploy_path}/logs`；Agent 与节点上的路径须同路径映射。日志目录/文件权限、轮转和可达节点属于应用/平台前置条件。
- Agent 每节点扫描和配置拉取默认 10 秒；Prometheus 的 Kubernetes 两个 scrape job 与 cAdvisor 采集周期默认 30 秒。30 秒由 `MONITOR_K8S_METRICS_INTERVAL_SECONDS` 调整为 15～300 秒，Monitor 重启后生效。
- Agent checkpoint 可使用环境提供的 RWX NFS PVC，但必须在实际 PVC 上验证锁、文件同步、原子重命名；RWX 不等同于各节点独立盘。NFS 阻塞可能拖住 Agent I/O。state 损坏或无法保存时以内存状态降级并重试，Agent 重启前未持久化的日志可能漏计或重复计数；损失窗口不是固定扫描周期。
- Prometheus TSDB 使用独立于 Agent state 的本地文件系统或合适的块存储卷，不以 NFS 作为正式 TSDB 存储。[Prometheus 官方存储说明](https://prometheus.io/docs/prometheus/latest/storage/#operational-aspects) 明确不支持 NFS，存在不可恢复的数据损坏风险；吞吐或 checkpoint 测试通过不能改变这一支持边界。此约束不否定 Agent checkpoint 使用 NFS。

## 3. Kubernetes 权限

Monitor discovery 按 namespace 使用 API Server 访问 `pods` 与 `statefulsets` 的 `get/list/watch`。指标抓取使用 Cluster 凭据访问 `nodes/proxy` 的 `get`，请求 `/api/v1/nodes/{node}/proxy/metrics/cadvisor`；此权限不能用 URL path 限制为只暴露 cAdvisor，属于高风险集群级权限，应由 Kubernetes 管理员单独审批。Agent 只读挂载日志并写自己的 checkpoint，不需要 Kubernetes API 写权限或 `pods/exec`。

当前代码无 `metrics.k8s.io` 回退，因此不需要为此功能申请该 API 权限。最小权限清单见[权限摘要](k8s-monitoring-k8s-permissions-summary.md)。提供方签发/轮换 Token；Monitor/Web 用户不需要具有创建 Role 或 Token 的权限。

## 4. 容量评审基准

评审规模为约 40 节点、每节点约 10 个 Pod，即约 400 个 Pod。

设同时活跃的序列数为 S，30 天样本数量为 `S × 30 × 86400 / interval_seconds`；30 秒时每条序列为 **86,400 个样本**，60 秒为 43,200 个。此公式不含索引、WAL、Head 和压缩开销，也不等于磁盘空间需求。

实际序列数受每 Pod 容器数、网络接口数、块设备数、日志目标/规则数、应用指标属性以及 Pod 重建历史影响。尤其日志默认目标通常按容器生成，不能按 Pod 直接估算。现有代码没有证明 400 Pod 环境需要的准确 IOPS、带宽、TSDB 空间或 Agent CPU/内存。部署联调应记录 `prometheus_tsdb_head_series`、TSDB 磁盘日增长、每节点 Agent 资源和日志增长；用 3～7 天代表性数据外推 30 天保留并为 WAL、Head 与 compaction 留冗余。达到预先设定的存储预警线或序列持续增长时应扩容。

| 组件 | 规划/模板基准 | 解释 |
| --- | --- | --- |
| Monitor Server / Prometheus | 按 400 Pod 的实测负载申请资源 | TSDB 默认保留 30 天，需覆盖 WAL/Head/compaction 峰值 |
| MySQL | 配置、快照、告警和 outbox | 不保存指标或原始日志；容量按记录增长和保留策略评估 |
| Agent，每节点 | request 100m/64Mi，limit 256m/256Mi | 当前 DaemonSet 模板值，不是性能承诺 |
| Agent state | **整套 DaemonSet 共用 1Gi RWX PVC** | 每节点独立子目录；不是每节点各申请 1Gi，游标回收避免按轮转无限累积 |

块存储候选 `topolvm` 仍需确认 device class、可用容量、Pod 调度约束、扩容及节点故障恢复策略。PV/StorageClass 的回收策略也需确认，不能把删除 claim 当作无损操作。模板资源值不能替代当前环境容量验证。

## 5. 验收与边界

开发验收覆盖 Go 单元测试、Web 测试与部署模板静态检查。环境联调应逐项验证：专用身份的真实 RBAC、TLS/Token 轮换、CMDB guid 查询、所有目标节点的日志挂载和 UID/GID 读取、PVC 锁/同步/重命名、Agent → Monitor 与 Prometheus → Agent 网络、TSDB 30 天增长、告警恢复和 SMTP 投递。

SMTP 是外部系统：outbox 和数据库状态可幂等，邮件与数据库不共享事务；发送超时/响应丢失可能导致重复邮件。不要将“邮件恰好一次”作为验收承诺。

静态代码检查或单元测试通过只能证明对应本地逻辑，不证明上述 Kubernetes/CMDB/SMTP 部署条件已验证。系统设计列出当前 API、表结构、恢复语义和未实现范围。
