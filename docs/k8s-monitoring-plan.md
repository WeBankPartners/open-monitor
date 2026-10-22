# WeCube Monitor Kubernetes 监控方案（评审修订稿）

> **WeCube 交付约束**：复用本项目 `build/register.xml` 的插件接口、系统参数、资源依赖及菜单授权机制。编排接口遵循平台 `requestId/inputs` 与 `resultCode/resultMessage/results.outputs` 契约，逐条保留 `guid/callbackParameter`；Token 使用平台敏感参数加密和现有 cipher 库。数据库通过插件包 `init.sql` 的版本标记段安装升级，不新增独立注册、认证或部署协议。namespace 仍需人工登记，再由 Monitor 自动发现其资源。设计目标不代表当前代码已全部完成。

本文用于评审范围、架构取舍、权限交付和容量预算；模块、接口、恢复机制见 [详细系统设计](k8s-monitoring-mvp-system-design.md)。

## 1. 结论

建议采用“Monitor 负责监控与告警，Kubernetes 提供方负责提供受限数据和采集运行条件”的模式：

- Kubernetes 对象状态：Monitor 使用 Kubernetes API 读取；
- CPU、内存以及网络/块设备 I/O 等资源指标：由 Monitor 的 Kubernetes 指标采集适配器读取并按环境/namespace 过滤，再交给自带 Prometheus 保存和查询，告警由 Monitor 评估器管理；Kubernetes 原生 kubelet/cAdvisor 指标作为主数据源，`metrics.k8s.io` 仅作为基础 CPU/内存的兼容回退；Agent `/metrics` 也由 Monitor 侧 Prometheus 抓取并写入同一 TSDB；
- 文件日志：Monitor 提供标准化日志 Agent 镜像、功能和指标契约；部署平台负责按方案安装和运维非特权 DaemonSet。Agent 使用只读 `hostPath` 读取日志，输出计数和状态指标；不保存原始日志；
- 告警、展示、通知、审计和发布变更关联：由 Monitor 负责；
- 应用发布、升级、回滚和日志集中存储：不属于本方案。

Monitor 不使用业务 Token 访问 kubelet。指标采集使用独立的集群级 cAdvisor 读取身份，并在 Monitor 内按环境登记的 namespace 白名单过滤；对象发现仍使用每个 namespace 独立的最小只读身份。这样权限边界与指标扩展能力分开处理。

## 2. 已有证据与适用边界

### 已确认

- 测试集群版本为 Kubernetes `v1.33.6`。
- `metrics.k8s.io/v1beta1` 已注册，`kubectl top pods --containers` 可以返回当前 Pod/容器 CPU 和内存值。
- 应用主要使用 StatefulSet，Pod 名称采用 `name-0`、`name-N` 形式。
- 应用日志通过预设 `hostPath` 写入；不同 Pod 的日志文件名不统一。
- 当前测试样本使用 `nfs.csi.k8s.io` StorageClass，存在 RWX PVC；该存储仅作为现状验证，不作为 Prometheus TSDB 生产推荐。集群已发现 `topolvm` StorageClass，正式部署优先申请 TopoLVM 块存储 PVC。
- 测试样本 StatefulSet 没有 `volumeClaimTemplates`，不能假设每个 ordinal 自动拥有独立 PVC。
- 通过测试管理员身份验证：`metrics.k8s.io` 可查询 Pod/容器 CPU、内存；API Server 代理的 kubelet `/metrics/cadvisor` 可返回带 namespace、pod、container 标签的容器指标。
- 集群存在 VictoriaMetrics 相关资源线索，但不作为本方案或 MVP 的前置依赖。

上述结果来自管理员身份和抽查样本，不代表专用 Token 权限、全节点覆盖或完整指标族已验收。待提供信息统一列于第 10 节。

## 3. 范围与对象模型

Monitor 监控**当前环境中人工登记并启用的 namespace** 内的 StatefulSet、Pod 和容器，并接入资源指标、网络/块设备 I/O 指标和日志指标。namespace 不自动发现；Service、EndpointSlice、PVC、Events 作为后续扩展，不作为首个 MVP 的必需闭环。

### 安装后的最小人工配置

安装插件后，运维只需完成三项必要配置：登记 Cluster 的 API Server/Token/可选 CA，登记需要监控的 Namespace（Namespace Token 可留空以继承 Cluster Token），以及部署日志 Agent DaemonSet 并填写 Monitor 地址、Cluster ID、Agent Token 和同路径日志挂载。Monitor 默认启用 Kubernetes 监控，Namespace 登记后自动创建基础告警规则、发现 StatefulSet/Pod/容器、生成 CMDB 日志目标并下发给已上报的 Agent。Agent 地址、Pod UID、节点迁移、日志游标和指标抓取目标均由系统自动维护。SMTP、通知接收人和人工日志规则属于后续补充配置，不是首个采集闭环的前置条件。

