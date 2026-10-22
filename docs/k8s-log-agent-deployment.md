# 日志 Agent DaemonSet 准备与部署步骤

本文适用于 Monitor 以 StatefulSet、日志 Agent 以 DaemonSet 部署在同一 Kubernetes 集群的场景。每个环境部署一套 Agent，覆盖运行该环境业务 Pod 的节点；不需要为每个业务 Namespace 重复部署。

默认先不应用 NetworkPolicy，验证集群内部通信后再按需调整。本文是部署操作指引，不代表已经通过真实环境联调。

## 1. 准备信息

| 项目 | 要求 |
| --- | --- |
| Agent 部署 Namespace | 已存在，通常与对应环境的 Monitor 相同 |
| Monitor Service | 确认集群内 Service 将端口 `19191` 转发到 Monitor 容器端口 `8080`，并能访问 `/monitor` 路径 |
| Cluster ID | 在 Monitor 集群管理页面手工登记；Agent 配置必须与其完全一致 |
| 业务 Namespace | 在 Monitor 中登记并启用；Namespace Token 留空时继承 Cluster Token |
| Agent Token | 必须与 Monitor 的 `MONITOR_K8S_LOG_AGENT_TOKEN` 完全一致；它不是 Kubernetes API Token。当前默认值是公开的共享凭据，生产环境应按环境替换为独立随机值，建议通过 Secret 注入 |
| Agent 镜像 | 使用当前代码构建、推送到集群可拉取的镜像仓库 |
| 初始化镜像 | 使用环境已有的 `package-init-container:1.0.0` 镜像；不同环境只替换镜像仓库前缀 |
| 日志根目录 | 节点上的绝对目录，只读、同路径挂载到 Agent；包含当前环境各实例的日志目录 |
| Agent 状态 PVC | YAML 自动创建 `monitor-k8s-log-agent-state`，使用 `nfs` StorageClass、`ReadWriteMany`、1Gi；允许 UID/GID `10001:10001` 创建独立子目录 |
| 调度与准入 | 允许 DaemonSet、只读日志 hostPath 和非 root 初始化容器；覆盖所有可能运行业务 Pod 的节点 |

可以通过 Kubernetes 查询，无需登录节点：

```bash
kubectl get svc -n <monitor-namespace>
kubectl get svc <monitor-service> -n <monitor-namespace> -o yaml
kubectl get endpointslice -n <monitor-namespace> \
  -l kubernetes.io/service-name=<monitor-service>
kubectl get networkpolicy -A
```

Monitor 地址填写为：

```text
http://<monitor-service>.<monitor-namespace>.svc.<cluster-dns-domain>:19191/monitor
```

集群 DNS 域通常是 `cluster.local`，以实际配置为准。`register.xml` 声明了固定 `19191:8080` 端口映射；Agent 使用集群内 Monitor Service 的 `19191` 端口，Service 后端转发到容器的 `8080`。由于 `register.xml` 当前声明的是 `<docker>` 资源，不能仅凭此映射断定 Kubernetes Service 已自动生成对应端口；部署后需用 `kubectl get svc -n <monitor-namespace> <monitor-service> -o yaml` 确认 `port: 19191` 和 `targetPort: 8080`。支持集群内 HTTP；不需要 Ingress、NodePort 或外部 LoadBalancer。

## 2. 确认自动关联日志的前提

1. 业务 Pod 带有 `instanceId` 标签，值对应 CMDB `app_instance` 的 guid。
2. Monitor 可以通过现有跨插件查询机制读取该记录的 `deploy_path`。
3. Monitor 自动将日志目标目录设为 `{deploy_path}/logs`；该目录必须是 Agent 可读取的绝对路径，并位于配置的日志根目录内。节点与 Agent 内看到的路径一致。
4. 文件位于 `{deploy_path}/logs` 下，匹配 `*.log`。默认规则匹配不区分大小写的 `ERROR/FATAL/PANIC`，不递归扫描子目录。应用指标目标同样在该目录匹配 `*_metric.log`。
5. 目录允许 UID/GID `10001:10001` 遍历，日志文件允许其读取。

