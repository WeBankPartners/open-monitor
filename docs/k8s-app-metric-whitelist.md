# 应用 metric.log 指标白名单

Agent 对 Monitor 自动生成的 `*_metric.log` 目标使用 `format: otlp_metric`，读取最后一条完整的 OTLP JSON 记录，并只转换下表所列的指标。Monitor 会为每个已发现的普通日志目标生成一个同 Pod、同容器的 metric 目标，因此不需要新增 Prometheus job。输出带有 `cluster_id`、`namespace`、`workload_type`、`workload_name`、`pod`、`pod_uid`、`container`、`instance_id` 和内部 `target` 标签；应用自身允许的维度经过白名单处理，原始主机、账号、路径、URI 等资源属性不会进入标签。

## 当前支持

白名单共 54 个名称。样本 A 命中 40 个、排除 67 个；样本 B 命中 16 个、排除 20 个（按唯一名称统计）。

| 来源名称 | 输出名称（统一加 `k8s_app_` 前缀） | 类型 |
| --- | --- | --- |
| `db.client.connections.idle.max` | `db_client_connections_idle_max` | gauge |
| `db.client.connections.idle.min` | `db_client_connections_idle_min` | gauge |
| `db.client.connections.max` | `db_client_connections_max` | gauge |
| `db.client.connections.pending_requests` | `db_client_connections_pending_requests` | gauge |
| `db.client.connections.usage` | `db_client_connections_usage` | gauge |
| `executor.active` | `executor_active` | gauge |
| `executor.completed` | `executor_completed_total` | counter |
| `executor.pool.core` | `executor_pool_core` | gauge |
| `executor.pool.max` | `executor_pool_max` | gauge |
| `executor.pool.size` | `executor_pool_size` | gauge |
| `executor.queue.remaining` | `executor_queue_remaining` | gauge |
| `executor.queued` | `executor_queued` | gauge |
| `hikaricp.connections.acquire` | `hikaricp_connections_acquire_seconds` | histogram |
| `hikaricp.connections.active` | `hikaricp_connections_active` | gauge |
| `hikaricp.connections.creation` | `hikaricp_connections_creation_seconds` | histogram |
| `hikaricp.connections.idle` | `hikaricp_connections_idle` | gauge |
| `hikaricp.connections.max` | `hikaricp_connections_max` | gauge |
| `hikaricp.connections.min` | `hikaricp_connections_min` | gauge |
| `hikaricp.connections.pending` | `hikaricp_connections_pending` | gauge |
| `hikaricp.connections.timeout` | `hikaricp_connections_timeout_total` | counter |
| `hikaricp.connections.usage` | `hikaricp_connections_usage_seconds` | histogram |
| `http.server.requests` | `http_server_requests_seconds` | histogram |
| `jdbc.connections.active` | `jdbc_connections_active` | gauge |
| `jdbc.connections.idle` | `jdbc_connections_idle` | gauge |
| `jdbc.connections.max` | `jdbc_connections_max` | gauge |
| `jdbc.connections.min` | `jdbc_connections_min` | gauge |
| `jdk.thread.pool.threads` | `jdk_thread_pool_threads` | gauge |
| `jdk.thread.pool.work.queue` | `jdk_thread_pool_work_queue` | gauge |
| `jvm.cpu.recent_utilization` | `jvm_cpu_recent_utilization` | gauge |
| `jvm.cpu.time` | `jvm_cpu_seconds_total` | counter |
| `jvm.gc.duration` | `jvm_gc_duration_seconds` | histogram |
| `jvm.gc.memory.allocated` | `jvm_gc_memory_allocated_bytes_total` | counter |
| `jvm.gc.memory.promoted` | `jvm_gc_memory_promoted_bytes_total` | counter |
| `jvm.gc.pause` | `jvm_gc_pause_seconds` | histogram |
| `jvm.memory.committed` | `jvm_memory_committed_bytes` | gauge |
| `jvm.memory.limit` | `jvm_memory_limit_bytes` | gauge |
| `jvm.memory.max` | `jvm_memory_max_bytes` | gauge |
| `jvm.memory.used` | `jvm_memory_used_bytes` | gauge |
| `jvm.running.time` | `jvm_running_time_seconds` | gauge |
| `jvm.thread.count` | `jvm_thread_count` | gauge |
| `jvm.threads.daemon` | `jvm_threads_daemon` | gauge |
| `jvm.threads.live` | `jvm_threads_live` | gauge |
| `jvm.threads.peak` | `jvm_threads_peak` | gauge |
| `jvm.threads.states` | `jvm_threads_states` | gauge |
| `jvm.trace.exception.events` | `jvm_trace_exception_events_total` | counter |
| `logback.events` | `logback_events_total` | counter |
| `process.cpu.time` | `process_cpu_seconds_total` | counter |
| `process.cpu.usage` | `process_cpu_usage` | gauge |
| `process.files.max` | `process_files_max` | gauge |
| `process.files.open` | `process_files_open` | gauge |
| `process.memory.rss` | `process_memory_rss_bytes` | gauge |
| `process.memory.vss` | `process_memory_vss_bytes` | gauge |
| `process.start.time` | `process_start_time_seconds` | gauge |
| `process.uptime` | `process_uptime_seconds` | gauge |

