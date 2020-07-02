USE `monitor`;

insert  into `button`(`id`,`group_id`,`name`,`b_type`,`b_text`,`refresh_panels`,`refresh_charts`,`option_group`,`refresh_button`) values (1,1,'time','select','Time',0,1,1,0),(2,1,'start_end','datetime','Date',0,1,0,0);

insert  into `chart`(`id`,`group_id`,`endpoint`,`metric`,`col`,`url`,`unit`,`title`,`grid_type`,`series_name`,`rate`,`agg_type`,`legend`) values (1,1,'','cpu.used.percent^cpu.detail.percent',6,'/dashboard/chart','%','cpu.used.percent','line','metric',0,'p95','$custom_metric'),(2,2,'','disk.read.bytes',6,'/dashboard/chart','B/s','disk.read.bytes','line','metric',0,'avg','$metric'),(3,2,'','disk.write.bytes',6,'/dashboard/chart','B/s','disk.write.bytes','line','metric',0,'avg','$metric'),(4,1,'','mem.used.percent',6,'/dashboard/chart','%','mem.used.percent','line','metric',0,'avg','$metric'),(6,3,'','net.if.in.bytes',6,'/dashboard/chart','B/s','net.if.in.bytes','line','metric',0,'avg','$metric'),(7,3,'','net.if.out.bytes',6,'/dashboard/chart','B/s','net.if.out.bytes','line','metric',0,'avg','$metric'),(8,2,'','disk.iops',6,'/dashboard/chart','','IOPS','line','metric',0,'avg','$metric'),(9,4,'','mysql.requests',6,'/dashboard/chart','','mysql.requests','line','metric',0,'avg','$command'),(10,4,'','mysql.threads.connected^mysql.threads.max',6,'/dashboard/chart','','mysql.threads.connected','line','metric',0,'avg','$metric'),(11,4,'','mysql.buffer.status',6,'/dashboard/chart','','mysql.buffer.status','line','metric',0,'avg','$state'),(12,5,'','redis.mem.used',6,'/dashboard/chart','B','redis.mem.used','line','metric',0,'avg','$metric'),(13,5,'','redis.db.keys',6,'/dashboard/chart','','redis.db.keys','line','metric',0,'avg','$db'),(14,5,'','redis.cmd.num',6,'/dashboard/chart','','redis.cmd.num','line','metric',0,'avg','$cmd'),(15,5,'','redis.expire.key',6,'/dashboard/chart','','redis.expire.key','line','metric',0,'avg','$metric'),(16,8,'','jvm.thread.count',6,'/dashboard/chart','','jvm.thread.count','line','metric',0,'avg','$metric'),(18,8,'','jvm.memory.heap.used^jvm.memory.heap.max',6,'/dashboard/chart','B','jvm.memory.heap.used','line','metric',0,'avg','$metric'),(19,8,'','jvm.gc.time',6,'/dashboard/chart','s','jvm.gc.time(2min)','line','metric',0,'avg','$name'),(20,1,'','mem.used^mem.total',6,'/dashboard/chart','B','mem.status','line','metric',0,'avg','$metric'),(21,1,'','load.1min',6,'/dashboard/chart','','load.1min','line','metric',0,'avg','$metric'),(22,7,'','process_cpu_used_percent',6,'/dashboard/chart','%','process.cpu.used','line','metric',0,'avg','$name'),(23,7,'','process_mem_byte',6,'/dashboard/chart','B','process.mem.used','line','metric',0,'avg','$name'),(24,6,'','tomcat.connection',6,'/dashboard/chart','','tomcat.connection','line','metric',0,'avg','$name'),(25,6,'','tomcat.request',6,'/dashboard/chart','','tomcat.request','line','metric',0,'avg','$name'),(26,9,'','app.metric',6,'/dashboard/chart','','${auto}','line','metric',0,'avg','$key'),(27,10,'','telnet_alive',6,'/dashboard/chart','','telnet.alive','line','metric',0,'avg','$metric');