SIT、UAT 各自部署独立 Monitor、数据库及 TSDB。SIT 人工登记 A/B，UAT 人工登记 C/D；环境来自部署配置，不能由接口请求任意覆盖。白名单为空即不采集业务数据，禁用 namespace 立即拒绝新样本入库。

Monitor 将 Kubernetes 的 StatefulSet/Pod 映射为“业务逻辑监控对象”。对当前约定的 StatefulSet 应用，名称是业务对象身份：同一作用域内 `namespace + workload_type + workload_name + pod_name + container_name` 相同，就视为同一个逻辑监控对象，CPU/内存趋势、资源规则和展示连续。

```text
逻辑监控对象键：
cluster_id + namespace + workload_type + workload_name + pod_name + container_name

运行实例维度（不作为业务对象主键）：
workload_uid + pod_uid + container_id
```

SIT、UAT 分别使用独立的 Monitor 部署、数据库、TSDB、凭据和配置边界，因此部署实例本身不再作为业务对象键字段；`cluster_id` 表示 Kubernetes 集群作用域。`workload_type` 当前为 StatefulSet，保留该字段避免未来接入其他 workload 类型时发生名称冲突。`workload_uid`、`pod_uid` 和 `container_id` 记录当前 Kubernetes 对象和容器运行实例，用于判断替换、重建、容器重启、Counter 重置、日志游标切换和审计，不改变逻辑监控对象的主键。

因此，StatefulSet 或其 Pod 被删除后按相同名称重建时，Monitor 默认继续使用同一个逻辑监控对象，CPU/内存趋势和规则保持连续；底层 Counter 必须按新的运行实例重新计算，日志游标按文件身份和运行实例安全切换，不能把两个实例的原始 Counter 直接相加。若同名资源实际上代表了新的业务对象，必须先停用/删除旧的监控配置或显式增加业务版本/实例标识；仅凭 Kubernetes UID 变化不能推翻当前“名称相同即同一监控对象”的业务约定。

StatefulSet 名称和 Pod ordinal 用于展示与槽位聚合；UID 用于运行实例追踪。Pod 消失不直接判定业务对象故障，要结合 StatefulSet 期望副本、新 Pod 状态和数据质量判断。

发布平台不由 Monitor 操作。若需要关联变更，发布平台可通过幂等事件接口通知 Monitor；该接口不属于首个 MVP 前置依赖。

### 3.1 采集启动与对象生命周期

采集有两级触发：

1. **人工登记 namespace**：通过 Web/WeCube 或对应注册接口写入 namespace、环境、集群和凭据引用。只有验证通过且启用的 namespace 才进入采集范围；不需要逐个调用接口添加 Pod 或 StatefulSet。
2. **namespace 内自动发现**：Monitor 对该 namespace 执行 StatefulSet/Pod `list`，随后建立 watch，并按周期全量校准。发现到的 StatefulSet、Pod 和普通业务容器自动加入对象索引、指标过滤范围和日志目标；删除或禁用 namespace 会撤销这些目标。Service、EndpointSlice 等未纳入 MVP。

日志 Agent 不需要在 Monitor 中预填每个地址：Agent 启动/轮询时自动注册节点、可达地址、端口和 Pod UID，Monitor 动态刷新抓取目标；Web/运维接口设置的手工地址具有覆盖优先级，避免自动注册覆盖人工修正。

Kubernetes 运行实例生命周期按 UID 判断，Monitor 业务对象按名称判断：

| 事件 | UID 行为 | Monitor 处理 |
|---|---|---|
| Pod/StatefulSet 原对象更新，`resourceVersion` 变化 | UID 不变 | 更新对象规格、状态、节点、容器和规则关联，保留同一对象的告警与历史 |
| `kubectl replace` 等操作实际删除后重新创建 | 新 UID | 逻辑监控对象名称不变则继续关联；更新运行实例维度、重置 Counter 计算和切换日志游标 |
| StatefulSet 滚动更新 | StatefulSet UID 通常不变；Pod 可能逐个变为新 UID | 保留 StatefulSet 级关联和逻辑对象告警，原始 Counter 与日志游标按运行实例分别结束并接管新实例 |
| Pod 被驱逐、删除后由控制器补建 | 新 Pod UID，通常可能调度到新节点 | 同名 Pod 继续关联逻辑监控对象；新运行实例重新计算 Counter，日志目标切换到新节点，不复用旧游标偏移 |

