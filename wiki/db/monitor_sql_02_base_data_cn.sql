USE `monitor`;

insert  into `button`(`id`,`group_id`,`name`,`b_type`,`b_text`,`refresh_panels`,`refresh_charts`,`option_group`,`refresh_button`) values (1,1,'time','select','时间段',0,1,1,0),(2,1,'start_end','datetime','时间区间',0,1,0,0);

insert  into `chart`(`id`,`group_id`,`endpoint`,`metric`,`col`,`url`,`unit`,`title`,`grid_type`,`series_name`,`rate`,`agg_type`,`legend`) values (1,1,'','cpu.used.percent',6,'/dashboard/chart','%','cpu使用率','line','metric',0,'p95','$metric'),(2,2,'','disk.read.bytes',6,'/dashboard/chart','B/s','磁盘读取速度','line','metric',0,'avg','$metric'),(3,2,'','disk.write.bytes',6,'/dashboard/chart','B/s','硬盘写入速度','line','metric',0,'avg','$metric'),(4,1,'','mem.used.percent',6,'/dashboard/chart','%','内存使用率','line','metric',0,'avg','$metric'),(5,1,'','load.1min',6,'/dashboard/chart','','系统负载','line','metric',0,'avg','$metric'),(6,3,'','net.if.in.bytes',6,'/dashboard/chart','B/s','网卡入口流量','line','metric',0,'avg','$metric'),(7,3,'','net.if.out.bytes',6,'/dashboard/chart','B/s','网卡出口流量','line','metric',0,'avg','$metric'),(8,2,'','disk.iops',6,'/dashboard/chart','','IOPS','line','metric',0,'avg','$metric'),(9,4,'','mysql.questions',6,'/dashboard/chart','','请求数','line','metric',0,'avg','$command'),(10,4,'','mysql.threads.connected',6,'/dashboard/chart','','当前连接数','line','metric',0,'avg','$metric'),(11,4,'','mysql.buffer.status',6,'/dashboard/chart','','缓冲池内存页','line','metric',0,'avg','$state'),(12,5,'','redis.mem.used',6,'/dashboard/chart','B','已用内存','line','metric',0,'avg','$metric'),(13,5,'','redis.db.keys',6,'/dashboard/chart','','key数量','line','metric',0,'avg','$db'),(14,5,'','redis.cmd.num',6,'/dashboard/chart','','指令计数','line','metric',0,'avg','$cmd'),(15,5,'','redis.expire.key',6,'/dashboard/chart','','过期key','line','metric',0,'avg','$metric'),(16,6,'','tomcat.connection',6,'/dashboard/chart','','连接数','line','metric',0,'avg','$port'),(17,6,'','tomcat.request',6,'/dashboard/chart','','请求数','line','metric',0,'avg','$port'),(18,6,'','jvm.memory.used',6,'/dashboard/chart','B','jvm内存池使用量','line','metric',0,'avg','$pool'),(19,6,'','jvm.gc.time',6,'/dashboard/chart','s','GC时间','line','metric',0,'avg','$gc');

insert  into `dashboard`(`id`,`dashboard_type`,`search_enable`,`search_id`,`button_enable`,`button_group`,`message_enable`,`message_group`,`message_url`,`panels_enable`,`panels_type`,`panels_group`,`panels_param`) values (1,'host',1,1,1,1,0,0,'',1,'tabs',1,'endpoint={endpoint}'),(2,'mysql',1,1,1,1,0,0,'',1,'tabs',2,'endpoint={endpoint}'),(3,'redis',1,1,1,1,0,0,'',1,'tabs',3,'endpoint={endpoint}'),(4,'tomcat',1,1,1,1,0,0,'',1,'tabs',4,'endpoint={endpoint}');

insert  into `option`(`id`,`group_id`,`option_text`,`option_value`) values (1,1,'30min','-1800'),(2,1,'1小时','-3600'),(3,1,'3小时','-10800'),(4,1,'12小时','-43200'),(5,1,'24小时','-86400');

insert  into `panel`(`id`,`group_id`,`title`,`tags_enable`,`tags_url`,`tags_key`,`chart_group`) values (1,1,'CPU/内存/负载',0,'/dashboard/tags','',1),(2,1,'磁盘',1,'/dashboard/tags','device',2),(3,1,'网络',1,'/dashboard/tags','device',3),(4,2,'mysql基础指标',0,'','',4),(5,3,'redis基础指标',0,'','',5),(6,4,'tomcat基础指标',0,'','',6);

insert  into `prom_metric`(`id`,`metric`,`metric_type`,`prom_ql`,`prom_main`) values (1,'cpu.used.percent','host','100-(avg by (instance) (rate(node_cpu_seconds_total{instance=\"$address\",mode=\"idle\"}[20s]))*100)',''),(2,'mem.used.percent','host','100-((node_memory_MemFree_bytes{instance=\"$address\"}+node_memory_Cached_bytes{instance=\"$address\"}+node_memory_Buffers_bytes{instance=\"$address\"})/node_memory_MemTotal_bytes)*100',''),(3,'load.1min','host','node_load1{instance=\"$address\"}',''),(4,'disk.read.bytes','host','irate(node_disk_read_bytes_total{instance=\"$address\",device=\"$device\"}[20s])','node_disk_read_bytes_total'),(5,'disk.write.bytes','host','irate(node_disk_written_bytes_total{instance=\"$address\",device=\"$device\"}[20s])','node_disk_write_bytes_total'),(6,'disk.iops','host','node_disk_io_now{instance=\"$address\",device=\"$device\"}',''),(7,'net.if.in.bytes','host','irate(node_network_receive_bytes_total{instance=\"$address\",device=\"$device\"}[20s])','node_network_receive_bytes_total'),(8,'net.if.out.bytes','host','irate(node_network_transmit_bytes_total{instance=\"$address\",device=\"$device\"}[20s])','node_network_transmit_bytes_total'),(9,'mysql.questions','mysql','increase(mysql_global_status_commands_total{instance=\"$address\",command=~\"select|insert|update|delete\"}[20s])',''),(10,'mysql.threads.connected','mysql','mysql_global_status_threads_connected{instance=\"$address\"}',''),(11,'mysql.buffer.status','mysql','mysql_global_status_buffer_pool_pages{instance=\"$address\",state=~\"old|misc|free|data\"}',''),(12,'redis.mem.used','redis','redis_memory_used_bytes{instance=\"$address\"}',''),(13,'redis.db.keys','redis','redis_db_keys{instance=\"$address\"}',''),(14,'redis.cmd.num','redis','increase(redis_commands_total{instance=\"$address\",cmd=~\"get|setex|config|auth|keys\"}[20s])',''),(15,'redis.expire.key','redis','increase(redis_expired_keys_total{instance=\"$address\"}[20s])',''),(16,'tomcat.connection','tomcat','tomcat_threadpool_connectioncount{instance="$address"}',''),(17,'tomcat.request','tomcat','increase(tomcat_requestcount_total{instance="$address"}[20s])',''),(18,'jvm.memory.used','tomcat','jvm_memory_pool_bytes_used{instance="$address"}',''),(19,'jvm.gc.time','tomcat','increase(jvm_gc_collection_seconds_count{instance="$address"}[20s])','');

insert  into `search`(`id`,`name`,`search_url`,`search_col`,`refresh_panels`,`refresh_message`) values (1,'endpoint','/dashboard/search?search=','endpoint',1,0);