insert  into `dashboard`(`id`,`dashboard_type`,`search_enable`,`search_id`,`button_enable`,`button_group`,`message_enable`,`message_group`,`message_url`,`panels_enable`,`panels_type`,`panels_group`,`panels_param`) values (1,'host',1,1,1,1,0,0,'',1,'tabs',1,'endpoint={endpoint}'),(2,'mysql',1,1,1,1,0,0,'',1,'tabs',2,'endpoint={endpoint}'),(3,'redis',1,1,1,1,0,0,'',1,'tabs',3,'endpoint={endpoint}'),(4,'tomcat',1,1,1,1,0,0,'',1,'tabs',4,'endpoint={endpoint}'),(5,'java',1,1,1,1,0,0,'',1,'tabs',5,'endpoint={endpoint}'),(6,'telnet',1,1,1,1,0,0,'',1,'tabs',6,'endpoint={endpoint}');

insert  into `option`(`id`,`group_id`,`option_text`,`option_value`) values (1,1,'30min','-1800'),(2,1,'1h','-3600'),(3,1,'3h','-10800'),(4,1,'12h','-43200'),(5,1,'24h','-86400');

insert  into `panel`(`id`,`group_id`,`title`,`tags_enable`,`tags_url`,`tags_key`,`chart_group`,`auto_display`) values (1,1,'CPU/Mem/Load',0,'/dashboard/tags','',1,0),(2,1,'Disk',1,'/dashboard/tags','device',2,0),(3,1,'Net',1,'/dashboard/tags','device',3,0),(4,2,'Mysql',0,'','',4,0),(5,3,'Redis',0,'','',5,0),(6,4,'Tomcat',0,'','',6,0),(7,1,'Process',0,'','',7,0),(8,5,'Java',0,'','',8,0),(9,1,'Business',0,'','key',9,1),(10,6,'Alive',0,'','',10,0);

