# 日志 Agent DaemonSet 准备与部署步骤

本文适用于 Monitor 以 StatefulSet、日志 Agent 以 DaemonSet 部署在同一 Kubernetes 集群的场景。每个环境部署一套 Agent，覆盖运行该环境业务 Pod 的节点；不需要为每个业务 Namespace 重复部署。

默认先不应用 NetworkPolicy，验证集群内部通信后再按需调整。本文是部署操作指引，不代表已经通过真实环境联调。

Monitor 部署参数 `MONITOR_K8S_LOCAL_COLLECTION_ENABLED` 默认 `N`，因此默认仅保留发现、日志目标/规则下发和 Agent exporter，不由内置 Prometheus 抓取或计算 Kubernetes 告警。需要内置采集时首次部署/升级显式设置 `Y`，并保留总开关开启；后续使用同模式部署。Agent ENV 和镜像不需要为本开关修改。升级前停旧实例、PVC 清理及外部逐 Pod 抓取验收见[本地采集与告警开关](k8s-local-collection-switch.md)。

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
4. 普通日志在 `{deploy_path}/logs` 下匹配 `*.log`，显式跳过 `_metric.log`，不递归扫描子目录；默认匹配不区分大小写的 `ERROR/FATAL/PANIC`。同目录另有 `*_metric.log` 的应用指标目标，无此可选文件时静默无样本，不影响普通日志告警。支持格式和指标见[白名单说明](k8s-app-metric-whitelist.md)。
5. 目录允许 UID/GID `10001:10001` 遍历，日志文件允许其读取。

Monitor 自动发现 Pod 所在节点并下发日志目标；不需要逐个登记 Pod、Agent 地址或日志规则。首次接管默认从文件末尾开始，因此验收应观察接管后新增的日志。指标中的 `pod` 是 Pod 名称，`instance_id` 是上述 CMDB guid。

新自动生成的目标 ID 使用 `k8s-log-<hash>`，应用指标日志目标另加 `-metric` 后缀。ID 由集群、Namespace、工作负载、Pod 名称、容器名和实例 ID 生成，不是 CMDB 记录的 guid。升级保留已有目标的 ID（包括旧 `cmdb-` 前缀）及自定义规则；同一逻辑目标在 Pod 重建或迁移节点后继续沿用原 ID，无需迁移数据库或 Agent checkpoint。

业务容器内存在目录，不代表同一路径一定存在于节点上。应检查业务 Pod 的 `volumes`、`volumeMounts` 和实际日志落盘方式；若日志仅在容器可写层或不同路径的卷内，本模板的同路径 hostPath 方案不能直接读取。

## 3. 准备镜像与认证

已有交付镜像时跳过构建。在仓库根目录执行：

```bash
make -C monitor-agent/k8s-log-agent image IMAGE=<registry>/monitor-k8s-log-agent:<tag>
docker push <registry>/monitor-k8s-log-agent:<tag>
```

模板 Token 默认与 `build/register.xml` 的 `MONITOR_K8S_LOG_AGENT_TOKEN` 相同，是公开的便捷默认值。实际配置须两侧一致；环境独立 Token 和可选 Secret 的操作见第 8 节。它只用于 Agent 与 Monitor 通信，不是 Kubernetes API Token。

## 4. 准备同一个 YAML 文件

`deploy/daemonset.yaml` 依次包含 PVC、ConfigMap 和 DaemonSet 三个文档，以 `---` 分隔。一次 apply 会创建 PVC `monitor-k8s-log-agent-state`、ConfigMap `monitor-k8s-log-agent-config` 和 DaemonSet `monitor-k8s-log-agent`。Agent 镜像默认命名为 `monitor-k8s-log-agent`。非敏感启动参数集中在 ConfigMap；Token 由 DaemonSet 环境变量提供（可按第 8 节改为 Secret）；日志目标和规则继续由 Monitor 动态下发。复制一份环境副本后编辑：

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

Monitor URL 整行替换；Agent 和初始化镜像共用仓库前缀。不确定 Service 端口和业务路径时，可查询以下信息，无需登录节点或提供 Token：

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

### NFS state 的部署条件

PVC 随 YAML 自动创建，无需预建节点 state 目录。`nfs` StorageClass 必须可动态提供 RWX；PVC 根目录或 `k8s-log-agent/<agent-namespace>` 父目录应允许 `10001:10001` 创建子目录。初始化容器以该身份创建 `k8s-log-agent/<namespace>/<node>`，不做递归 chown/chmod；同节点 Pod 重建沿用原游标。

各节点共用一个 1Gi claim、按子目录隔离，没有独立容量配额和存储故障隔离。正常轮转会回收游标，不按轮转次数无限增长；仍应按当前文件、规则和目标数观察用量。PV 回收策略由实际 PV/StorageClass 决定；删除 claim 前确认后端数据是否会被回收。