Monitor 自动发现 Pod 所在节点并下发日志目标；不需要逐个登记 Pod、Agent 地址或日志规则。首次接管默认从文件末尾开始，因此验收应观察接管后新增的日志。指标中的 `pod` 是 Pod 名称，`instance_id` 是上述 CMDB guid。

业务容器内存在目录，不代表同一路径一定存在于节点上。应检查业务 Pod 的 `volumes`、`volumeMounts` 和实际日志落盘方式；若日志仅在容器可写层或不同路径的卷内，本模板的同路径 hostPath 方案不能直接读取。

## 3. 准备镜像和认证 Token

已有交付镜像时跳过构建。在仓库根目录执行：

```bash
make -C monitor-agent/k8s-log-agent image IMAGE=<registry>/<agent-image>:<tag>
docker push <registry>/<agent-image>:<tag>
```

Monitor 与 Agent 必须使用完全相同的 Bearer Token。当前 `build/register.xml` 与 DaemonSet 模板使用相同的便捷默认值；Monitor 保持默认参数且 DaemonSet 未修改时，无需额外创建 Secret 即可完成认证。这个默认值是公开、固定的共享凭据，不应被视为秘密或用于不可信网络。生产部署应按环境生成独立随机 Token：在 WeCube 中设置 `MONITOR_K8S_LOG_AGENT_TOKEN`，并将 DaemonSet 的 `LOG_AGENT_AUTH_TOKEN` 改为同一个值。日志规则仍由 Monitor API 动态下发，不放入 ConfigMap。

生产环境建议将 Agent Token 改为 Secret 注入，避免把实际凭据保存在 YAML、Git 或命令历史中；如使用 Secret，需将 DaemonSet 的 `LOG_AGENT_AUTH_TOKEN` 改为 `secretKeyRef`，并在 Agent Namespace 创建对应 Secret。以下命令在 Bash 执行，交互输入 Token，避免将真实值写进命令历史或文档：

```bash
AGENT_NS='<agent-namespace>'
read -r -s -p 'Agent token: ' agent_token
printf '\n'
printf '%s' "$agent_token" | kubectl -n "$AGENT_NS" create secret generic monitor-k8s-log-agent-auth \
  --from-file=token=/dev/stdin --dry-run=client -o yaml | kubectl apply -f -
unset agent_token
```

Secret 必须位于 Agent 的 Namespace。不要将真实 Token、实际环境配置或 Secret 导出文件提交到仓库。Token 更新后需重启 Agent Pod，因为 ENV 不会随 Secret 自动刷新；Monitor 侧也需同步更新并生效。

## 4. 准备同一个 YAML 文件

`deploy/daemonset.yaml` 依次包含 PVC、ConfigMap 和 DaemonSet 三个文档，以 `---` 分隔。一次 apply 会创建 PVC `monitor-k8s-log-agent-state`、ConfigMap `monitor-k8s-log-agent-config` 和 DaemonSet `monitor-k8s-log-agent`。Agent 镜像默认命名为 `monitor-k8s-log-agent`。非敏感启动参数集中在 ConfigMap；Token 由 DaemonSet 环境变量提供（可按上一节改为 Secret）；日志目标和规则继续由 Monitor 动态下发。复制一份环境副本后编辑：

```bash
cp monitor-agent/k8s-log-agent/deploy/daemonset.yaml /tmp/k8s-log-agent-deploy.yaml
```

头部列出了所有替换项、查询方式和以下 UAT 示例。这些都是示例代号，不能直接用于实际部署：