insert  into `prom_metric`(`id`,`metric`,`metric_type`,`prom_ql`,`prom_main`) values (1,'cpu.used.percent','host','100-(avg by (instance) (rate(node_cpu_seconds_total{instance=\"$address\",mode=\"idle\"}[20s]))*100)',''),(2,'mem.used.percent','host','100-((node_memory_MemFree_bytes{instance=\"$address\"}+node_memory_Cached_bytes{instance=\"$address\"}+node_memory_Buffers_bytes{instance=\"$address\"})/node_memory_MemTotal_bytes{instance=\"$address\"})*100',''),(3,'load.1min','host','node_load1{instance=\"$address\"}',''),(4,'disk.read.bytes','host','irate(node_disk_read_bytes_total{instance=\"$address\"}[20s])','node_disk_read_bytes_total'),(5,'disk.write.bytes','host','irate(node_disk_written_bytes_total{instance=\"$address\"}[20s])','node_disk_write_bytes_total'),(6,'disk.iops','host','irate(node_disk_reads_completed_total{instance=\"$address\"}[20s])+irate(node_disk_writes_completed_total{instance=\"$address\"}[20s])',''),(7,'net.if.in.bytes','host','irate(node_network_receive_bytes_total{instance=\"$address\"}[20s])','node_network_receive_bytes_total'),(8,'net.if.out.bytes','host','irate(node_network_transmit_bytes_total{instance=\"$address\"}[20s])','node_network_transmit_bytes_total'),(9,'mysql.requests','mysql','increase(mysql_global_status_commands_total{instance=\"$address\",command=~\"select|insert|update|delete|commit|rollback\"}[20s])',''),(10,'mysql.threads.connected','mysql','mysql_global_status_threads_connected{instance=\"$address\"}',''),(11,'mysql.buffer.status','mysql','mysql_global_status_buffer_pool_pages{instance=\"$address\",state=~\"old|misc|free|data\"}',''),(12,'redis.mem.used','redis','redis_memory_used_bytes{instance=\"$address\"}',''),(13,'redis.db.keys','redis','redis_db_keys{instance=\"$address\"}',''),(14,'redis.cmd.num','redis','increase(redis_commands_total{instance=\"$address\",cmd=~\"get|setex|config|auth|keys\"}[20s])',''),(15,'redis.expire.key','redis','increase(redis_expired_keys_total{instance=\"$address\"}[20s])',''),(16,'tomcat.connection','tomcat','Tomcat_ThreadPool_connectionCount{instance=\"$address\"}',''),(17,'tomcat.request','tomcat','Tomcat_Servlet_requestCount{instance=\"$address\"}',''),(18,'jvm.memory.heap.used','java','java_lang_Memory_HeapMemoryUsage_used{instance=\"$address\"}',''),(19,'jvm.gc.time','java','increase(java_lang_GarbageCollector_CollectionTime{instance=\"$address\"}[2m])',''),(20,'heap.mem.used.percent','tomcat','java_lang_Memory_HeapMemoryUsage_used{instance=\"$address\"}/java_lang_Memory_HeapMemoryUsage_max{instance=\"$address\"} * 100',''),(21,'gc.marksweep.time','tomcat','increase(java_lang_GarbageCollector_CollectionTime{name=~\".*MarkSweep.*\",instance=\"$address\"}[2m])',''),(22,'mysql.threads.max','mysql','mysql_global_variables_max_connections',''),(23,'mysql.connect.used.percent','mysql','mysql_global_status_max_used_connections{instance=\"$address\"}/mysql_global_variables_max_connections{instance=\"$address\"} * 100',''),(24,'mysql.alive','mysql','mysql_up{instance=\"$address\"}',''),(25,'redis.alive','redis','redis_up{instance=\"$address\"}',''),(26,'redis.client.used.percent','redis','redis_connected_clients{instance=\"$address\"}/redis_config_maxclients{instance=\"$address\"} * 100',''),(27,'mem.used','host','node_memory_MemTotal_bytes{instance=\"$address\"}-(node_memory_MemFree_bytes{instance=\"$address\"}+node_memory_Cached_bytes{instance=\"$address\"}+node_memory_Buffers_bytes{instance=\"$address\"})',''),(28,'mem.total','host','node_memory_MemTotal_bytes{instance=\"$address\"}',''),(29,'cpu.detail.percent','host','avg by (mode) (rate(node_cpu_seconds_total{instance=\"$address\",mode=~\"user|system|iowait\"}[20s]))*100',''),(30,'jvm.memory.heap.max','tomcat','java_lang_Memory_HeapMemoryUsage_max{instance=\"$address\"}',''),(31,'app.metric','host','node_business_monitor_value{instance=\"$address\"}',''),(32,'process_alive_count','host','node_process_monitor_count_current{instance=\"$address\"}',''),(33,'process_cpu_used_percent','host','node_process_monitor_cpu{instance=\"$address\"}',''),(34,'process_mem_byte','host','node_process_monitor_mem{instance=\"$address\"}',''),(35,'jvm.thread.count','java','java_lang_Threading_ThreadCount{instance=\"$address\"}',''),(36,'ping_alive','host','ping_alive{guid=\"$guid\"}',''),(37,'telnet_alive','host','telnet_alive{guid=\"$guid\"}',''),(38,'ping_time','host','ping_time{guid=\"$guid\"}','');

insert  into `search`(`id`,`name`,`search_url`,`search_col`,`refresh_panels`,`refresh_message`) values (1,'endpoint','/dashboard/search?search=','endpoint',1,0);

insert  into `grp`(`id`,`name`,`create_at`) VALUES (1,'default_host_group',NOW()),(2,'default_mysql_group',NOW()),(3,'default_redis_group',NOW()),(4,'default_tomcat_group',NOW()),(5,'default_java_group',NOW()),(6,'default_ping_group',NOW()),(7,'default_telnet_group',NOW());

