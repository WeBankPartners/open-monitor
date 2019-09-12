## 用PromQL表示的基础指标
### 主机
| 指标名 | 采集的指标 | promQL表达式 |
| ---- | ---- | ---- |
| cpu使用率 | node_cpu_seconds_total | 100-(avg by (instance) (rate(node_cpu_seconds_total{instance="$address",mode="idle"}[20s]))*100) |
| 内存使用率 | node_memory_MemFree_bytes+node_memory_Cached_bytes+node_memory_Buffers_bytes | 100-((node_memory_MemFree_bytes{instance="$address"}+node_memory_Cached_bytes{instance="$address"}+node_memory_Buffers_bytes{instance="$address"})/node_memory_MemTotal_bytes)*100 |
| 系统负载 | node_load1 | node_load1{instance="$address"} |
| 磁盘读速率 | node_disk_read_bytes_total | irate(node_disk_read_bytes_total{instance="$address",device="$device"}[20s]) |
| 磁盘写速率 | node_disk_written_bytes_total | irate(node_disk_written_bytes_total{instance="$address",device="$device"}[20s]) |
| 磁盘iops | node_disk_io_now | node_disk_io_now{instance="$address",device="$device"} |
| 网卡入流量 | node_network_receive_bytes_total | irate(node_network_receive_bytes_total{instance="$address",device="$device"}[20s]) |
| 网卡出流量 | node_network_transmit_bytes_total | irate(node_network_transmit_bytes_total{instance="$address",device="$device"}[20s]) |

### Mysql
| 指标名 | 采集的指标 | promQL表达式 |
| ---- | ---- | ---- |
| 请求数 | mysql_global_status_commands_total | increase(mysql_global_status_commands_total{instance="$address",command=~"select&#124;insert&#124;update&#124;delete"}[20s]) |
| 连接数 | mysql_global_status_threads_connected | mysql_global_status_threads_connected{instance="$address"} |
| 缓存池内存页 | mysql_global_status_buffer_pool_pages | mysql_global_status_buffer_pool_pages{instance="$address",state=~"old&#124;misc&#124;free&#124;data"} |

### Redis
| 指标名 | 采集的指标 | promQL表达式 |
| ---- | ---- | ---- |
| 内存占用 | redis_memory_used_bytes | redis_memory_used_bytes{instance="$address"} |
| key数量 | redis_db_keys | redis_db_keys{instance="$address"} |
| 指令请求数量 | redis_commands_total | increase(redis_commands_total{instance="$address",cmd=~"get&#124;setex&#124;config&#124;auth&#124;keys"}[20s]) |
| 过期key数量 | redis_expired_keys_total | increase(redis_expired_keys_total{instance="$address"}[20s]) |