| 替换项 | UAT 示例 | 实际值如何确定 |
| --- | --- | --- |
| `REPLACE_AGENT_NAMESPACE` | `monitor-uat` | Agent 部署 Namespace，必须与 state PVC 和可选认证 Secret 相同；通常与对应环境的 Monitor 相同 |
| `REPLACE_MONITOR_URL` | `http://monitor-server.monitor-uat.svc:19191/monitor` | 将 Monitor Service 的 DNS 名和 Namespace 组成完整 URL；固定使用 Service 端口 `19191`，不要填容器端口 `8080` |
| `REPLACE_CLUSTER_ID` | `uat-cluster-example` | Monitor Web 集群管理页面登记的 Cluster ID；可手工指定，两边必须一致 |
| `REPLACE_NODE_LOG_ROOT` | `/var/data` | 应用方确认节点日志根目录，结合业务 Pod 的 `volumes`、`volumeMounts` 验证 |
| `REPLACE_INIT_IMAGE_PREFIX` | `registry.example.invalid/namespace` | 环境镜像仓库前缀，不带末尾 `/`；Agent 与初始化镜像共用该前缀 |
| `REPLACE_AGENT_IMAGE` | `monitor-k8s-log-agent:3.8.0` | 本次构建并推送的 Agent 镜像路径和标签，不含仓库前缀；最终地址为 `<前缀>/monitor-k8s-log-agent:3.8.0` |

只需在 ConfigMap 中编辑 `LOG_AGENT_MONITOR_URL` 这一行，完整填写协议、Service DNS 名、Namespace 和 `/monitor` 路径，避免分开替换 Service 名、Namespace 和端口。示例 Monitor 地址为 `http://monitor-server.monitor-uat.svc:19191/monitor`。Agent 镜像最终使用 `REPLACE_INIT_IMAGE_PREFIX/REPLACE_AGENT_IMAGE`，而初始化镜像使用同一前缀下的 `package-init-container:1.0.0`。不确定实际值时，可以提供以下命令中与部署相关的结果，以及你决定使用的 Cluster ID、Agent 镜像和日志根目录；无需提供任何 Token，也无需登录节点：

```bash
kubectl get svc -A -o custom-columns='NAMESPACE:.metadata.namespace,NAME:.metadata.name,PORTS:.spec.ports[*].port,TARGETS:.spec.ports[*].targetPort'
kubectl get pvc -n <agent-namespace> \
  -o custom-columns='NAME:.metadata.name,STATUS:.status.phase,ACCESS:.spec.accessModes[*],SIZE:.spec.resources.requests.storage,CLASS:.spec.storageClassName'
kubectl get pod <business-pod> -n <business-namespace> \
  -o jsonpath='{range .spec.containers[*]}{.name}{": "}{.volumeMounts}{"\n"}{end}{range .spec.volumes[*]}{.name}{": "}{.hostPath}{" "}{.persistentVolumeClaim}{"\n"}{end}'
```

当前 CMDB `deploy_path` 的日志统一位于其 `logs/` 子目录；例如 `deploy_path=/var/data/SIT_APP_INSTANCE` 时，Monitor 自动下发的日志目录是 `/var/data/SIT_APP_INSTANCE/logs`。因此 `LOG_AGENT_ROOT`、Agent 的日志 `mountPath` 和 `hostPath.path` 均设置为 `/var/data`，Agent 即可读取该目标而无需再做路径转换。Kubernetes 不会把 ConfigMap ENV 自动展开到 `hostPath.path`，因此这三处仍须一起替换。若 Pod 使用 PVC，仅凭容器内路径不能确认节点 hostPath，需按第 2 节确认实际日志落盘方式。

默认 `TZ=Asia/Shanghai`、`LOG_AGENT_STATE_DIR=/state`、`LOG_AGENT_HTTP_ADDR=:9105`，通常无需修改。改变监听端口时必须同步修改 DaemonSet 的端口、探针及相关网络策略，不能只改 ConfigMap。`LOG_AGENT_POLL_INTERVAL=10s` 控制配置拉取和文件扫描；Prometheus 抓取 `/metrics` 并写入 TSDB 的默认间隔是 `30s`，与 `k8s_monitor` 共用 Monitor 的 `MONITOR_K8S_METRICS_INTERVAL_SECONDS`。修改后者并重启 Monitor 才会更新持久化 scrape 配置。