insert  into `tpl`(`id`,`grp_id`,`endpoint_id`,`notify_url`,`create_user`,`update_user`,`action_user`,`action_role`,`create_at`,`update_at`) values (1,1,0,'','','','','','2020-01-16 15:24:03','2020-01-16 15:35:48'),(2,4,0,'','','','','','2020-01-16 17:24:02','2020-01-19 14:45:26'),(3,2,0,'','','','','','2020-01-19 11:22:05','2020-01-19 14:43:09'),(4,3,0,'','','','','','2020-01-19 14:30:33','2020-01-19 14:30:33'),(5,5,0,'','','','','','2020-01-19 14:30:33','2020-01-19 14:30:33'),(6,6,0,'','','','','','2020-01-19 14:30:33','2020-01-19 14:30:33'),(7,7,0,'','','','','','2020-01-19 14:30:33','2020-01-19 14:30:33');

insert  into `strategy`(`id`,`tpl_id`,`metric`,`expr`,`cond`,`last`,`priority`,`content`,`config_type`) values (1,1,'cpu.used.percent','100-(avg by (instance) (rate(node_cpu_seconds_total{instance=\"$address\",mode=\"idle\"}[20s]))*100)','>85','60s','medium','cpu used high',''),(2,1,'mem.used.percent','100-((node_memory_MemFree_bytes{instance=\"$address\"}+node_memory_Cached_bytes{instance=\"$address\"}+node_memory_Buffers_bytes{instance=\"$address\"})/node_memory_MemTotal_bytes)*100','>85','60s','medium','memory used high',''),(3,1,'load.1min','node_load1{instance=\"$address\"}','>3','60s','medium','sys load busy',''),(4,1,'disk.read.bytes','irate(node_disk_read_bytes_total{instance=\"$address\"}[20s])','>10485760','60s','low','disk read busy',''),(5,1,'disk.write.bytes','irate(node_disk_written_bytes_total{instance=\"$address\"}[20s])','>10485760','60s','low','disk write busy',''),(6,1,'net.if.in.bytes','irate(node_network_receive_bytes_total{instance=\"$address\"}[20s])','>52428800','60s','medium','net flow is high',''),(7,1,'net.if.out.bytes','irate(node_network_transmit_bytes_total{instance=\"$address\"}[20s])','>52428800','60s','medium','net flow is high',''),(8,5,'heap.mem.used.percent','java_lang_Memory_HeapMemoryUsage_used{instance=\"$address\"}/java_lang_Memory_HeapMemoryUsage_max{instance=\"$address\"} * 100','>80','60s','medium','heap memory use high',''),(9,5,'gc.marksweep.time','increase(java_lang_GarbageCollector_CollectionTime{name=~\".*MarkSweep.*\",instance=\"$address\"}[2m])','>60','60s','medium','marksweep gc frequent occurrence',''),(10,3,'mysql.connect.used.percent','mysql_global_status_max_used_connections{instance=\"$address\"}/mysql_global_variables_max_connections{instance=\"$address\"} * 100','>80','60s','medium','mysql connection used too many',''),(11,3,'mysql.alive','mysql_up{instance=\"$address\"}','<1','30s','high','mysql instance down',''),(12,4,'redis.alive','redis_up{instance=\"$address\"}','<1','30s','high','redis instance down',''),(13,4,'redis.client.used.percent','redis_connected_clients{instance=\"$address\"}/redis_config_maxclients{instance=\"$address\"} * 100','>50','60s','medium','redis connected too many client ',''),(14,1,'cpu.used.percent','100-(avg by (instance) (rate(node_cpu_seconds_total{instance=\"$address\",mode=\"idle\"}[20s]))*100)','>95','60s','high','cpu high',''),(15,1,'ping_alive','ping_alive{guid=\"$guid\"}','>0','60s','high','ping fail',''),(16,6,'ping_alive','ping_alive{guid=\"$guid\"}','>0','60s','high','ping fail',''),(17,7,'telnet_alive','telnet_alive{guid=\"$guid\"}','>0','60s','high','telnet fail',''),(18,1,'mem.used.percent','100-((node_memory_MemFree_bytes{instance=\"$address\"}+node_memory_Cached_bytes{instance=\"$address\"}+node_memory_Buffers_bytes{instance=\"$address\"})/node_memory_MemTotal_bytes)*100','>50','60s','low','memory used over 50%','');