同名对象在本方案中默认代表同一逻辑监控对象；UID 变化只表示底层运行实例变化。`resourceVersion` 只用于 watch 一致性，不能作为长期对象身份。旧对象的原始样本按运行实例 UID 保留，逻辑对象查询包含重建前后的实例历史，曲线按 Pod UID 分开绘制；当前查询默认只展示仍在登记范围内的活动对象。

Pod 的 `spec.nodeName` 通常在调度后不原地修改；如果平台通过删除/重建实现主机切换，通常表现为旧 Pod 结束、新 Pod UID 出现。Monitor 也兼容观察到 nodeName 变化的运行时事件，并将其视为目标位置变化。Monitor 收到 Pod/容器 watch 事件后重新计算节点抓取集合和日志 Agent 配置：旧节点目标撤销，新节点目标加入；新节点上的 Agent 必须已由部署平台运行并可读对应 hostPath。Agent 不跨节点读取旧节点文件，也不保证节点下线期间的日志零丢失。

## 4. 推荐架构

```text
Kubernetes namespace
  ├─ Kubernetes API：StatefulSet / Pod / container
  ├─ Monitor 日志 Agent DaemonSet：只读 hostPath → 日志计数和状态指标
  └─ kubelet/cAdvisor：CPU / memory / I/O / filesystem 等容器指标

Monitor
  ├─ 对象发现与生命周期同步
  ├─ Kubernetes 指标适配
  ├─ Monitor Prometheus / TSDB
  ├─ 日志指标规则
  ├─ 告警、通知、历史和审计
  ├─ 集群 / namespace / StatefulSet / Pod 展示
  └─ 可选发布变更关联接口
```

对象发现仅对人工登记、验证过且启用的 namespace（DEGRADED 继续重试） 使用 `list/watch`，首个 MVP 只读取 StatefulSet、Pod 和容器所需字段，并通过周期性全量同步校准 watch 断连、过期 `resourceVersion`、Pod 重建和扩缩容。不同环境的 namespace 必须由人分别登记，不能从集群 Namespace 列表或 Pod 自动创建。

## 5. 各类监控设计

### 5.1 StatefulSet、Pod 和容器

监控：

- 期望、副本、当前、Ready、更新中和可用副本；
- ordinal 缺失、Pending、ContainerCreating、CrashLoopBackOff、OOMKilled；
- Ready 状态异常；Liveness 具体失败原因需 Events 等后续数据源，不能仅凭重启推断；
- 重启次数持续增长；
- StatefulSet 更新停滞；
- Pod、容器和工作负载的 CPU、内存、网络和块设备 I/O；网络按 Pod/interface 选择唯一来源，避免重复计数。

StatefulSet 归属以 `ownerReferences` 和 selector 校验。告警唯一键使用逻辑监控对象名称组合和容器名称；工作负载 UID、Pod UID 与 `container_id` 仅作为运行实例字段，避免把同名重建错误拆成新的业务对象，同时保留实例切换审计。

### 5.2 文件日志

日志链路为：

```text
只读 hostPath → Monitor 提供的日志 Agent DaemonSet → 聚合指标 → Monitor 告警
```

日志 Agent 不保存原始日志，只保存游标、文件身份和聚合状态。至少支持：固定目录、glob/正则文件匹配、StatefulSet/Pod label/container 继承和覆盖、关键字/正则、编码和轮转。窗口、阈值和优先级由 Monitor 规则管理，Agent 只输出计数和状态。

Monitor 交付的 Agent 至少应具备：

- 按配置读取一个或多个白名单目录和文件匹配规则；
- 处理新文件、rename、copytruncate、轮转、Agent 重启和节点切换；
- 按规则输出计数、最近匹配时间、文件缺失、读取错误和 Agent 存活指标；
- 使用固定低基数标签（cluster、namespace、workload、pod UID、container、rule），禁止文件名和日志正文进入标签；
- 支持配置校验、版本号、原子加载和失败回滚；
- 提供 `/metrics` 和健康检查接口，并限制自身 CPU、内存、文件句柄和读取速率；
- 记录英文错误和堆栈，不输出日志正文和凭据；
- 以非特权、只读 hostPath 运行，并能在权限不足时明确报告而不是静默漏采。

Agent 不负责应用发布、原始日志存储、全文检索、告警通知或修改 Kubernetes 对象。