数值单位会统一为 Prometheus 习惯：字节使用 bytes，时长使用 seconds；累计单调 Sum 输出为 counter，非单调 Sum 按 gauge 处理；Histogram 输出 `_bucket`、`_sum`、`_count`，不会凭空生成分位数。HTTP 原始 URI 被丢弃，但其完整属性集合会形成不透明 `series_id`，以避免不同 URI 的累计值被错误合并；这仍然可能产生较高基数，应由应用端控制 URI 维度数量。

## 当前样本中明确排除

以下名称来自用户提供的两类样本，当前不在白名单中：

```text
application.ready.time
application.started.time
db.druid.commit.count
db.druid.error.count
db.druid.execute.batch.count
db.druid.execute.count
db.druid.execute.query.count
db.druid.execute.totals
db.druid.execute.update.count
db.druid.logic.connect.error.count
db.druid.physical.connect.error.count
db.druid.rollback.count
disk.free
disk.total
hikaricp.connections
http.server.requests.active
jvm.buffer.count
jvm.buffer.memory.used
jvm.buffer.total.capacity
jvm.class.count
jvm.class.loaded
jvm.class.unloaded
jvm.classes.loaded
jvm.classes.unloaded
jvm.compilation.time
jvm.cpu.count
jvm.gc.concurrent.phase.time
jvm.gc.live.data.size
jvm.gc.max.data.size
jvm.gc.overhead
jvm.info
jvm.memory.usage.after.gc
jvm.memory.used_after_last_gc
jvm.threads.started
log.exporter.datapoints
log.exporter.exported
log.exporter.seen
process.memory.swap
processedSpans
queueSize
spring.security.authentications
spring.security.authentications.active
spring.security.authorizations
spring.security.authorizations.active
spring.security.filterchains
spring.security.filterchains.AuthenticationFilter.after
spring.security.filterchains.AuthenticationFilter.before
spring.security.filterchains.access.exceptions.after
spring.security.filterchains.access.exceptions.before
spring.security.filterchains.active
spring.security.filterchains.authentication.anonymous.after
spring.security.filterchains.authentication.anonymous.before
spring.security.filterchains.authentication.basic.after
spring.security.filterchains.authentication.basic.before
spring.security.filterchains.authentication.bearer.after
spring.security.filterchains.authentication.bearer.before
spring.security.filterchains.authorization.after
spring.security.filterchains.authorization.before
spring.security.filterchains.context.async.after
spring.security.filterchains.context.async.before
spring.security.filterchains.context.holder.after
spring.security.filterchains.context.holder.before
spring.security.filterchains.context.servlet.after
spring.security.filterchains.context.servlet.before
spring.security.filterchains.cors.after
spring.security.filterchains.cors.before
spring.security.filterchains.header.after
spring.security.filterchains.header.before
spring.security.filterchains.logout.after
spring.security.filterchains.logout.before
spring.security.filterchains.requestcache.after
spring.security.filterchains.requestcache.before
spring.security.filterchains.session.urlencoding.after
spring.security.filterchains.session.urlencoding.before
spring.security.http.secured.requests
spring.security.http.secured.requests.active
system.cpu.count
system.cpu.usage
system.load.average.1m
tasks.scheduled.execution
tasks.scheduled.execution.active
tomcat.sessions.active.current
tomcat.sessions.active.max
tomcat.sessions.alive.max
tomcat.sessions.created
tomcat.sessions.expired
tomcat.sessions.rejected
```