insert  into `user`(`name`,`passwd`,`display_name`,`email`,`phone`,`ext_contact_one`,`ext_contact_two`,`creator`,`created`) values ('admin','yfBtw71KfVF7jBEyoX+ByyB+5FuEyjFYnd3CHXHR910=','admin','admin@xx.com','12345678910','','',0,NOW());

#@v1.5.0.5-begin@;
insert  into `dashboard`(`dashboard_type`,`search_enable`,`search_id`,`button_enable`,`button_group`,`message_enable`,`message_group`,`message_url`,`panels_enable`,`panels_type`,`panels_group`,`panels_param`) values ('nginx',1,1,1,1,0,0,'',1,'tabs',7,'endpoint={endpoint}');
insert  into `panel`(`group_id`,`title`,`tags_enable`,`tags_url`,`tags_key`,`chart_group`,`auto_display`) values (7,'Nginx',0,'','',11,0);
insert  into `chart`(`group_id`,`endpoint`,`metric`,`col`,`url`,`unit`,`title`,`grid_type`,`series_name`,`rate`,`agg_type`,`legend`) values (11,'','nginx.connect.active',6,'/dashboard/chart','','connect.active','line','metric',0,'avg','$metric'),(11,'','nginx.handle.request',6,'/dashboard/chart','','handle.request','line','metric',0,'avg','$metric');
insert  into `prom_metric`(`metric`,`metric_type`,`prom_ql`,`prom_main`) values ('nginx.connect.active','nginx','nginx_connections_active{instance=\"$address\"}',''),('nginx.handle.request','nginx','increase(nginx_connections_handled{instance=\"$address\"}[20s])','');
insert  into `grp`(`name`,`create_at`) VALUES ('default_nginx_group',NOW());
#@v1.5.0.5-end@;

#@v1.5.0.6-begin@;
insert  into `dashboard`(`dashboard_type`,`search_enable`,`search_id`,`button_enable`,`button_group`,`message_enable`,`message_group`,`message_url`,`panels_enable`,`panels_type`,`panels_group`,`panels_param`) values ('http',1,1,1,1,0,0,'',1,'tabs',8,'endpoint={endpoint}');
insert  into `panel`(`group_id`,`title`,`tags_enable`,`tags_url`,`tags_key`,`chart_group`,`auto_display`) values (8,'http',0,'','',12,0);
insert  into `chart`(`group_id`,`endpoint`,`metric`,`col`,`url`,`unit`,`title`,`grid_type`,`series_name`,`rate`,`agg_type`,`legend`) values (12,'','http.status',6,'/dashboard/chart','','http.status','line','metric',0,'avg','$metric');
insert  into `prom_metric`(`metric`,`metric_type`,`prom_ql`,`prom_main`) values ('http.status','http','http_status{guid=\"$guid\"}','');
insert  into `grp`(`name`,`create_at`) VALUES ('default_http_group',NOW());
insert  into `tpl`(`grp_id`,`endpoint_id`,`create_at`) SELECT id,0,NOW() FROM grp WHERE `name`='default_http_group';
insert  into `strategy`(`tpl_id`,`metric`,`expr`,`cond`,`last`,`priority`,`content`,`config_type`) SELECT id,'http.status','http_status{guid="$guid"}','<3','60s','medium','http check fail','' FROM tpl ORDER BY id DESC LIMIT 1;
insert  into `strategy`(`tpl_id`,`metric`,`expr`,`cond`,`last`,`priority`,`content`,`config_type`) SELECT id,'http.status','http_status{guid="$guid"}','>=300','60s','high','http response status problem','' FROM tpl ORDER BY id DESC LIMIT 1;
#@v1.5.0.6-end@;