指标命名和类型统一见 [系统设计第 5.5 节](k8s-monitoring-mvp-system-design.md#55-agent-指标和健康)，此处不重复定义。

应用的 `*_metric.log` 使用独立 OTLP 目标自动采集并输出 `k8s_app_*`，继续通过既有 Agent Prometheus job 入库，关联同一 Pod 和 CMDB guid。该文件不参与普通关键字匹配；缺失时不影响普通日志告警。白名单、排除项、基数上限和采集边界见 [应用指标说明](k8s-app-metric-whitelist.md)。应用指标可在统一对象页面选择查看曲线；应用指标默认阈值告警仍属后续扩展。

日志告警包括关键字超阈值、ERROR/FATAL 持续出现、文件缺失、读取失败、长时间无更新、轮转后未恢复和 Agent 不可用。

日志目录示例没有体现 ordinal，且 StatefulSet 反亲和不一定保证不同副本在不同节点。因此必须先确认文件名、`subPath` 和实例隔离方式：能唯一关联 Pod 时产生 Pod/ordinal 级告警；无法唯一关联时只产生 StatefulSet/工作负载级告警。

只读取登记的日志文件白名单，禁止递归扫描目录、跟随目录外符号链接或把动态文件名和日志正文放入指标标签。日志轮转、节点切换或 Agent 重启可能造成漏读，不能承诺原始日志零丢失。

### 5.3 CPU、内存、I/O 和 Kubernetes 原生指标

Monitor 的 Kubernetes 指标采集适配器负责抓取、过滤和标准化指标，自带 Prometheus 负责保存、查询和提供规则评估所需的数据；不要求 Kubernetes 提供方额外部署或开放 Prometheus/VictoriaMetrics 查询服务。

#### 主路径：kubelet cAdvisor

Monitor 使用独立的集群级 cAdvisor 读取身份，通过 API Server 代理抓取每个目标节点：

```text
GET /api/v1/nodes/{node}/proxy/metrics/cadvisor
```

节点目标从已登记 Pod 的 `spec.nodeName` 去重得到。该接口可提供容器 CPU counter、working-set memory、网络、块设备和 I/O 等可扩展指标。实际启用的 I/O metric family、标签和重启语义必须在目标集群验收时核对。cAdvisor 响应本身可能包含多个 namespace，因此 Monitor 必须在采集器中按当前 Monitor 实例的环境、人工登记 namespace、Pod 名称/UID 映射和业务容器白名单过滤；未登记 namespace 的样本不得进入 TSDB、规则或 Web。该过滤是应用层范围控制，不替代 `nodes/proxy` 的 Kubernetes 授权。CPU/内存/I/O 按业务容器过滤；网络按 Pod/interface 选唯一来源，不能因 sandbox 或空 container 标签而误删，也不能重复计入多个容器。块设备 I/O 不能替代 NFS/NAS 实际读写诊断。

cAdvisor 权限边界及授权模板见 [第 6.1 节](#61-monitor-api-身份)。

#### 兼容回退：`metrics.k8s.io`

如果 cAdvisor 权限尚未批准或集群不提供 cAdvisor，Monitor 可使用每个已登记 namespace 的只读 ServiceAccount Token 调用：

```text
GET /apis/metrics.k8s.io/v1beta1/namespaces/{namespace}/pods
GET /apis/metrics.k8s.io/v1beta1/namespaces/{namespace}/pods/{pod}
```

它只提供当前 CPU/内存采样值和 timestamp/window，没有 I/O、throttling 或历史；Monitor 仍将采样写入自己的 TSDB。该回退身份只需 namespace 级 `pods.metrics.k8s.io` `get/list`。

CPU 以核数展示，内存以 working-set bytes 展示；request/limit 从 Pod 规格读取，缺失或为零时不计算百分比。cAdvisor CPU counter 使用 `rate()`，内存和 I/O 字节等按指标类型处理；`metrics.k8s.io` CPU 是源窗口平均值，不能再次使用 `rate()`。原始采样失败进入 `Unknown`，不覆盖最近成功值；连续 3 个采样周期失败后产生数据源告警。

### 5.4 PVC 与 NAS

PVC/NAS 深度监控不属于首个 MVP。后续如纳入，PVC 状态通过 Kubernetes API 监控，NAS 容量、配额、inode、挂载健康、延迟和读写错误由提供方或 NAS 平台提供。

## 6. Kubernetes 提供方交付要求

以下是标准交付要求，不能以现有业务 Token 是否满足作为方案前提。

### 6.1 Monitor API 身份

每个 namespace 提供独立、只读、可轮换的 Monitor 身份：

| API 资源 | 权限 |
|---|---|
| core/v1 `pods` | `get/list/watch` |
| apps/v1 `statefulsets` | `get/list/watch` |
| metrics.k8s.io/v1beta1 `pods`（仅启用回退时） | `get/list` |

主路径需要另行提供集群级 `nodes/proxy get` 身份；节点目标从已登记 Pod 的 `spec.nodeName` 去重得到，因此不必额外授予 `nodes get/list/watch`。`nodes/proxy` 不能按 URL 后缀限制为仅 `/metrics/cadvisor`，不是 namespace 隔离的权限；提供方必须单独审批其暴露范围。`metrics.k8s.io` 回退身份仍按 namespace 提供。两类身份均不需要 Secret 全量读取、`pods/exec`、任何写权限或集群 Namespace 列举。

#### cAdvisor 采集身份

下面是主路径的授权示例。若按已登记 Pod 推导节点，不必授予 `nodes get/list/watch`：

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: monitor-cadvisor-reader
  namespace: <MONITOR_NAMESPACE>
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: monitor-cadvisor-reader
rules:
  - apiGroups: [""]
    resources: ["nodes/proxy"]
    verbs: ["get"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: monitor-cadvisor-reader
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: monitor-cadvisor-reader
subjects:
  - kind: ServiceAccount
    name: monitor-cadvisor-reader
    namespace: <MONITOR_NAMESPACE>
```

虽然这里只声明 GET，`nodes/proxy` 不能按 URL 后缀限制为仅 `/metrics/cadvisor`，也不是 namespace 隔离权限；提供方必须单独审批其暴露范围。可通过 `resourceNames` 限定节点，但不能限定 kubelet 路径。测试管理员能读取端点不代表该 Monitor 身份已经获批。

#### namespace 对象身份

以下模板每个已登记 namespace 应用一次，namespace 需事先存在；其中 `metrics.k8s.io` 规则只有启用回退时才加入；由有 RBAC 管理权限的提供方执行，Monitor 运行身份不需要创建授权的权限。

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: monitor-reader
  namespace: <NAMESPACE>
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: monitor-reader
  namespace: <NAMESPACE>
rules:
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["apps"]
    resources: ["statefulsets"]
    verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: monitor-reader
  namespace: <NAMESPACE>
subjects:
  - kind: ServiceAccount
    name: monitor-reader
    namespace: <NAMESPACE>
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: monitor-reader
```

启用 `metrics.k8s.io` 回退时，再向该 Role 增加：

```yaml
  - apiGroups: ["metrics.k8s.io"]
    resources: ["pods"]
    verbs: ["get", "list"]
```

将模板占位符替换后保存为 `monitor-reader-rbac.yaml`：

```bash
kubectl apply -f monitor-reader-rbac.yaml
MONITOR_NS='<NAMESPACE>'
kubectl -n "$MONITOR_NS" create token monitor-reader --duration=24h
```

TokenRequest 实际有效期由 API Server 决定，Token audience 必须适用于 API Server。提供方负责签发和轮换，Monitor 支持安全更新凭据；不要依赖请求一年就一定获得一年有效期。Token 通过凭据管理渠道交付，不写入文档或日志。Monitor 不能用自己的上述只读权限签发新 Token。cAdvisor 主路径身份由提供方在其 SA 所在 namespace 使用同样的 `create token` 命令签发，例如 `kubectl -n "<MONITOR_NAMESPACE>" create token monitor-cadvisor-reader --duration=24h`，这里使用采集 ServiceAccount 所在 namespace，不是业务 namespace。

### 6.2 指标服务

Kubernetes 提供方需要交付 API Server 地址、CA、集群级 cAdvisor Token 和节点代理访问限制；若启用回退，再交付各 namespace 的 `metrics.k8s.io` Token。两类 Token 不混用。外部 Prometheus/VictoriaMetrics 查询地址、认证和租户隔离不是 MVP 前置条件。

### 6.3 日志 Agent 运行条件

Monitor 负责提供日志 Agent 镜像、启动参数、配置格式、指标命名、健康检查、版本兼容、资源建议和验收标准；部署平台负责安装、升级、回滚和运行维护 DaemonSet。默认按**每个集群/环境安全隔离边界部署一套 DaemonSet，每个目标节点运行一个 Agent Pod**，不按 namespace 重复部署。单套 Agent 可通过版本化配置读取多个 namespace 的已授权日志目录；只有在 namespace 之间存在必须隔离的 hostPath、网络、凭据或运维边界时，才拆分为多套 DaemonSet。Monitor 不直接执行 DaemonSet 部署，也不负责应用发布流程。Agent 必须满足：

- 只读挂载租户专属日志根目录或明确白名单路径；
- `privileged=false`、`hostNetwork=false`、`hostPID=false`、`hostIPC=false`、`allowPrivilegeEscalation=false`，并删除不需要的 capabilities；
- 不挂载宿主机根目录、容器运行时 socket 或其他租户目录；
- 提供可读取目录的 UID/GID/ACL，不依赖提权；
- 提供受限且可写的游标目录；
- 配置必要的调度、准入和 NetworkPolicy 条件，但不要求 hostNetwork；
- Agent 可获取所在节点名，并在重启后续读。

部署平台需要按 Monitor 发布的版本和配置契约部署，不应自行修改指标名称、标签、文件读取语义或安全约束。Monitor 发布 Agent 新版本时提供兼容性说明和回滚版本；平台负责灰度、升级窗口和失败回滚。

### 6.4 NAS 与日志契约

提供方或应用管理方必须提供：hostPath 在节点间的路径映射、文件名模式、编码、轮转、多副本实例隔离规则和日志目录只读访问方式。

凭据不得出现在工作负载 annotations/env、示例 YAML、日志或工单中，应通过批准的凭据管理机制注入。

## 7. 告警、展示和外部接口

### 7.1 告警类别与状态

首个 MVP 告警类别：CPU/内存、网络/块设备 I/O、StatefulSet/Pod 基础状态、文件日志，以及 Monitor、Agent、指标查询和通知链路自身健康。应用 `metric.log` 指标已纳入采集和统一对象图表，JVM/连接池/业务阈值告警作为后续扩展；PVC/NAS 和 Kubernetes Events 也作为后续扩展。cAdvisor 的块设备 I/O 仅表示容器/节点可见的块设备累计量，不等同于 NAS/NFS 的后端吞吐、延迟、配额或 inode 使用量。

告警实例唯一键建议为：

```text
cluster_id + namespace + workload_type + workload_name + pod_name + container_name + rule_id + normalized_labels + incident_generation
```

`incident_generation` 在同一对象/规则从 Resolved 后再次满足触发条件时递增，避免历史事件与新事件互相覆盖。数据库另保留对象/规则/标签的当前活动唯一约束，保证同一时刻只有一个活动告警。

分别记录评估状态（`Pending/Firing/Resolved/Unknown`）、处置状态（未确认、已确认）和通知状态（待发送、成功、失败）。确认不代表故障恢复，通知失败不改变告警是否触发。静默、抑制和升级作为后续版本。

### 7.2 展示与通知

提供集群、namespace、StatefulSet、ordinal、Pod、日志命中趋势、未恢复告警和历史告警视图。PVC 和 Events 页面作为后续扩展。MVP 通知使用 SMTP 邮件；其他插件面板、WeCube 流程回调和通用 Webhook 作为后续扩展。通知任务持久化、可重试、可去重并记录失败原因；SMTP 已接受后进程在状态更新前崩溃时可能重复发送，按至少一次投递设计。

### 7.3 发布平台接口

发布平台接口不是首个 MVP 的前置依赖。后续可提供：

1. 变更通知：发布开始、成功、失败、回滚开始、回滚完成，使用幂等 `event_id`，用于变更关联；
2. 监控查询：查询 StatefulSet Ready、未恢复高优先级告警、OOMKilled、资源异常和日志异常。

namespace 和监控对象发现不依赖发布接口；namespace 必须先人工登记。

## 8. 可靠性、安全和容量要求

指标仅保存于 TSDB，不使用 MySQL 归档或历史回查。容量按组件和实际负载评估，30 天 retention 控制旧数据淘汰；容量阈值用于预警扩容；固定容量不意味着无限保存历史。容量公式、部署资源和存储约束统一见本文件第 8.1 节。

- API 同步处理 watch 断连、过期 `resourceVersion`、限流、退避和重列举；Monitor 启动或重启后按“恢复配置 → 全量 list → 原子替换当前索引 → 建立 watch”补齐停机期间的资源变化。
- SIT/UAT 各自部署独立 Monitor 实例；每个环境默认只有一个活动进程，暂不实现同环境多实例锁和租约，部署必须避免重叠实例。数据库唯一键、配置版本和通知任务状态防止单实例重复提交，后续多副本部署前再补充租约。日志目标切换时保留游标。
- 配置使用期望版本、实际版本、发布状态、重试和回滚；数据库唯一键为并发兜底，外部通知按至少一次投递设计并携带幂等键。
- 进程异常记录英文堆栈，重启后恢复任务和游标；采集缺失、查询失败和通知积压分别告警。
- 只持久化对象白名单字段，不记录完整 annotations/env，防止资源对象中的凭据进入日志或数据库。
- 生产接入使用 CA 校验；不使用跳过 TLS 校验作为长期配置。
- 按 namespace/Pod 数量、活跃日志文件数、峰值日志写入量、指标基数、规则数和保留周期做容量预算；本次先采用下述 MVP 基准值，联调后按实测校准。

### 8.1 MVP 初始容量结论（评审基准）

以下为规划假设，实际环境总量待清点：**单个 namespace、60 个 Pod、每个 Pod 平均 1 个业务容器（不计 initContainer）、默认 30 秒采集、30 天保留、每 Pod 最多 10 条日志规则**。容量估算以单设备/单接口、无额外 recording rule 为简化假设，包含首批 CPU、内存、网络和块设备 I/O 序列，不把 Kubernetes endpoint 暴露的全部原始指标无条件写入 TSDB。

指标数量按以下口径估算：

| 指标类别 | 单对象序列估算 | 数量 |
|---|---:|---:|
| Pod/容器资源与状态 | 10 条/Pod（CPU、memory、CPU request、memory limit、restart、ready、网络收发、块设备读写） | 60 × 10 = 600 |
| 日志规则 | 5 条/规则（match、file missing、read error、last match timestamp、last read timestamp） | 60 × 10 × 5 = 3,000 |
| StatefulSet/workload、数据质量、Agent 健康及 Monitor 自身 | 约 200～400 条 | 200～400 |
| **预计活跃序列** |  | **约 3,800～4,000 条** |

因此，本次评审将 **4,000 条作为预计活跃序列数，5,000 条作为单个估算单位的预算预警值（不是实测系统极限）**，并按默认 30 秒采集和 30 天保留计算：

- 预计样本量：`4,000 × 259,200 = 1,036,800,000` 个样本；
- 预算预警值对应样本量：`5,000 × 259,200 = 1,296,000,000` 个样本。

建议首个 MVP 环境申请以下容量：

| 组件 | 初始申请值 | 说明 |
|---|---:|---|
| Monitor Server | 4 vCPU / 8 GiB RAM | API、对象同步、规则评估、Web、SMTP 任务；单实例内部联调基准 |
| TSDB 数据卷 | **50 GiB 起步** | 优先使用 `topolvm` 块存储 PVC，文件系统按 StorageClass 使用 XFS；按一个估算单位申请，多 namespace 环境按实测序列和增长率扩容。30 天保留；其中约 30 GiB 作为容量预警值（不是提前删除阈值），预留 WAL、compaction、文件系统和故障恢复空间；旧指标到期直接删除 |
| MySQL 数据与索引 | 20 GiB | 只保存配置、对象索引、告警、审计、同步/通知任务，不保存指标样本 |
| 日志 Agent（每个目标节点一份） | 0.25 vCPU / 256 MiB RAM | 集群级 DaemonSet；按该节点承载的全部目标 namespace 日志总吞吐和规则数评估；不保存原始日志副本 |
| Agent 游标空间（每个目标节点一份） | 1 GiB | 由该节点上的单个 Agent 共享；仅保存读取游标和运行状态，需使用独立可写目录 |

50 GiB 是一个估算单位的初始 TSDB 申请值，不是所有 namespace 的固定总容量。实际运行时必须通过 `prometheus_tsdb_head_series` 和 TSDB 日增长校准。若活跃序列超过 5,000 条、Pod 超过 60 个、日志规则超过每 Pod 10 条，或磁盘使用率达到 70%，应重新评估，不继续沿用该值。

每环境容量公式：S = 各 namespace 的原始资源/设备/接口序列之和 + 5 × 日志规则目标数 + 质量/Agent/对象序列；样本数 = S × 259,200。两个同规模 namespace 约 7,600～8,000 序列起步，不能沿用单 namespace 上限。原始 Counter 重启换身份和设备维度会增加基数，必须实测。

50 GiB 仅为初始申请，按压缩样本约 1～2 bytes 的粗估不能覆盖全部 WAL、Head、索引和 compaction；用 3～7 天实测增长外推 30 天并预留空间。30 GiB 仅预警；若设置 `retention.size` 可能提前删除不足 30 天的数据，因此本设计不使用它承诺保留 30 天。旧块删除有调度延迟。

Prometheus 主 TSDB 必须申请独立本地盘或受支持的块存储 PVC，不使用当前 NFS/NAS PVC 作为正式生产存储。当前集群的 `topolvm` StorageClass 是优先候选：其 provisioner 为 `topolvm.cybozu.com`，文件系统为 XFS，`WaitForFirstConsumer` 可按 Pod 调度位置绑定本地设备，并支持扩容。正式申请前仍须确认目标节点有可用 device class、50 GiB 容量、Monitor Pod 的调度约束和节点故障后的恢复策略；`Delete` 回收策略也必须由平台确认是否符合生产数据保留要求。

这里的 NFS 结论不是“NAS 吞吐一定不够”，而是 Prometheus TSDB/WAL 依赖本地文件系统对追加写、`fsync`、`mmap`、文件重命名、目录操作、锁和崩溃恢复提供稳定且可验证的语义；NAS 的网络断连、服务端故障、缓存/锁语义和恢复时延会使 WAL 回放、head block、compaction 或 checkpoint 出现写入失败、长时间阻塞、数据损坏或恢复不完整。即使 IOPS 和带宽压测通过，也不能证明这些故障语义满足 Prometheus 的支持边界，因此不能把“压测通过”作为允许生产使用 NFS 的充分条件。当前 NAS 只能用于开发、临时联调或经过书面风险豁免的非关键环境；若无法申请块存储，必须在上线评审中明确接受 TSDB 不可用、指标历史丢失或损坏的风险，并不得宣称满足本方案的 30 天可靠保留目标。不同环境不得共享数据目录。Prometheus、MySQL 的计算资源需另行预算，Monitor Server 的 4 CPU/8 GiB 不是整个系统总资源。

## 9. 验收标准

- 人工登记的 StatefulSet、Pod、容器能发现并持续同步；
- Pod 重建、扩缩容、watch 重连不会串用旧实例告警；
- 资源指标按环境和 namespace 隔离，当前值、历史趋势和数据缺失状态语义清楚；
- cAdvisor CPU、内存、网络和块设备 I/O 的实际 metric family、标签映射、Pod 重建消歧和容器重启处理通过验收；并明确块设备 I/O 不替代 NAS/NFS 后端指标；
- 日志规则能处理 glob、轮转、Agent 重启和多副本归属限制，不保存原始日志；
- 告警去重、确认、恢复、通知重试和失败记录可验证；
- 同环境单活动实例并发提交不会重复创建规则、同一事件或通知任务；故障恢复后再次触发会产生新事件，SMTP 仍可能重复投递；
- API、指标查询、日志 Agent、告警和通知自身异常均可观测。

## 10. 待提供信息

无需继续使用 admin 重复证明端点存在；专用 Token、Agent 和存储条件仍必须在联调时验证。Monitor 需要交付日志 Agent 的镜像和部署/配置规范，部署平台负责实际安装和运维。只需由提供方确认：

1. Monitor API 专用身份、权限、轮换和 CA；
2. cAdvisor 的业务容器 CPU、内存、网络和块设备 I/O 指标覆盖、集群级 `nodes/proxy` 身份及批准范围；如需回退，再确认 `metrics.k8s.io` 可用性和采样 timestamp/window；
3. 日志文件名、编码、轮转和多副本实例隔离规则；
4. Agent 配置发布方式、运行 UID/GID/ACL、游标目录，以及 Agent `/metrics` 的稳定抓取入口；
5. 日志 Agent 的镜像交付与版本来源、部署责任、hostPath 白名单、只读访问、指标格式和非特权安全约束。

提交 YAML、命令输出或截图前，必须删除 Token、JWT、Secret、用户名、密码、内部地址和包仓库凭据。

## 默认日志告警与统一对象入口

Namespace 登记在同一数据库事务中创建 `default-log-error`（`log_matches > 0`），重复登记不覆盖人工修改。Agent 的 ERROR/FATAL/PANIC 是匹配计数规则；Monitor 使用 `increase(k8s_log_match_total[window])` 的作用域过滤查询评估窗口内新增数，其中 `window` 默认为 300 秒，并随抓取周期取 `max(300s, 4×抓取周期)`，不以累计总数直接告警。默认连续 3 个有效正常评估周期恢复；首次尚无足够样本、抓取失败、目标读取异常或数据过期时进入 unknown，不误判恢复。SMTP 与接收策略仍需人工配置。

日志与资源统一使用 `cluster_id / namespace / workload_type / workload_name / pod / pod_uid / container` 标签。日志匹配指标额外保留 `instance_id`（CMDB guid）；UID 隔离运行实例，名称组合维持业务对象身份。Agent 新增 `k8s_log_target_healthy`，当前扫描存在日志文件且所有匹配文件读取成功为 1，否则为 0。

Web 在管理员配置中提供“监控对象”页面，Namespace 行也可直接跳转。支持 cluster/namespace、工作负载、Pod 名称和 instanceId 筛选，统一查看资源、日志、应用指标曲线及当前关联告警状态；支持预设和最长 30 天的自定义时间范围。查询在 Prometheus 端按对象过滤，保留不同 Pod UID 的历史序列，缺失样本留空；自动刷新跟随配置的采集间隔，默认 30 秒。详细操作与指标语义见 [对象指标查看说明](k8s-object-metrics.md)。初始日志读取位置是文件末尾，首次 Prometheus 样本之前的匹配不保证由增量查询捕获；不承诺逐条日志的精确一次告警。