排除不等于指标没有意义，而是当前 MVP 没有为它定义稳定的单位、类型、维度和告警语义。应用原有 `ERROR` 等日志关键字仍按普通日志规则处理；OTLP 指标中的异常计数不会自动变成日志正文告警。

## 后续增加方式

在 `monitor-agent/k8s-log-agent/internal/agent/otlp_metric.go` 的 `buildAppMetricWhitelist` 增加 `metricSpec`，明确输出名、`gauge`/`counter`/`histogram` 类型、单位和允许的属性；随后为单位换算、累计语义、重复点和 Prometheus 输出补充单元测试。若来源是 Delta Sum、Summary、Exponential Histogram，必须先扩展解析器并定义重启/分位数语义，不能只把名称加入白名单。Druid 计数、磁盘空间和会话指标还应先确认采集端的类型与标签，再决定是否作为默认指标。

增加白名单不需要新增 Prometheus scrape job。构建并部署新 Agent 镜像、完成 DaemonSet 滚动更新后，已有 metric 目标即可使用新能力；仅推送镜像不会更新运行中的 Agent。若需在 Monitor 页面查询和绘图，还须同步维护 `monitor-server/services/k8s/metric_catalog.go` 的查询目录并补充测试；若要默认告警，还需增加明确阈值、窗口和恢复条件。

## 采集边界与验证

- 每次扫描只读取文件尾部最多约 2 MiB，选择最后一条完整记录，单记录不超过 1 MiB。读取当前累计快照，不重放或累加历史记录；重启后仍从最新快照恢复。
- 源数据时间优先使用 `timeUnixNano`，没有该字段时使用行首时间并按 Agent 的 `TZ` 解析。超过 5 分钟或领先当前时间超过 5 秒的数据不输出，不能用旧值伪装实时值。采集端应至少每 5 分钟输出一次，生产建议更短。
- 单目标最多 4096 条输出样本，超限使该目标不可用。此限制约束当前快照，不能限制不断变化的 URI 在 TSDB 中产生的历史序列总数；`series_id` 是属性集合的散列，`source_id` 是文件身份路径的散列，均不展示原始值，也不降低基数。
- `k8s_app_metric_available` 表示目标是否存在可用的新鲜样本；`k8s_app_metric_read_error_total` 表示读取/解析失败次数；`k8s_app_metric_sample_timestamp_seconds` 是最近源数据时间。没有匹配的 `*_metric.log` 文件时，应用指标目标静默跳过，不输出这些状态序列；文件名匹配但打开、读取或解析失败时才输出可用性和错误计数。文件缺失不影响普通关键字日志目标的健康状态。
- 相同身份、相同时间、不同数值的重复序列被排除，其他指标继续输出。样本 B 的 `jvm.trace.exception.events` 存在这种冲突，因此该样本不会输出异常计数。若要启用，需要源端消除冲突，或明确两个 instrumentation scope 的业务含义后增加受控 scope 维度。
- Histogram 没有 bucket 时仅输出 sum/count，可以计算均值，不能计算 P95/P99。`jdk.thread.pool.*` 按来源的非单调 Sum 解释为 gauge，不能仅凭 type 标签名称就当作 counter。
- Monitor 的“监控对象”页面和查询目录已支持应用指标，具体入口见 [对象指标查看说明](k8s-object-metrics.md)。采集白名单不等于默认告警清单；当前没有自动创建 JVM/连接池阈值告警。
- 两份实际样本解析验证分别得到 103 和 131 条样本；样本含历史时间，因此此数字只验证解析兼容性，并不表示这些旧数据会被实时输出。原始业务样本不提交到仓库。