必须在实际 PVC 上验证文件锁、同步和原子重命名。state 操作返回错误时可内存降级重试；NFS 调用阻塞仍可能阻塞 Agent。PVC 无法挂载或初始化目录失败会阻止 Pod 启动。具体游标与丢失窗口见[Agent checkpoint 说明](../monitor-agent/k8s-log-agent/README.md#文件与-checkpoint)。

## 5. 部署并检查运行状态

设置实际 Namespace，检查占位符，再执行部署；占位符检查忽略头部注释，确认无输出后才继续 apply：

```bash
AGENT_NS='<agent-namespace>'
rg -n '^[[:space:]]*[^#[:space:]].*REPLACE_' /tmp/k8s-log-agent-deploy.yaml
kubectl apply --dry-run=server -f /tmp/k8s-log-agent-deploy.yaml
kubectl apply -f /tmp/k8s-log-agent-deploy.yaml
kubectl -n "$AGENT_NS" rollout status daemonset/monitor-k8s-log-agent --timeout=180s
kubectl -n "$AGENT_NS" get daemonset monitor-k8s-log-agent
kubectl -n "$AGENT_NS" get pods -l app=monitor-k8s-log-agent -o wide
```

`AGENT_NS` 必须与 YAML 的 Namespace 一致。服务端 dry-run 不代表目录或运行权限已验证。出现异常时检查：

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

## 联调排障日志

Monitor 应用日志使用以下事件名定位链路：`Kubernetes discovery snapshot ready` 表示 Kubernetes 清单已得到，`Kubernetes log target resolution failed` 表示某个 `instanceId` 的 CMDB 查询或路径解析失败，`Kubernetes log agent config snapshot prepared` 表示节点配置已生成待持久化。Agent 日志中的 `agent registration failed`、`configuration not published yet`、`configuration refresh failed`、`configuration applied` 和 `log scan summary` 分别对应登记、等待配置、拉取失败、应用配置和本地目录读取；重复错误会限频。日志不会输出 Bearer Token、完整 Authorization、CMDB 响应正文或业务日志内容。

可在 Monitor 容器执行：

```sh
grep -nE 'Kubernetes discovery snapshot ready|Kubernetes log target resolution failed|Kubernetes log agent config snapshot prepared|CMDB query' logs/open-monitor.log
grep -nE 'internal/k8s/log-agent/(register|config)' logs/open-monitor-access.log
```

可在 Agent 容器执行：

```sh
grep -nE 'agent registration|configuration (not published|refresh failed|applied)|log scan summary' /path/to/agent.log
```
- Namespace 登记自动创建 `default-log-error`：`log_matches > 0`，最近增量窗口默认 5 分钟（间隔较长时至少覆盖 4 个采样间隔），连续有效正常评估达到恢复条件后恢复；采集异常进入 unknown，不判为恢复。
- Namespace 行的“监控对象”入口可以查看同一 Pod/容器的资源、日志指标和告警状态。
- Prometheus 抓取正常，接管后新增的匹配日志使 `k8s_log_match_total` 增长，包含正确的 `pod`、`instance_id` 标签。
- `k8s_log_file_missing`、`k8s_log_read_error_total` 未持续报告文件缺失或读取错误。
- 新版 Agent 命中新日志且摘要预算可用时，`/metrics` 额外输出 `k8s_log_match_sample_timestamp_seconds`，含文件名和起始字节偏移。旧断点和目标 JSON 无需迁移；外部 Prometheus/Grafana 的正文接入与限额说明见[日志命中摘要](k8s-log-match-samples.md)。
- 需要调整目录、规则或 Agent 地址时，再通过 Web 管理功能补充配置。SMTP 和通知接收人按需人工设置，不影响采集与告警状态记录。

变更部署副本后重新 apply；ConfigMap/Secret 注入的 ENV 不会自动刷新，需执行一次滚动重启。更新 Monitor 端 Token 时也要使对应服务配置生效：

```bash
kubectl apply -f /tmp/k8s-log-agent-deploy.yaml
kubectl -n "$AGENT_NS" rollout restart daemonset/monitor-k8s-log-agent
kubectl -n "$AGENT_NS" rollout status daemonset/monitor-k8s-log-agent --timeout=180s
```

保留状态目录以便重启后恢复游标。节点迁移后由新节点 Agent 接管当前日志，不保证恢复旧节点不可访问的日志，也不提供原始日志集中存储。

相关文件：[DaemonSet 模板](../monitor-agent/k8s-log-agent/deploy/daemonset.yaml)、[可选 NetworkPolicy 模板](../monitor-agent/k8s-log-agent/deploy/networkpolicy.yaml)、[Agent README](../monitor-agent/k8s-log-agent/README.md)。

## 8. 可选：独立 Token 与 Secret

建议为环境设置独立随机 Token；在 Monitor 的 `MONITOR_K8S_LOG_AGENT_TOKEN` 和 Agent 使用相同值。若改用 Secret，须同时把 DaemonSet 的 `LOG_AGENT_AUTH_TOKEN` 从 `value` 改为 `valueFrom.secretKeyRef`，指向 `monitor-k8s-log-agent-auth` 的 `token` 键。Secret 位于 Agent Namespace，不随当前模板自动创建；这是可选变更，默认部署不依赖它。

以下命令在 Bash 中交互输入，不把 Token 写进命令历史：

```bash
read -r -s -p 'Agent token: ' agent_token
printf '\n'
printf '%s' "$agent_token" | kubectl -n "$AGENT_NS" create secret generic monitor-k8s-log-agent-auth \
  --from-file=token=/dev/stdin --dry-run=client -o yaml | kubectl apply -f -
unset agent_token
```

不要将 Token 或环境 Secret 导出文件提交到仓库。ENV 刷新方式见第 7 节；动态日志规则和排除条件在 Web 修改后由 Agent 下轮成功拉取生效，无需重启，具体语义见[规则说明](../monitor-agent/k8s-log-agent/README.md#可选排除条件)。