`POD_NAMESPACE`、`LOG_AGENT_NODE_NAME` 和 `LOG_AGENT_ADVERTISE_ADDRESS` 由 Downward API 自动注入；无需登记节点地址。初始化容器使用环境已有的 `package-init-container:1.0.0` 镜像，模板的显式 `command` 会覆盖它的默认 entrypoint，只执行 state 目录初始化，不执行镜像自带的定时化脚本。补充 `imagePullSecrets`、`nodeSelector`、`tolerations` 时按实际集群配置处理。Agent 必须覆盖所有当前环境业务节点，且这些节点已有日志根目录；`hostPath.type: Directory` 不会创建业务日志目录。

### state 使用自动创建的 NFS PVC

模板中的 PVC 声明会由 `kubectl apply -f` 自动创建，名称为 `monitor-k8s-log-agent-state`，使用 `nfs` StorageClass、`ReadWriteMany` 和 1Gi 请求量。确认环境提供名为 `nfs` 且支持动态 RWX 的 StorageClass；PVC 进入 `Bound` 后 DaemonSet Pod 才能挂载。每个节点使用 `subPathExpr: k8s-log-agent/$(POD_NAMESPACE)/$(LOG_AGENT_NODE_NAME)` 挂载自己的 state 子目录到 `/state`。例如 `k8s-log-agent/monitor-uat/node-example`，节点名和 Namespace 自动注入；同节点 Pod 重建继续使用原游标，不采用每次变化的 Pod 名称或 UID。SIT/UAT 使用不同 Namespace 时，子目录也不同。DaemonSet 不支持 `volumeClaimTemplates`，因此由一个 RWX PVC 提供各节点互相隔离的子目录。

- PVC 在 Agent Namespace 中自动创建；它必须能从所有目标节点挂载。PVC 的 `Delete` 回收策略意味着删除 claim 会删除其后端卷数据，避免把该 PVC 当作临时资源删除。
- `prepare-state` 与 Agent 均以 `10001:10001` 运行，用 `umask 077` 创建自己的目录并检查可写性。不使用 root `chown`、递归 `chmod` 或 Pod `fsGroup` 修改共享卷数据；NFS root-squash 不需要关闭。PVC 根目录或预置的 `k8s-log-agent/<agent-namespace>` 父目录必须允许该 UID/GID 创建子目录；若已有权限满足要求，无需额外处理，否则需存储方提供可写目录/PVC。
- `subPathExpr` 隔离读取位置和计数文件，不能提供独立 PVC 的容量配额、权限和故障隔离。复用已有业务 PVC 时，应避免该前缀与其他用途冲突；所有节点仍共用 NFS 的可用空间和可用性。同一 Namespace 下不要用另一个 Agent 部署复用相同子目录。
- Agent 在 NFS 上以临时文件、文件同步和原子重命名提交 checkpoint。文件锁用于阻止同一节点目录被两个 Agent 同时写入；锁被其他进程持有时，新 Agent 会拒绝启动。NFS 的读写/锁错误不会中断已启动 Agent 的日志扫描：它会保留内存游标和计数、持续提供指标，并在后续扫描重试保存。必须在实际 PVC 上验证读写、同步、重命名及锁语义。
- 成功扫描目标目录后，Agent 会删除不再匹配的旧文件游标，将其匹配计数汇总到每目标/规则的累计记录；因此正常轮转不会按轮转次数持续增加 checkpoint。目录或文件读取失败时暂缓该目标的清理，避免临时故障导致游标丢失。
- checkpoint 损坏或 state 读写失败时，Agent 会以 `k8s_log_agent_state_available=0` 标记降级，并增加 `k8s_log_agent_state_resets_total` 或 `k8s_log_agent_state_persist_errors_total`。新接管的普通日志从当前文件末尾开始；已有内存游标继续推进。NFS 恢复后，Agent 将当前内存 checkpoint 写回，不加载可能过期的旧 checkpoint。若进程在内存状态写回前重启，计数可能重置，期间的日志可能漏计；文件轮转等情况下也可能重复计数。此功能不保存原始日志，也不保证精确一次。
- Kubernetes 无法挂载 PVC、`prepare-state` 无法创建挂载子目录等 Pod 启动前故障仍会阻止 Agent 容器启动；应用进程无法绕过 Kubernetes 的卷挂载和 initContainer 阶段。

无需创建节点上的 state 路径。若初始化失败，用第 5 节的 Pod 事件和 `prepare-state` 日志排查 PVC 挂载、权限问题。

ConfigMap、DaemonSet Token 或 Secret 修改后执行 rollout restart，因为 ENV 不会自动刷新：

```bash
kubectl apply -f /tmp/k8s-log-agent-deploy.yaml
kubectl -n "$AGENT_NS" rollout restart daemonset/monitor-k8s-log-agent
kubectl -n "$AGENT_NS" rollout status daemonset/monitor-k8s-log-agent --timeout=180s
```

后续误报排除条件通过 Web 的 Agent 日志目标配置调整，无需修改 DaemonSet 或重启 Agent。Server 自动管理版本，下一次成功拉取后生效；不会撤销历史告警。完整语义见 [Agent 规则说明](../monitor-agent/k8s-log-agent/README.md#可选排除条件)。

## 5. 部署并检查运行状态

先确认占位符已经替换完，第一条命令忽略头部说明注释，应无输出：

```bash
rg -n '^[[:space:]]*[^#[:space:]].*REPLACE_' /tmp/k8s-log-agent-deploy.yaml
kubectl apply --dry-run=server -f /tmp/k8s-log-agent-deploy.yaml
kubectl apply -f /tmp/k8s-log-agent-deploy.yaml
kubectl -n "$AGENT_NS" rollout status daemonset/monitor-k8s-log-agent --timeout=180s
kubectl -n "$AGENT_NS" get daemonset monitor-k8s-log-agent
kubectl -n "$AGENT_NS" get pods -l app=monitor-k8s-log-agent -o wide
```

服务端 dry-run 不代表 Pod 准入、节点目录或运行权限已经验证通过。出现异常时选取对应 Pod 检查：

```bash
kubectl -n "$AGENT_NS" describe pod <agent-pod>
kubectl -n "$AGENT_NS" logs <agent-pod> -c prepare-state
kubectl -n "$AGENT_NS" logs <agent-pod> -c agent --tail=100
```

DaemonSet 的就绪数量应达到预期节点数量，同时核对业务 Pod 所在节点都有 Agent。`/healthz` 仅表示进程存活；`/readyz` 和 `/metrics` 在配置长期不可用时会返回 503。

成功取得配置并至少扫描一个周期后，可以直接从 Pod 检查 state 挂载和 checkpoint，无需节点权限，也不要输出 checkpoint 正文：

```bash
kubectl -n "$AGENT_NS" exec <agent-pod> -c agent -- \
  sh -c 'id; ls -ld /state; test -f /state/checkpoint.json'
```

这只能确认当前 Pod 的读写结果；实际 NFS 锁和重启续读行为仍需联调验证，不能由 PVC 的 `RWX` 标记单独证明。

## 6. 验证两条集群内通信链路

| 发起方 | 目标 | 验证结果 |
| --- | --- | --- |
| Agent | Monitor Service DNS + 服务端口 | Monitor 页面出现自动注册的 Agent，并能成功取得配置 |
| 实际执行抓取的 Prometheus | Agent Pod IP + `9105` | `/metrics` 返回指标，Prometheus 对应抓取目标正常 |

从实际 Prometheus 容器检查某个 Agent（容器需要有 `wget`）：

```bash
kubectl -n <monitor-namespace> exec <monitor-pod> -c <prometheus-container> -- \
  wget -qO- 'http://<agent-pod-ip>:9105/metrics'
```

若容器没有诊断工具，直接通过 Prometheus Targets 页面检查对应抓取任务；不要为了验证而修改运行镜像。上述 `kubectl exec` 属于操作人员诊断权限，不需要授予 Agent 或 Monitor 的采集 Token。

两条链路正常就不需要额外应用 `networkpolicy.yaml`。跨 Namespace 的某次 Service 访问成功，不能证明不存在网络隔离，也不能证明反向抓取一定成功。

如果不通，依次检查 DNS、Service EndpointSlice、端口/监听、Agent 配置就绪状态及已有网络策略。确认由策略阻断后，再针对源和目标两端的隔离情况放行：

- Agent 到 Monitor 的实际服务路径和端口；Service 转换前后端口的策略匹配应按网络插件行为确认。
- Prometheus 到 Agent 的 TCP `9105`。
- 受出站隔离的 Pod 到集群 DNS 的 UDP/TCP `53`，或平台实际使用的 DNS 路径。

**不要直接套用当前 `networkpolicy.yaml`：**它会对选中的 Agent 启用入站/出站隔离，只配置了出站 TCP `80/443`，没有 DNS 放行，且未适配实际 Monitor 端口；入站仅允许模板中指定的 Namespace。需要按实际拓扑修改后才可应用。NetworkPolicy 的 ingress/egress 是流量方向，与 Kubernetes Ingress 入口资源不同。

## 7. 验收与后续调整

- Monitor 集群、Namespace 登记正确，Pod/容器已自动发现。
- Agent 列表显示在线，自动上报地址与实际 Pod IP 一致，未保留错误的人工地址覆盖。
- 有符合条件的业务 Pod 时，自动日志目标数量大于零；配置中的目录对应 CMDB `{deploy_path}/logs`。
- Namespace 登记自动创建 `default-log-error`：`log_matches > 0`，最近 5 分钟新增匹配数大于零即触发，连续 3 个有效正常评估周期恢复；采集异常进入 unknown，不判为恢复。
- Namespace 行的“监控对象”入口可以查看同一 Pod/容器的资源、日志指标和告警状态。
- Prometheus 抓取正常，接管后新增的匹配日志使 `k8s_log_match_total` 增长，包含正确的 `pod`、`instance_id` 标签。
- `k8s_log_file_missing`、`k8s_log_read_error_total` 未持续报告文件缺失或读取错误。
- 需要调整目录、规则或 Agent 地址时，再通过 Web 管理功能补充配置。SMTP 和通知接收人按需人工设置，不影响采集与告警状态记录。

变更镜像、ENV 或挂载后重新 apply 部署副本；修改 Token 后也需重启 DaemonSet：

```bash
kubectl -n "$AGENT_NS" rollout restart daemonset/monitor-k8s-log-agent
kubectl -n "$AGENT_NS" rollout status daemonset/monitor-k8s-log-agent --timeout=180s
```

保留状态目录以便重启后恢复游标。节点迁移后由新节点 Agent 接管当前日志，不保证恢复旧节点不可访问的日志，也不提供原始日志集中存储。

相关文件：[DaemonSet 模板](../monitor-agent/k8s-log-agent/deploy/daemonset.yaml)、[可选 NetworkPolicy 模板](../monitor-agent/k8s-log-agent/deploy/networkpolicy.yaml)、[Agent README](../monitor-agent/k8s-log-agent/README.md)。

## 应用指标日志

Monitor 自动为日志目标生成同目录的 `*_metric.log` 采集目标。应用输出受支持的 OTLP JSON 后，Agent 使用既有 `/metrics` 和 Prometheus job 暴露 `k8s_app_*`，并携带 Pod 与 CMDB guid 标签。普通关键字扫描排除此类文件，避免将 JSON 内的 ERROR 标签误报为日志错误。缺少该可选文件不影响普通日志告警。详见 [白名单与边界](k8s-app-metric-whitelist.md)。
