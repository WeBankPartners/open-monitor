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

insert  into `strategy`(`id`,`tpl_id`,`metric`,`expr`,`cond`,`last`,`priority`,`content`,`config_type`) values (1,1,'cpu.used.percent','100-(avg by (instance) (rate(node_cpu_seconds_total{instance=\"$address\",mode=\"idle\"}[20s]))*100)','>85','60s','medium','cpu used high',''),(2,1,'mem.used.percent','100-((node_memory_MemFree_bytes{instance=\"$address\"}+node_memory_Cached_bytes{instance=\"$address\"}+node_memory_Buffers_bytes{instance=\"$address\"})/node_memory_MemTotal_bytes)*100','>85','60s','medium','memory used high',''),(3,1,'load.1min','node_load1{instance=\"$address\"}','>3','60s','medium','sys load busy',''),(4,1,'disk.read.bytes','irate(node_disk_read_bytes_total{instance=\"$address\"}[20s])','>10485760','60s','low','disk read busy',''),(5,1,'disk.write.bytes','irate(node_disk_written_bytes_total{instance=\"$address\"}[20s])','>10485760','60s','low','disk write busy',''),(6,1,'net.if.in.bytes','irate(node_network_receive_bytes_total{instance=\"$address\"}[20s])','>52428800','60s','medium','net flow is high',''),(7,1,'net.if.out.bytes','irate(node_network_transmit_bytes_total{instance=\"$address\"}[20s])','>52428800','60s','medium','net flow is high',''),(8,5,'heap.mem.used.percent','java_lang_Memory_HeapMemoryUsage_used{instance=\"$address\"}/java_lang_Memory_HeapMemoryUsage_max{instance=\"$address\"} * 100','>80','60s','medium','heap memory use high',''),(9,5,'gc.marksweep.time','increase(java_lang_GarbageCollector_CollectionTime{name=~\".*MarkSweep.*\",instance=\"$address\"}[2m])','>60','60s','medium','marksweep gc frequent occurrence',''),(10,3,'mysql.connect.used.percent','mysql_global_status_max_used_connections{instance=\"$address\"}/mysql_global_variables_max_connections{instance=\"$address\"} * 100','>80','60s','medium','mysql connection used too many',''),(11,3,'mysql.alive','mysql_up{instance=\"$address\"}','<1','30s','high','mysql instance down',''),(12,4,'redis.alive','redis_up{instance=\"$address\"}','<1','30s','high','redis instance down',''),(13,4,'redis.client.used.percent','redis_connected_clients{instance=\"$address\"}/redis_config_maxclients{instance=\"$address\"} * 100','>50','60s','medium','redis connected too many client ',''),(14,1,'cpu.used.percent','100-(avg by (instance) (rate(node_cpu_seconds_total{instance=\"$address\",mode=\"idle\"}[20s]))*100)','>95','60s','high','cpu high',''),(15,1,'ping_alive','ping_alive{guid=\"$guid\"}','==1','120s','high','ping fail',''),(16,6,'ping_alive','ping_alive{guid=\"$guid\"}','==1','120s','high','ping fail',''),(17,7,'telnet_alive','telnet_alive{guid=\"$guid\"}','>0','60s','high','telnet fail',''),(18,1,'mem.used.percent','100-((node_memory_MemFree_bytes{instance=\"$address\"}+node_memory_Cached_bytes{instance=\"$address\"}+node_memory_Buffers_bytes{instance=\"$address\"})/node_memory_MemTotal_bytes)*100','>50','60s','low','memory used over 50%','');

insert  into `user`(`name`,`passwd`,`display_name`,`email`,`phone`,`ext_contact_one`,`ext_contact_two`,`created`) values ('admin','yfBtw71KfVF7jBEyoX+ByyB+5FuEyjFYnd3CHXHR910=','admin','admin@xx.com','12345678910','','',NOW());

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

#@v1.6.0.3-begin@;
DROP TABLE IF EXISTS `alive_check_queue`;
CREATE TABLE `alive_check_queue` (
  `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
  `message` VARCHAR(255) NOT NULL,
  `update_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=INNODB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;
#@v1.6.0.3-end@;

#@v1.7.0.12-begin@;
ALTER TABLE log_monitor ADD COLUMN priority VARCHAR(50) DEFAULT 'high';
#@v1.7.0.12-end@;

#@v1.7.1.9-begin@;
DELETE FROM strategy WHERE tpl_id=1 AND metric='process_alive_count';
INSERT  INTO `strategy`(`tpl_id`,`metric`,`expr`,`cond`,`last`,`priority`,`content`,`config_type`) VALUE (1,'process_alive_count','node_process_monitor_count_current{instance=\"$address\"}','==0','30s','high','process down','');
DELETE FROM prom_metric WHERE metric='volume_used_percent';
INSERT  INTO `prom_metric`(`metric`,`metric_type`,`prom_ql`,`prom_main`) VALUES ('volume_used_percent','host','100-(node_filesystem_avail_bytes{fstype=~\"ext3|ext4|xfs\",instance=\"$address\"}/node_filesystem_size_bytes{fstype=~\"ext3|ext4|xfs\",instance=\"$address\"})*100','');
DELETE FROM strategy WHERE tpl_id=1 AND metric='volume_used_percent';
INSERT  INTO `strategy`(`tpl_id`,`metric`,`expr`,`cond`,`last`,`priority`,`content`,`config_type`) VALUE (1,'volume_used_percent','100-(node_filesystem_avail_bytes{fstype=~\"ext3|ext4|xfs\",instance=\"$address\"}/node_filesystem_size_bytes{fstype=~\"ext3|ext4|xfs\",instance=\"$address\"})*100','>=85','60s','high','disk space alert','');
#@v1.7.1.9-end@;

#@v1.7.1.12-begin@;
ALTER TABLE process_monitor ADD COLUMN display_name VARCHAR(255) DEFAULT '';
CREATE TABLE IF NOT EXISTS `db_monitor` (
  `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
  `endpoint_guid` VARCHAR(255) NOT NULL,
  `name` VARCHAR(255) DEFAULT '',
  `sql` VARCHAR(2000) DEFAULT '',
  `sys_panel` VARCHAR(50) DEFAULT '',
  `update_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=INNODB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;
DELETE FROM prom_metric WHERE metric='db_monitor_count';
INSERT  INTO `prom_metric`(`metric`,`metric_type`,`prom_ql`,`prom_main`) VALUES ('db_monitor_count','mysql','db_monitor_count{guid="$guid"}','');
#@v1.7.1.12-end@;

#@v1.7.1.15-begin@;
DELETE FROM panel WHERE title='DataMonitor' AND group_id=2;
DELETE FROM chart WHERE group_id=13 AND metric='db_monitor_count';
INSERT  INTO `panel`(`group_id`,`title`,`tags_enable`,`tags_url`,`tags_key`,`chart_group`,`auto_display`) VALUES (2,'DataMonitor',0,'','',13,0);
INSERT  INTO `chart`(`group_id`,`endpoint`,`metric`,`col`,`url`,`unit`,`title`,`grid_type`,`series_name`,`rate`,`agg_type`,`legend`) VALUES (13,'','db_monitor_count',6,'/dashboard/chart','','db_data_count','line','metric',0,'avg','$name');
DELETE FROM prom_metric WHERE metric='db_count_change';
INSERT  INTO `prom_metric`(`metric`,`metric_type`,`prom_ql`,`prom_main`) VALUES ('db_count_change','mysql','idelta(db_monitor_count{guid=\"$guid\"}[20s])','');
#@v1.7.1.15-end@;

#@v1.7.1.20-begin@;
CREATE TABLE `rel_role_custom_dashboard` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `role_id` INT UNSIGNED NOT NULL,
  `custom_dashboard_id` INT UNSIGNED NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=INNODB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;
#@v1.7.1.20-end@;

#@v1.8.0.3-begin@;
ALTER TABLE process_monitor CHANGE COLUMN `name` `tags` VARCHAR(255) DEFAULT '';
ALTER TABLE process_monitor ADD COLUMN process_name VARCHAR(255) DEFAULT '';
#@v1.8.0.3-end@;

#@v1.9.0-begin@;
ALTER TABLE role ADD COLUMN main_dashboard INT DEFAULT 0;
#@v1.9.0-end@;

#@v1.9.0.4-begin@;
alter table alarm_custom add column use_umg_policy varchar(50) default '';
alter table alarm_custom add column alert_way varchar(50) default '';
#@v1.9.0.4-end@;

#@v1.9.0.5-begin@;
update prom_metric set prom_ql='ping_alive{guid=\"$guid\",e_guid=\"$guid\"}' where metric='ping_alive';
update prom_metric set prom_ql='telnet_alive{guid=\"$guid\",e_guid=\"$guid\"}' where metric='telnet_alive';
update prom_metric set prom_ql='http_status{guid=\"$guid\",e_guid=\"$guid\"}' where metric='http.status';
insert into prom_metric(metric,metric_type,prom_ql) values ('ping_alive','ping','ping_alive{guid=\"$guid\",e_guid=\"$guid\"}'),('telnet_alive','telnet','telnet_alive{guid=\"$guid\",e_guid=\"$guid\"}');
update strategy set expr='ping_alive{guid=\"$guid\",e_guid=\"$guid\"}' where metric='ping_alive';
update strategy set expr='telnet_alive{guid=\"$guid\",e_guid=\"$guid\"}' where metric='telnet_alive';
update strategy set expr='http_status{guid=\"$guid\",e_guid=\"$guid\"}' where metric='http.status';
insert  into `dashboard`(`dashboard_type`,`search_enable`,`search_id`,`button_enable`,`button_group`,`message_enable`,`message_group`,`message_url`,`panels_enable`,`panels_type`,`panels_group`,`panels_param`) values ('ping',1,1,1,1,0,0,'',1,'tabs',9,'endpoint={endpoint}');
insert  into `panel`(`group_id`,`title`,`tags_enable`,`tags_url`,`tags_key`,`chart_group`,`auto_display`) values (9,'Ping',0,'','',14,0);
insert  into `chart`(`group_id`,`endpoint`,`metric`,`col`,`url`,`unit`,`title`,`grid_type`,`series_name`,`rate`,`agg_type`,`legend`) values (14,'','ping_alive',6,'/dashboard/chart','','ping.alive','line','metric',0,'avg','$metric');
#@v1.9.0.5-end@;

#@v1.9.0.12-begin@;
CREATE TABLE `kubernetes_cluster` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `cluster_name` varchar(50) NOT NULL unique,
  `api_server` varchar(100) NOT NULL,
  `token` text,
  `create_at` DATETIME DEFAULT NULL,
  `update_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;
CREATE TABLE `kubernetes_endpoint_rel` (
  `id` INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `kubernete_id` INT(10) UNSIGNED NOT NULL,
  `endpoint_guid` varchar(100) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;
insert  into `dashboard`(`dashboard_type`,`search_enable`,`search_id`,`button_enable`,`button_group`,`message_enable`,`message_group`,`message_url`,`panels_enable`,`panels_type`,`panels_group`,`panels_param`) values ('pod',1,1,1,1,0,0,'',1,'tabs',10,'endpoint={endpoint}');
insert  into `panel`(`group_id`,`title`,`tags_enable`,`tags_url`,`tags_key`,`chart_group`,`auto_display`) values (10,'pod',0,'','',15,0);
insert  into `chart`(`group_id`,`endpoint`,`metric`,`col`,`url`,`unit`,`title`,`grid_type`,`series_name`,`rate`,`agg_type`,`legend`) values (15,'','pod.cpu.used.percent',6,'/dashboard/chart','%','pod.cpu','line','metric',0,'p95','$metric');
insert  into `chart`(`group_id`,`endpoint`,`metric`,`col`,`url`,`unit`,`title`,`grid_type`,`series_name`,`rate`,`agg_type`,`legend`) values (15,'','pod.mem.used.percent',6,'/dashboard/chart','%','pod.memory','line','metric',0,'avg','$metric');
insert  into `prom_metric`(`metric`,`metric_type`,`prom_ql`,`prom_main`) values ('pod.cpu.used.percent','pod','rate(container_cpu_usage_seconds_total{pod=\"$pod\",job=\"k8s-cadvisor-$k8s_cluster\"}[60s])*100','');
insert  into `prom_metric`(`metric`,`metric_type`,`prom_ql`,`prom_main`) values ('pod.mem.used.percent','pod','100-(container_memory_usage_bytes{pod=\"$pod\",job=\"k8s-cadvisor-$k8s_cluster\"}/container_memory_max_usage_bytes{pod=\"$pod\",job=\"k8s-cadvisor-$k8s_cluster\"})*100','');
insert  into `grp`(`name`,`create_at`) VALUES ('default_pod_group',NOW());
insert  into `tpl`(`grp_id`,`endpoint_id`,`create_at`) SELECT id,0,NOW() FROM grp WHERE `name`='default_pod_group';
insert  into `strategy`(`tpl_id`,`metric`,`expr`,`cond`,`last`,`priority`,`content`,`config_type`) SELECT id,'pod.cpu.used.percent','rate(container_cpu_usage_seconds_total{pod="$pod",job="k8s-cadvisor-$k8s_cluster"}[60s])*100','>=85','90s','high','pod cpu used high','' FROM tpl ORDER BY id DESC LIMIT 1;
insert  into `strategy`(`tpl_id`,`metric`,`expr`,`cond`,`last`,`priority`,`content`,`config_type`) SELECT id,'pod.mem.used.percent','100-(container_memory_usage_bytes{pod="$pod",job="k8s-cadvisor-$k8s_cluster"}/container_memory_max_usage_bytes{pod="$pod",job="k8s-cadvisor-$k8s_cluster"})*100','>=85','90s','medium','pod memory used high','' FROM tpl ORDER BY id DESC LIMIT 1;
#@v1.9.0.12-end@;

#@v1.10.0.2-begin@;
DROP TABLE IF EXISTS `alert_window`;
CREATE TABLE `alert_window` (
  `id` INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `endpoint` varchar(50) NOT NULL,
  `start` varchar(50) NOT NULL,
  `end` varchar(50) NOT NULL,
  `weekday` varchar(50) NOT NULL,
  `update_user` varchar(50) NOT NULL,
  `update_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;
#@v1.10.0.2-end@;

#@v1.10.0.3-begin@;
DROP TABLE IF EXISTS `business_monitor_cfg`;
CREATE TABLE `business_monitor_cfg` (
  `id` INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `business_monitor_id` int NOT NULL,
  `regular` varchar(255) NOT NULL,
  `tags` varchar(255) default '',
  `string_map` text,
  `metric_config` text,
  `update_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;
#@v1.10.0.3-end@;

#@v1.10.0.5-begin@;
update chart set legend='$app_metric' where legend='$key';
#@v1.10.0.5-end@;

#@v1.11.0-begin@;
create index alarm_endpoint_idx on alarm(endpoint);
create index alarm_metric_idx on alarm(s_metric);
create index alarm_status_idx on alarm(status);
create index alarm_priority_idx on alarm(s_priority);
#@v1.11.0-end@;

#@v1.11.0.7-begin@;
alter table alarm_custom modify column alert_info varchar(2048) default '';
alter table alarm_custom modify column alert_ip varchar(50) default '';
alter table alarm_custom modify column alert_obj varchar(50) default '';
alter table alarm_custom modify column alert_title varchar(100) default '';
alter table alarm_custom modify column alert_reciver varchar(500) default '';
alter table alarm_custom modify column remark_info varchar(256) default '';
alter table alarm_custom modify column sub_system_id varchar(10) default '';
#@v1.11.0.7-end@;

#@v1.11.0.8-begin@;
alter table strategy add column notify_enable tinyint default 1;
alter table strategy add column notify_delay int default 0;
#@v1.11.0.8-end@;

#@v1.11.1.2-begin@;
update prom_metric set prom_ql='mysql_global_status_threads_connected{instance=\"$address\"}/mysql_global_variables_max_connections{instance=\"$address\"} * 100' where metric='mysql.connect.used.percent';
update strategy set expr='mysql_global_status_threads_connected{instance=\"$address\"}/mysql_global_variables_max_connections{instance=\"$address\"} * 100' where metric='mysql.connect.used.percent';
alter table role add column disable tinyint default 0;
#@v1.11.1.2-end@;

#@v1.11.2.1-begin@;
alter table log_monitor add column notify_enable tinyint default 1;
#@v1.11.2.1-end@;

#@v1.11.2.4-begin@;
alter table log_monitor add column owner_endpoint varchar(255) default '';
#@v1.11.2.4-end@;

#@v1.11.3-begin@;
alter table alarm drop index `alarm_unique_index`;
create index `alarm_unique_index_sec` on alarm (`strategy_id`,`endpoint`,`status`,`start`);
alter table alarm modify column tags varchar(1024) default '';
#@v1.11.3-end@;

#@v1.11.3.2-begin@;
alter table kubernetes_endpoint_rel add column pod_guid varchar(100) not null;
alter table kubernetes_endpoint_rel add column namespace varchar(50) not null;
#@v1.11.3.2-end@;

#@v1.11.3.3-begin@;
update prom_metric set prom_ql='sum by (pod,job)(rate(container_cpu_usage_seconds_total{pod=\"$pod\",job=\"k8s-cadvisor-$k8s_cluster\"}[60s])*100)' where metric='pod.cpu.used.percent';
update prom_metric set prom_ql='100-(sum by (pod,job)(container_memory_usage_bytes{pod=\"$pod\",job=\"k8s-cadvisor-$k8s_cluster\"})/sum by (pod,job)(container_memory_max_usage_bytes{pod=\"$pod\",job=\"k8s-cadvisor-$k8s_cluster\"}))*100' where metric='pod.mem.used.percent';
update strategy set expr='sum by (pod,job)(rate(container_cpu_usage_seconds_total{pod="$pod",job="k8s-cadvisor-$k8s_cluster"}[60s])*100)' where metric='pod.cpu.used.percent';
update strategy set expr='100-(sum by (pod,job)(container_memory_usage_bytes{pod="$pod",job="k8s-cadvisor-$k8s_cluster"})/sum by (pod,job)(container_memory_max_usage_bytes{pod="$pod",job="k8s-cadvisor-$k8s_cluster"}))*100' where metric='pod.mem.used.percent';
#@v1.11.3.3-end@;

#@v1.11.3.4-begin@;
alter table alarm_custom modify column alert_title  varchar(1024) default '';
alter table alarm drop index `alarm_unique_index_sec`;
create unique index `alarm_unique_index_sec` on alarm (`strategy_id`,`endpoint`,`status`,`start`);
#@v1.11.3.4-end@;

#@v1.11.4.2-begin@;
CREATE TABLE `snmp_exporter` (
  `id` varchar(255) NOT NULL PRIMARY KEY,
  `address` varchar(255) NOT NULL,
  `modules` varchar(255) default 'if_mib',
  `create_at` datetime,
  `update_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
CREATE TABLE `snmp_endpoint_rel` (
  `id` INT(10) UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
  `snmp_exporter` varchar(255) NOT NULL,
  `endpoint_guid` varchar(100) NOT NULL,
  `target` varchar(100) NOT NULL
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;
create unique index prom_metric_unique on prom_metric(metric,metric_type);
insert  into `grp`(`name`,`create_at`) VALUES ('default_snmp_group',NOW());
alter table grp add column endpoint_type varchar(100);
update grp set endpoint_type='host' where name='default_host_group';
update grp set endpoint_type='mysql' where name='default_mysql_group';
update grp set endpoint_type='redis' where name='default_redis_group';
update grp set endpoint_type='java' where name='default_java_group';
update grp set endpoint_type='ping' where name='default_ping_group';
update grp set endpoint_type='telnet' where name='default_telnet_group';
update grp set endpoint_type='nginx' where name='default_nginx_group';
update grp set endpoint_type='http' where name='default_http_group';
update grp set endpoint_type='pod' where name='default_pod_group';
update grp set endpoint_type='snmp' where name='default_snmp_group';
#@v1.11.4.2-end@;

#@v1.11.5.1-begin@;
alter table snmp_exporter add column scrape_interval int default 10;
alter table endpoint add column cluster varchar(64) default 'default';
#@v1.11.5.1-end@;

#@v1.11.5.2-begin@;
CREATE TABLE `cluster` (
  `id` varchar(255) NOT NULL PRIMARY KEY,
  `display_name` varchar(255),
  `remote_agent_address` varchar(255),
  `prom_address` varchar(255)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
insert into cluster value ('default','default','','');
#@v1.11.5.2-end@;

#@v1.12.0.1-begin@;
alter table alarm add column custom_message varchar(500);
alter table alarm_custom add column custom_message varchar(500);
#@v1.12.0.1-end@;

#@v1.12.1.1-begin@;
alter table endpoint add column tags varchar(255);
alter table alarm add column endpoint_tags varchar(255);
#@v1.12.1.1-end@;

#@v1.12.3.1-begin@;
alter table business_monitor_cfg add column config_type varchar(16) default 'json';
alter table business_monitor_cfg add column agg_type varchar(16) default 'avg';
#@v1.12.3.1-end@;

#@v1.13.0.1-begin@;
insert into grp(name,endpoint_type) value ('default_process_group','process');
insert  into `dashboard`(`dashboard_type`,`search_enable`,`search_id`,`button_enable`,`button_group`,`message_enable`,`message_group`,`message_url`,`panels_enable`,`panels_type`,`panels_group`,`panels_param`) values ('process',1,1,1,1,0,0,'',1,'tabs',11,'endpoint={endpoint}');
insert  into `panel`(`group_id`,`title`,`tags_enable`,`tags_url`,`tags_key`,`chart_group`,`auto_display`) values (11,'process',0,'','',18,0);
insert  into `chart`(`group_id`,`endpoint`,`metric`,`col`,`url`,`unit`,`title`,`grid_type`,`series_name`,`rate`,`agg_type`,`legend`) values (18,'','process_cpu_used_percent',6,'/dashboard/chart','','process.cpu.used','line','metric',0,'avg','$metric'),(18,'','process_mem_byte',6,'/dashboard/chart','','process.mem.used','line','metric',0,'avg','$metric');
insert  into `prom_metric`(`metric`,`metric_type`,`prom_ql`,`prom_main`) values ('process_alive_count','process','node_process_monitor_count_current{instance=\"$address\",process_guid=\"$guid\"}',''),('process_cpu_used_percent','process','node_process_monitor_cpu{instance=\"$address\",process_guid=\"$guid\"}',''),('process_mem_byte','process','node_process_monitor_mem{instance=\"$address\",process_guid=\"$guid\"}','');

CREATE TABLE `service_group` (
  `guid` varchar(64) NOT NULL PRIMARY KEY,
  `display_name` varchar(255) NOT NULL,
  `description` varchar(255),
  `parent` varchar(64) DEFAULT NULL,
  `service_type` varchar(32) NOT NULL,
  `update_time` varchar(32),
  CONSTRAINT `service_group_parent` FOREIGN KEY (`parent`) REFERENCES `service_group` (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `cluster_new` (
  `guid` varchar(128) NOT NULL PRIMARY KEY,
  `display_name` varchar(255),
  `remote_agent_address` varchar(255),
  `prometheus_address` varchar(255)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

insert into cluster_new(guid,display_name) value ('default','default');

CREATE TABLE `monitor_type` (
  `guid` varchar(32) NOT NULL PRIMARY KEY,
  `display_name` varchar(255),
  `description` varchar(255)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

insert into monitor_type(guid,display_name) value ('host','host'),('mysql','mysql'),('redis','redis'),('java','java'),('tomcat','tomcat'),('nginx','nginx'),('ping','ping'),('pingext','pingext'),('telnet','telnet'),('telnetext','telnetext'),('http','http'),('httpext','httpext'),('windows','windows'),('snmp','snmp'),('process','process'),('pod','pod');

CREATE TABLE `endpoint_new` (
  `guid` varchar(160) NOT NULL PRIMARY KEY,
  `name` varchar(96) NOT NULL,
  `ip` varchar(32),
  `monitor_type` varchar(32) NOT NULL,
  `agent_version` varchar(64),
  `agent_address` varchar(64),
  `step` int default 10,
  `endpoint_version` varchar(64),
  `endpoint_address` varchar(64),
  `cluster` varchar(128) NOT NULL,
  `alarm_enable` tinyint default 1,
  `tags` varchar(255),
  `extend_param` text,
  `description` varchar(255),
  `update_time` varchar(32),
  CONSTRAINT `endpoint_cluster` FOREIGN KEY (`cluster`) REFERENCES `cluster_new` (`guid`),
  CONSTRAINT `endpoint_monitor_type` FOREIGN KEY (`monitor_type`) REFERENCES `monitor_type` (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `endpoint_service_rel` (
  `guid` varchar(64) NOT NULL PRIMARY KEY,
  `endpoint` varchar(160) NOT NULL,
  `service_group` varchar(64) NOT NULL,
  CONSTRAINT `e_service_rel_e` FOREIGN KEY (`endpoint`) REFERENCES `endpoint_new` (`guid`),
  CONSTRAINT `e_service_rel_s` FOREIGN KEY (`service_group`) REFERENCES `service_group` (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `log_metric_monitor` (
  `guid` varchar(64) NOT NULL PRIMARY KEY,
  `service_group` varchar(64) NOT NULL,
  `log_path` varchar(255) NOT NULL,
  `metric_type` varchar(16) default 'json',
  `monitor_type` varchar(32) DEFAULT NULL,
  `update_time` varchar(32),
  CONSTRAINT `log_monitor_service_group` FOREIGN KEY (`service_group`) REFERENCES `service_group` (`guid`),
  CONSTRAINT `log_monitor_monitor_type` FOREIGN KEY (`monitor_type`) REFERENCES `monitor_type` (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `log_metric_json` (
  `guid` varchar(64) NOT NULL PRIMARY KEY,
  `log_metric_monitor` varchar(64) NOT NULL,
  `json_regular` varchar(255),
  `tags` varchar(64),
  `update_time` varchar(32),
  CONSTRAINT `log_monitor_json_monitor` FOREIGN KEY (`log_metric_monitor`) REFERENCES `log_metric_monitor` (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `log_metric_config` (
  `guid` varchar(64) NOT NULL PRIMARY KEY,
  `log_metric_monitor` varchar(64) DEFAULT NULL,
  `log_metric_json` varchar(64) DEFAULT NULL,
  `metric` varchar(64) NOT NULL,
  `display_name` varchar(255),
  `json_key` varchar(255),
  `regular` varchar(255),
  `step` int default 10,
  `agg_type` varchar(16) DEFAULT 'avg',
  `update_time` varchar(32),
  CONSTRAINT `log_monitor_config_monitor` FOREIGN KEY (`log_metric_monitor`) REFERENCES `log_metric_monitor` (`guid`),
  CONSTRAINT `log_monitor_config_json` FOREIGN KEY (`log_metric_json`) REFERENCES `log_metric_json` (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `log_metric_string_map` (
  `guid` varchar(64) NOT NULL PRIMARY KEY,
  `log_metric_config` varchar(64) NOT NULL,
  `source_value` varchar(255),
  `regulative` tinyint default 0,
  `target_value` varchar(64),
  `update_time` varchar(32),
  CONSTRAINT `log_monitor_string_config` FOREIGN KEY (`log_metric_config`) REFERENCES `log_metric_config` (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `log_metric_endpoint_rel` (
  `guid` varchar(64) NOT NULL PRIMARY KEY,
  `log_metric_monitor` varchar(64) NOT NULL,
  `source_endpoint` varchar(160),
  `target_endpoint` varchar(160),
  CONSTRAINT `log_monitor_endpoint_metric` FOREIGN KEY (`log_metric_monitor`) REFERENCES `log_metric_monitor` (`guid`),
  CONSTRAINT `log_monitor_endpoint_source` FOREIGN KEY (`source_endpoint`) REFERENCES `endpoint_new` (`guid`),
  CONSTRAINT `log_monitor_endpoint_target` FOREIGN KEY (`target_endpoint`) REFERENCES `endpoint_new` (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `db_metric_monitor` (
  `guid` varchar(64) NOT NULL PRIMARY KEY,
  `service_group` varchar(64) NOT NULL,
  `metric_sql` text,
  `metric` varchar(64) NOT NULL,
  `display_name` varchar(255),
  `step` int default 10,
  `monitor_type` varchar(32) DEFAULT NULL,
  `update_time` varchar(32),
  CONSTRAINT `db_monitor_service_group` FOREIGN KEY (`service_group`) REFERENCES `service_group` (`guid`),
  CONSTRAINT `db_monitor_monitor_type` FOREIGN KEY (`monitor_type`) REFERENCES `monitor_type` (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `db_metric_endpoint_rel` (
  `guid` varchar(64) NOT NULL PRIMARY KEY,
  `db_metric_monitor` varchar(64) NOT NULL,
  `source_endpoint` varchar(160),
  `target_endpoint` varchar(160),
  CONSTRAINT `db_monitor_endpoint_metric` FOREIGN KEY (`db_metric_monitor`) REFERENCES `db_metric_monitor` (`guid`),
  CONSTRAINT `db_monitor_endpoint_source` FOREIGN KEY (`source_endpoint`) REFERENCES `endpoint_new` (`guid`),
  CONSTRAINT `db_monitor_endpoint_target` FOREIGN KEY (`target_endpoint`) REFERENCES `endpoint_new` (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
#@v1.13.0.1-end@;

#@v1.13.0.2-begin@;
CREATE TABLE `endpoint_group` (
  `guid` varchar(64) NOT NULL PRIMARY KEY,
  `display_name` varchar(255) NOT NULL,
  `description` varchar(255),
  `monitor_type` varchar(32) NOT NULL,
  `service_group` varchar(64),
  `alarm_window` varchar(255),
  `update_time` varchar(32),
  CONSTRAINT `endpoint_group_monitor_type` FOREIGN KEY (`monitor_type`) REFERENCES `monitor_type` (`guid`),
  CONSTRAINT `endpoint_group_service_group` FOREIGN KEY (`service_group`) REFERENCES `service_group` (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `endpoint_group_rel` (
  `guid` varchar(64) NOT NULL PRIMARY KEY,
  `endpoint` varchar(160) NOT NULL,
  `endpoint_group` varchar(64),
  CONSTRAINT `endpoint_group_e` FOREIGN KEY (`endpoint`) REFERENCES `endpoint_new` (`guid`),
  CONSTRAINT `endpoint_group_g` FOREIGN KEY (`endpoint_group`) REFERENCES `endpoint_group` (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `role_new` (
  `guid` varchar(64) NOT NULL PRIMARY KEY,
  `display_name` varchar(255) NOT NULL,
  `email` varchar(255),
  `phone` varchar(32),
  `update_time` varchar(32)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `metric` (
  `guid` varchar(128) NOT NULL PRIMARY KEY,
  `metric` varchar(64) NOT NULL,
  `monitor_type` varchar(32) NOT NULL,
  `prom_expr` text,
  `tag_owner` varchar(64),
  `update_time` varchar(32),
  CONSTRAINT `metric_monitor_type` FOREIGN KEY (`monitor_type`) REFERENCES `monitor_type` (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `alarm_strategy` (
  `guid` varchar(64) NOT NULL PRIMARY KEY,
  `endpoint_group` varchar(64) NOT NULL,
  `metric` varchar(128) NOT NULL,
  `condition` varchar(32) NOT NULL,
  `last` varchar(16) NOT NULL,
  `priority` varchar(16) DEFAULT 'low',
  `content` text,
  `notify_enable` tinyint default 1,
  `notify_delay_second` int default 0,
  `update_time` varchar(32),
  CONSTRAINT `strategy_endpoint_group` FOREIGN KEY (`endpoint_group`) REFERENCES `endpoint_group` (`guid`),
  CONSTRAINT `strategy_metric` FOREIGN KEY (`metric`) REFERENCES `metric` (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `notify` (
  `guid` varchar(64) NOT NULL PRIMARY KEY,
  `endpoint_group` varchar(64),
  `service_group` varchar(64),
  `alarm_strategy` varchar(64),
  `alarm_action` varchar(32) default 'firing',
  `alarm_priority` varchar(32),
  `notify_num` int default 1,
  `proc_callback_name` varchar(64),
  `proc_callback_key` varchar(64),
  `callback_url` varchar(255),
  `callback_param` varchar(255),
  CONSTRAINT `notify_endpoint_group` FOREIGN KEY (`endpoint_group`) REFERENCES `endpoint_group` (`guid`),
  CONSTRAINT `notify_service_group` FOREIGN KEY (`service_group`) REFERENCES `service_group` (`guid`),
  CONSTRAINT `notify_alarm_strategy` FOREIGN KEY (`alarm_strategy`) REFERENCES `alarm_strategy` (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `notify_role_rel` (
  `guid` varchar(64) NOT NULL PRIMARY KEY,
  `notify` varchar(64) NOT NULL,
  `role` varchar(64) NOT NULL,
  CONSTRAINT `notify_role_n` FOREIGN KEY (`notify`) REFERENCES `notify` (`guid`),
  CONSTRAINT `notify_role_r` FOREIGN KEY (`role`) REFERENCES `role_new` (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

alter table alarm add column alarm_strategy varchar(64);
insert into role_new(guid,display_name,email) select name,display_name,email from role;
insert into endpoint_group(guid,display_name,description,monitor_type) select name,name,description,endpoint_type from grp where endpoint_type is not null;
insert into metric(guid,metric,monitor_type,prom_expr,tag_owner) select t1.* from (select CONCAT(metric,'__',metric_type),metric,metric_type,prom_ql,prom_main from prom_metric where metric_type<>'tomcat' union select CONCAT(metric,'__java'),metric,'java',prom_ql,prom_main from prom_metric where metric_type='tomcat') t1;
insert into alarm_strategy select CONCAT('old_',t1.id),t3.name,t4.guid ,t1.cond,t1.`last`,t1.priority,t1.content,t1.notify_enable,t1.notify_delay,'' as update_time from strategy t1 left join tpl t2 on t1.tpl_id=t2.id left join grp t3 on t2.grp_id=t3.id left join metric t4 on (t1.metric=t4.metric and t3.endpoint_type=t4.monitor_type) where t2.grp_id>0 and t4.guid is not null;
insert into endpoint_new(guid,name,ip,monitor_type,agent_address,step,endpoint_address,cluster) select guid,name,ip,export_type,address_agent,step,address,cluster from endpoint where address_agent<>'';
insert into endpoint_new(guid,name,ip,monitor_type,agent_address,step,endpoint_address,cluster) select guid,name,ip,export_type,address,step,address,cluster from endpoint where address_agent='';
insert into endpoint_group_rel select concat(t1.endpoint_id,'__',t1.grp_id),t2.guid,t3.name from grp_endpoint t1 left join endpoint t2 on t1.endpoint_id=t2.id left join grp t3 on t1.grp_id=t3.id where t2.guid is not null and t3.name is not null;
#@v1.13.0.2-end@;

#@v1.13.0.6-begin@;
CREATE TABLE `sys_parameter` (
  `guid` varchar(64) NOT NULL PRIMARY KEY,
  `param_key` varchar(64) NOT NULL,
  `param_value` text
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
alter table alarm_strategy drop foreign key strategy_metric;
insert into sys_parameter(guid,param_key,param_value) value ('metric_template_01','metric_template','{"name":"custom","prom_ql":"$a","param":"$a"}');
insert into sys_parameter(guid,param_key,param_value) value ('metric_template_02','metric_template','{"name":"percent01","prom_ql":"100*(sum($a)/(sum($b) > 0) or vector(0))","param":"$a,$b"}');
insert into sys_parameter(guid,param_key,param_value) value ('metric_template_03','metric_template','{"name":"percent02","prom_ql":"100*(1-(sum($a)/(sum($b) > 0) or vector(0)))","param":"$a,$b"}');
insert into sys_parameter(guid,param_key,param_value) value ('metric_template_04','metric_template','{"name":"percent03","prom_ql":"100*(sum($a)/((sum($a)+sum($b)) > 0) or vector(0))","param":"$a,$b"}');
insert into sys_parameter(guid,param_key,param_value) value ('metric_template_05','metric_template','{"name":"percent04","prom_ql":"100*((sum($a)-sum($b))/(sum($a) > 0) or vector(0))","param":"$a,$b"}');
#@v1.13.0.6-end@;

#@v1.13.0.18-begin@;
alter table service_group drop foreign key service_group_parent;
alter table log_metric_monitor drop foreign key log_monitor_service_group;
alter table db_metric_monitor drop foreign key db_monitor_service_group;
alter table metric add column log_metric_monitor varchar(64);
alter table metric add column db_metric_monitor varchar(64);
delete from panel where title='DataMonitor';
delete from prom_metric where metric='process_alive_count';
delete from alarm_strategy where metric like '%process_alive_count%';
insert into alarm_strategy(guid,endpoint_group,metric,`condition`,last,priority,content,notify_enable) value ('old_process_group','default_process_group','process_alive_count__process','==0','60s','high','process down',1);
#@v1.13.0.18-end@;

#@v1.13.0.21-begin@;
update prom_metric set prom_ql='node_process_monitor_count_current{process_guid="$guid"}' where metric_type='process' and metric='process_alive_count';
update prom_metric set prom_ql='node_process_monitor_cpu{process_guid="$guid"}' where metric_type='process' and metric='process_cpu_used_percent';
update prom_metric set prom_ql='node_process_monitor_mem{process_guid="$guid"}' where metric_type='process' and metric='process_mem_byte';
update metric set prom_expr='node_process_monitor_count_current{process_guid="$guid"}' where monitor_type='process' and metric='process_alive_count';
update metric set prom_expr='node_process_monitor_cpu{process_guid="$guid"}' where monitor_type='process' and metric='process_cpu_used_percent';
update metric set prom_expr='node_process_monitor_mem{process_guid="$guid"}' where monitor_type='process' and metric='process_mem_byte';
delete from strategy where id<=24;
delete from metric where metric='app.metric';
#@v1.13.0.21-end@;

#@v1.13.0.24-begin@;
CREATE TABLE `service_group_role_rel` (
  `guid` varchar(64) NOT NULL PRIMARY KEY,
  `service_group` varchar(64) NOT NULL,
  `role` varchar(64) NOT NULL,
  CONSTRAINT `service_role_s` FOREIGN KEY (`service_group`) REFERENCES `service_group` (`guid`),
  CONSTRAINT `service_role_r` FOREIGN KEY (`role`) REFERENCES `role_new` (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `log_keyword_monitor` (
  `guid` varchar(64) NOT NULL PRIMARY KEY,
  `service_group` varchar(64) NOT NULL,
  `log_path` varchar(255) NOT NULL,
  `monitor_type` varchar(32) NOT NULL,
  `update_time` varchar(32),
  CONSTRAINT `log_keyword_monitor_type` FOREIGN KEY (`monitor_type`) REFERENCES `monitor_type` (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `log_keyword_config` (
  `guid` varchar(64) NOT NULL PRIMARY KEY,
  `log_keyword_monitor` varchar(64) NOT NULL,
  `keyword` varchar(255) NOT NULL,
  `regulative` tinyint default 0,
  `notify_enable` tinyint default 1,
  `priority` varchar(16) default 'low',
  `update_time` varchar(32),
  CONSTRAINT `log_keyword_config_monitor` FOREIGN KEY (`log_keyword_monitor`) REFERENCES `log_keyword_monitor` (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `log_keyword_endpoint_rel` (
  `guid` varchar(64) NOT NULL PRIMARY KEY,
  `log_keyword_monitor` varchar(64) NOT NULL,
  `source_endpoint` varchar(160),
  `target_endpoint` varchar(160),
  CONSTRAINT `log_keyword_endpoint_monitor` FOREIGN KEY (`log_keyword_monitor`) REFERENCES `log_keyword_monitor` (`guid`),
  CONSTRAINT `log_keyword_endpoint_source` FOREIGN KEY (`source_endpoint`) REFERENCES `endpoint_new` (`guid`),
  CONSTRAINT `log_keyword_endpoint_target` FOREIGN KEY (`target_endpoint`) REFERENCES `endpoint_new` (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
#@v1.13.0.24-end@;

#@v1.13.0.25-begin@;
alter table metric drop column log_metric_monitor;
alter table metric drop column db_metric_monitor;
alter table metric add column service_group varchar(64);
alter table metric drop foreign key metric_monitor_type;
alter table panel add column service_group varchar(64) default null;
alter table metric add column workspace varchar(16) default 'any_object';
delete from sys_parameter;
insert into sys_parameter(guid,param_key,param_value) value ('metric_template_01','metric_template','{"name":"custom","prom_expr":"@a","param":"@a"}');
insert into sys_parameter(guid,param_key,param_value) value ('metric_template_02','metric_template','{"name":"percent[a/b]","prom_expr":"100*(sum(@a)/sum(@b) > 0 or vector(0))","param":"@a,@b"}');
insert into sys_parameter(guid,param_key,param_value) value ('metric_template_03','metric_template','{"name":"percent[(a-b)/a]","prom_expr":"100*((sum(@a)-sum(@b))/sum(@a) > 0 or vector(0))","param":"@a,@b"}');
insert into sys_parameter(guid,param_key,param_value) value ('metric_template_04','metric_template','{"name":"percent[a/(a+b)]","prom_expr":"100*(sum(@a)/((sum(@a)+sum(@b)) > 0) or vector(0))","param":"@a,@b"}');
insert into sys_parameter(guid,param_key,param_value) value ('metric_template_05','service_metric_template','{"name":"custom","prom_expr":"@a","param":"@a"}');
insert into sys_parameter(guid,param_key,param_value) value ('metric_template_06','service_metric_template','{"name":"sum(a)","prom_expr":"sum(@a)","param":"@a"}');
insert into sys_parameter(guid,param_key,param_value) value ('metric_template_07','service_metric_template','{"name":"avg(a)","prom_expr":"avg(@a)","param":"@a"}');
insert into sys_parameter(guid,param_key,param_value) value ('metric_template_08','service_metric_template','{"name":"max(a)","prom_expr":"max(@a)","param":"@a"}');
insert into sys_parameter(guid,param_key,param_value) value ('metric_template_09','service_metric_template','{"name":"min(a)","prom_expr":"min(@a)","param":"@a"}');
insert into sys_parameter(guid,param_key,param_value) value ('metric_template_10','service_metric_template','{"name":"percent[sum(a)/sum(b)]","prom_expr":"100*(sum(@a)/(sum(@b) > 0) or vector(0))","param":"@a,@b"}');
insert into sys_parameter(guid,param_key,param_value) value ('metric_template_11','service_metric_template','{"name":"percent[(sum(a)-sum(b))/sum(a)]","prom_expr":"100*((sum(@a)-sum(@b))/(sum(@a) > 0) or vector(0))","param":"@a,@b"}');
insert into sys_parameter(guid,param_key,param_value) value ('metric_template_12','service_metric_template','{"name":"percent[sum(a)/(sum(a)+sum(b))]","prom_expr":"100*(sum(@a)/((sum(@a)+sum(@b)) > 0) or vector(0))","param":"@a,@b"}');
#@v1.13.0.25-end@;

#@v2.0.0-begin@;
insert into notify(guid,endpoint_group,alarm_action) select concat(guid,'_firing'),guid,'firing' from endpoint_group;
insert into notify(guid,endpoint_group,alarm_action) select concat(guid,'_ok'),guid,'ok' from endpoint_group;
insert into service_group(guid,display_name,service_type) select guid,display_name,obj_type from panel_recursive where parent='';
insert into service_group(guid,display_name,service_type,parent) select guid,display_name,obj_type,parent from panel_recursive where parent<>'';
#@v2.0.0-end@;

#@v2.0.0.5-begin@;
alter table prom_metric modify prom_ql varchar(4096);
#@v2.0.0.5-end@;

#@v2.0.0.11-begin@;
alter table alarm modify content text;
#@v2.0.0.11-end@;

#@v2.0.0.12-begin@;
update alarm_strategy set update_time=now() where update_time is null or update_time='';
alter table alarm_strategy modify update_time datetime;
update endpoint_new set update_time=now() where update_time is null or update_time='';
alter table endpoint_new modify update_time datetime;
update endpoint_group set update_time=now() where update_time is null or update_time='';
alter table endpoint_group modify update_time datetime;
update service_group set update_time=now() where update_time is null or update_time='';
alter table service_group modify update_time datetime;
#@v2.0.0.12-end@;

#@v2.0.0.16-begin@;
alter table alarm_strategy add column active_window varchar(32) default '00:00-23:59';
#@v2.0.0.16-end@;

#@v2.0.0.29-begin@;
alter table rel_role_custom_dashboard add column permission varchar(32) default 'mgmt';
#@v2.0.0.29-end@;

#@v2.0.0.35-begin@;
alter table alarm drop index `alarm_unique_index_sec`;
alter table alarm modify endpoint_tags varchar(64);
update alarm set endpoint_tags=null where endpoint_tags='';
create unique index `alarm_tags_unique_idx` on alarm (`endpoint_tags`);
#@v2.0.0.35-end@;

#@v2.0.0.46-begin@;
alter table log_keyword_config add column content text default null;
#@v2.0.0.46-end@;

#@v2.0.1.11-begin@;
alter table log_metric_config add column tag_config text default null;
INSERT INTO metric (guid,metric,monitor_type,prom_expr,tag_owner,update_time,service_group,workspace) VALUES ('ping_loss__host','ping_loss','host','ping_loss_percent{guid="$guid",e_guid="$guid"}','',NULL,NULL,'any_object'),('ping_loss__ping','ping_loss','ping','ping_loss_percent{guid="$guid",e_guid="$guid"}','',NULL,NULL,'any_object');
INSERT INTO alarm_strategy (guid,endpoint_group,metric,`condition`,`last`,priority,content,notify_enable,notify_delay_second,update_time,active_window) VALUES ('new_host_ping_loss','default_host_group','ping_loss__host','>70','60s','high','ping packet loss',1,0,'2023-04-23 10:11:34','00:00-23:59'),('new_ping_ping_loss','default_ping_group','ping_loss__ping','>70','60s','high','ping packet loss',1,0,'2023-04-23 10:11:34','00:00-23:59');
#@v2.0.1.11-end@;

#@v2.0.2.35-begin@;
alter table agent_manager add column agent_remote_port varchar(255) default null;
#@v2.0.2.35-end@;

#@v2.0.4.7-begin@;
INSERT INTO metric (guid,metric,monitor_type,prom_expr,tag_owner,update_time,service_group,workspace) VALUES ('file.handler.free.percent__host','file.handler.free.percent','host','node_filesystem_files_free{instance="$address",mountpoint ="/",fstype="rootfs"} / node_filesystem_files{instance="$address",mountpoint ="/",fstype="rootfs"} * 100',NULL,'2023-12-12 17:22:09',NULL,'any_object');
#@v2.0.4.7-end@;

#@v2.0.5.3-begin@;
CREATE TABLE `log_keyword_alarm` (
     `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
     `alarm_id` int(11) NOT NULL,
     `endpoint` varchar(255) NOT NULL,
     `status` varchar(20) NOT NULL,
     `content` text,
     `tags` varchar(1024) DEFAULT '',
     `start_value` double DEFAULT NULL,
     `end_value` double DEFAULT NULL,
     `updated_time` datetime DEFAULT NULL,
     PRIMARY KEY (`id`),
     KEY `idx_log_keyword_alarm_id` (`alarm_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;
insert into log_keyword_alarm(alarm_id,endpoint,status,content,tags,start_value,end_value,updated_time) select id,endpoint,status,content,tags,start_value,end_value,`start` from alarm where s_metric='log_monitor' and id in (select max(id) from alarm where s_metric='log_monitor' group by tags);
#@v2.0.5.3-end@;

#@v2.0.6.1-begin@;
alter table notify add column proc_callback_mode varchar(16) default 'manual' comment '回调模式 -> manual(手动) | auto(自动)';
alter table notify add column description text default null;
alter table alarm add column notify_id varchar(64) default null;
alter table custom_dashboard add column panel_groups text default null;
alter table endpoint_metric modify column metric varchar(1024) default null;
#@v2.0.6.1-end@;

#@v2.0.6.10-begin@;
CREATE TABLE `alarm_notify` (
    `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
    `alarm_id` int(11) NOT NULL,
    `notify_id` varchar(64) NOT NULL,
    `endpoint` varchar(255) DEFAULT NULL,
    `metric` varchar(255) DEFAULT NULL,
    `status` varchar(20) NOT NULL,
    `proc_def_key` varchar(255) DEFAULT NULL,
    `proc_def_name` varchar(255) DEFAULT NULL,
    `notify_description` text DEFAULT NULL,
    `proc_ins_id` varchar(64) DEFAULT NULL,
    `created_user` varchar(64) DEFAULT NULL,
    `created_time` datetime DEFAULT NULL,
    `updated_time` datetime DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;
#@v2.0.6.10-end@;

#@v2.0.7.1-begin@;
CREATE TABLE `log_monitor_template` (
    `guid` varchar(64) NOT NULL COMMENT '唯一标识',
    `name` varchar(255) NOT NULL COMMENT '名称',
    `log_type` varchar(32) NOT NULL COMMENT '日志类型->json(json格式) | regular(通用正则格式)',
    `json_regular` varchar(255) NOT NULL COMMENT '提取json正则',
    `demo_log` text DEFAULT NULL COMMENT '样例日志',
    `calc_result` text DEFAULT NULL COMMENT '试算结果',
    `create_user` varchar(32) DEFAULT NULL COMMENT '创建人',
    `update_user` varchar(32) DEFAULT NULL COMMENT '更新人',
    `create_time` datetime DEFAULT NULL COMMENT '创建时间',
    `update_time` datetime DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `log_monitor_template_role` (
     `guid` varchar(64) NOT NULL COMMENT '唯一标识',
     `log_monitor_template` varchar(64) NOT NULL COMMENT '业务监控模版',
     `role` varchar(64) NOT NULL COMMENT '角色',
     `permission` varchar(32) DEFAULT NULL COMMENT '权限->USE(使用) | MGMT(管理)',
     PRIMARY KEY (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `log_param_template` (
      `guid` varchar(64) NOT NULL COMMENT '唯一标识',
      `log_monitor_template` varchar(64) NOT NULL COMMENT '业务监控模版',
      `name` varchar(64) NOT NULL COMMENT '名称',
      `display_name` varchar(255) DEFAULT NULL COMMENT '指标显示名',
      `json_key` varchar(64) DEFAULT NULL COMMENT 'json的key',
      `regular` varchar(255) DEFAULT NULL COMMENT '正则',
      `demo_match_value` varchar(255) DEFAULT NULL COMMENT '匹配结果',
      `create_user` varchar(32) DEFAULT NULL COMMENT '创建人',
      `update_user` varchar(32) DEFAULT NULL COMMENT '更新人',
      `create_time` datetime DEFAULT NULL COMMENT '创建时间',
      `update_time` datetime DEFAULT NULL COMMENT '更新时间',
      PRIMARY KEY (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `log_metric_template` (
   `guid` varchar(64) NOT NULL COMMENT '唯一标识',
   `log_monitor_template` varchar(64) NOT NULL COMMENT '业务监控模版',
   `log_param_name` varchar(64) DEFAULT NULL COMMENT '日志参数名',
   `metric` varchar(64) NOT NULL COMMENT '指标',
   `display_name` varchar(255) DEFAULT NULL COMMENT '指标显示名',
   `step` int(11) DEFAULT 10 COMMENT '间隔',
   `agg_type` varchar(16) DEFAULT 'agg' COMMENT '聚合类型',
   `tag_config` text DEFAULT NULL COMMENT '标签配置',
   `create_user` varchar(32) DEFAULT NULL COMMENT '创建人',
   `update_user` varchar(32) DEFAULT NULL COMMENT '更新人',
   `create_time` datetime DEFAULT NULL COMMENT '创建时间',
   `update_time` datetime DEFAULT NULL COMMENT '更新时间',
   PRIMARY KEY (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `log_metric_group` (
    `guid` varchar(64) NOT NULL COMMENT '唯一标识',
    `name` varchar(64) NOT NULL COMMENT '名称',
    `log_type` varchar(32) NOT NULL COMMENT '日志类型->json(json格式) | regular(通用正则格式) | custom(自定义)',
    `log_metric_monitor` varchar(64) NOT NULL COMMENT '业务日志监控',
    `log_monitor_template` varchar(64) DEFAULT NULL COMMENT '引用模版',
    `demo_log` text DEFAULT NULL COMMENT '样例日志',
    `calc_result` text DEFAULT NULL COMMENT '试算结果',
    `create_user` varchar(32) DEFAULT NULL COMMENT '创建人',
    `update_user` varchar(32) DEFAULT NULL COMMENT '更新人',
    `create_time` datetime DEFAULT NULL COMMENT '创建时间',
    `update_time` datetime DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `log_metric_param` (
    `guid` varchar(64) NOT NULL COMMENT '唯一标识',
    `name` varchar(64) NOT NULL COMMENT '名称',
    `display_name` varchar(255) DEFAULT NULL COMMENT '参数名',
    `log_metric_group` varchar(64) NOT NULL COMMENT '业务指标组',
    `regular` varchar(255) DEFAULT NULL COMMENT '提取正则',
    `demo_match_value` varchar(255) DEFAULT NULL COMMENT '匹配结果',
    `create_user` varchar(32) DEFAULT NULL COMMENT '创建人',
    `update_user` varchar(32) DEFAULT NULL COMMENT '更新人',
    `create_time` datetime DEFAULT NULL COMMENT '创建时间',
    `update_time` datetime DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

alter table log_metric_monitor add column `create_user` varchar(32) DEFAULT NULL COMMENT '创建人';
alter table log_metric_monitor add column `update_user` varchar(32) DEFAULT NULL COMMENT '更新人';
alter table log_metric_monitor add column `create_time` datetime DEFAULT NULL COMMENT '创建时间';
alter table log_metric_monitor modify column `update_time` datetime DEFAULT NULL COMMENT '更新时间';

alter table log_metric_config add column `log_metric_group` varchar(64) DEFAULT NULL COMMENT '业务指标组';
alter table log_metric_config add column `log_param_name` varchar(64) DEFAULT NULL COMMENT '日志参数名';
alter table log_metric_config add column `create_user` varchar(32) DEFAULT NULL COMMENT '创建人';
alter table log_metric_config add column `update_user` varchar(32) DEFAULT NULL COMMENT '更新人';
alter table log_metric_config add column `create_time` datetime DEFAULT NULL COMMENT '创建时间';

alter table log_metric_string_map add column `log_metric_group` varchar(64) DEFAULT NULL COMMENT '业务指标组';
alter table log_metric_string_map add column `log_param_name` varchar(64) DEFAULT NULL COMMENT '日志参数名';
alter table log_metric_string_map add column `value_type` varchar(64) DEFAULT NULL COMMENT '值的类型,成功失败等';

insert into log_metric_group(guid,name,log_type,log_metric_monitor,create_user,create_time,update_user,update_time) select concat('lmg_',guid),display_name,'custom',log_metric_monitor,'old_data',now(),'old_data',now() from log_metric_config where log_metric_json is null;
insert into log_metric_param(guid,name,display_name,log_metric_group,regular,create_user,create_time) select concat('lmp_',guid),metric,display_name,concat('lmg_',guid),regular,'old_data',now() from log_metric_config where log_metric_json is null;
update log_metric_config set log_metric_group=concat('lmg_',guid),log_param_name=metric,update_user='old_data' where log_metric_group is null and log_metric_json is null;
ALTER TABLE log_metric_string_map DROP FOREIGN KEY log_monitor_string_config;
alter table log_metric_group add column metric_prefix_code varchar(64) default null comment '指标前缀';
#@v2.0.7.1-end@;

#@v2.0.8.1-begin@;
CREATE TABLE `alarm_strategy_metric` (
                                         `guid` varchar(64) NOT null COMMENT '唯一标识',
                                         `alarm_strategy` varchar(64) NOT NULL COMMENT '告警配置表',
                                         `metric` varchar(128) NOT NULL COMMENT '指标',
                                         `condition` varchar(32) NOT NULL COMMENT '条件',
                                         `last` varchar(16) NOT NULL COMMENT '持续时间',
                                         `crc_hash` varchar(64) DEFAULT NULL COMMENT 'hash',
                                         `create_time` datetime DEFAULT NULL COMMENT '创建时间',
                                         `update_time` datetime DEFAULT NULL COMMENT '更新时间',
                                         PRIMARY KEY (`guid`),
                                         KEY `idx__strategy_metric__metric` (`metric`),
                                         CONSTRAINT `fk__strategy_metric__alarm_strategy` FOREIGN KEY (`alarm_strategy`) REFERENCES `alarm_strategy` (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `alarm_strategy_tag` (
                                      `guid` varchar(64) NOT NULL COMMENT '唯一标识',
                                      `alarm_strategy_metric` varchar(64) NOT NULL COMMENT '告警配置指标',
                                      `name` varchar(64) NOT NULL COMMENT '标签名',
                                      PRIMARY KEY (`guid`),
                                      CONSTRAINT `fk__strategy_tag__metric` FOREIGN KEY (`alarm_strategy_metric`) REFERENCES `alarm_strategy_metric` (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `alarm_strategy_tag_value` (
                                            `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
                                            `alarm_strategy_tag` varchar(64) NOT NULL COMMENT '告警配置标签值',
                                            `value` varchar(255) DEFAULT NULL COMMENT '标签值',
                                            PRIMARY KEY (`id`),
                                            CONSTRAINT `fk__strategy_tag_value__tag` FOREIGN KEY (`alarm_strategy_tag`) REFERENCES `alarm_strategy_tag` (`guid`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;
alter table alarm_strategy add column name varchar(128) default null comment '名称';
alter table rel_role_custom_dashboard rename to custom_dashboard_role_rel;
alter table custom_dashboard modify column `name` varchar(255) NOT NULL;
alter table metric add column log_metric_config varchar(64) default null;
alter table metric add column log_metric_template varchar(64) default null;
update alarm_strategy set name=substring_index(metric,'__',1) where name is null;

CREATE TABLE `alarm_condition` (
                                   `guid` varchar(64) NOT NULL COMMENT '唯一标识',
                                   `alarm_strategy` varchar(64) DEFAULT NULL COMMENT '告警配置表',
                                   `endpoint` varchar(255) NOT NULL COMMENT '监控对象',
                                   `status` varchar(20) NOT NULL COMMENT '状态',
                                   `metric` varchar(255) NOT NULL COMMENT '指标',
                                   `expr` varchar(500) NOT NULL COMMENT '指标表达式',
                                   `cond` varchar(50) NOT NULL COMMENT '条件',
                                   `last` varchar(50) NOT NULL COMMENT '持续时间',
                                   `priority` varchar(50) NOT NULL COMMENT '级别',
                                   `crc_hash` varchar(64) DEFAULT NULL COMMENT '告警配置hash',
                                   `tags` varchar(1024) DEFAULT NULL COMMENT '告警标签',
                                   `start_value` double DEFAULT NULL COMMENT '异常值',
                                   `start` datetime DEFAULT NULL COMMENT '异常时间',
                                   `end_value` double DEFAULT NULL COMMENT '恢复值',
                                   `end` datetime DEFAULT NULL COMMENT '恢复时间',
                                   `unique_hash` varchar(64) DEFAULT NULL COMMENT '告警唯一hash',
                                   PRIMARY KEY (`guid`),
                                   UNIQUE KEY `alarm_condition_unique_idx` (`unique_hash`),
                                   KEY `alarm_condition_endpoint_idx` (`endpoint`),
                                   KEY `alarm_condition_metric_idx` (`metric`),
                                   KEY `alarm_condition_status_idx` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `alarm_condition_rel` (
                                       `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
                                       `alarm` int(11) unsigned NOT NULL COMMENT '告警id',
                                       `alarm_condition` varchar(64) DEFAULT NULL COMMENT '条件id',
                                       PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

alter table alarm add column alarm_name varchar(64) default null comment '告警名称';
CREATE TABLE `remote_write_config` (
                                       `id` varchar(64) NOT NULL,
                                       `address` varchar(255) NOT NULL,
                                       `create_user` varchar(64) DEFAULT NULL,
                                       `update_user` varchar(64) DEFAULT NULL,
                                       `create_at` datetime DEFAULT NULL,
                                       `update_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                       PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
alter table alarm_strategy add column update_user varchar(64) default null;


CREATE TABLE IF NOT EXISTS `main_dashboard` (
    `guid` varchar(64) NOT NULL,
    `role_id` varchar(64) NOT NULL,
    `custom_dashboard` int(11)  unsigned not null COMMENT '首页看板表',
    PRIMARY KEY (`guid`),
    CONSTRAINT `fore_main_dashboard_custom_dashboard` FOREIGN KEY (`custom_dashboard`) REFERENCES `custom_dashboard` (`id`)
    )ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '首页看板表';


CREATE TABLE IF NOT EXISTS `custom_chart` (
    `guid` varchar(64) NOT NULL,
    `source_dashboard` int(11) NOT NULL COMMENT '源看板',
    `public` tinyint(1) default 0 COMMENT '是否公共',
    `name` varchar(255) default null COMMENT '图表名称',
    `chart_type` varchar(32) not null COMMENT '曲线图/饼图,line//pie',
    `line_type` varchar(32) not null COMMENT '折线/柱状/面积,line/bar/area',
    `aggregate` varchar(32) not null COMMENT '聚合类型',
    `agg_step` int(11) default 0  COMMENT '聚合间隔',
    `unit` varchar(32) default null COMMENT '单位',
    `create_user` varchar(64)  default null COMMENT '创建人',
    `update_user` varchar(64) default null COMMENT '更新人',
    `create_time` datetime  default null COMMENT '创建时间',
    `update_time` datetime  default null COMMENT '更新时间',
    `chart_template` varchar(100)  default null COMMENT '图表模板',
    PRIMARY KEY (`guid`)
    )ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '自定义看板图表表';



CREATE TABLE IF NOT EXISTS `custom_dashboard_chart_rel` (
    `guid` varchar(64) NOT NULL,
    `custom_dashboard` int(11) unsigned NOT NULL COMMENT '所属看板',
    `dashboard_chart` varchar(64)  not null COMMENT '所属看板图表',
    `group` varchar(64)  null COMMENT '所属分组',
    `display_config` text  null COMMENT '视图位置与长宽',
    `create_user` varchar(64)  null COMMENT '创建人',
    `updated_user` varchar(64)  null COMMENT '更新人',
    `create_time` datetime  null COMMENT '创建时间',
    `update_time` datetime  null COMMENT '更新时间',
    PRIMARY KEY (`guid`),
    CONSTRAINT `fore_custom_dashboard_chart_rel_custom_dashboard` FOREIGN KEY (`custom_dashboard`) REFERENCES `custom_dashboard` (`id`),
    CONSTRAINT `fore_custom_dashboard_chart_rel_dashboard_chart` FOREIGN KEY (`dashboard_chart`) REFERENCES `custom_chart` (`guid`)
    )ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '自定义看板图表关系表';


CREATE TABLE IF NOT EXISTS `custom_chart_series` (
    `guid` varchar(64) NOT NULL,
    `dashboard_chart` varchar(64) NOT NULL COMMENT '所属看板图表',
    `endpoint` varchar(255) NOT NULL COMMENT '监控对象',
    `service_group` varchar(64) NULL COMMENT '层级对象',
    `endpoint_name` varchar(255) NULL COMMENT '层级对象',
    `monitor_type` varchar(32) NULL COMMENT '监控类型',
    `metric` varchar(128) NULL COMMENT '指标',
    `color_group` varchar(32) NULL COMMENT '默认色系',
    PRIMARY KEY (`guid`),
    CONSTRAINT `fore_custom_chart_series_dashboard_chart` FOREIGN KEY (`dashboard_chart`) REFERENCES `custom_chart` (`guid`)
    )ENGINE= InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '自定义看板图表关系表';



CREATE TABLE IF NOT EXISTS `custom_chart_permission` (
    `guid` varchar(64) NOT NULL,
    `dashboard_chart` varchar(64) NOT NULL COMMENT '所属看板图表',
    `role_id` varchar(64) NOT NULL COMMENT '角色名',
    `permission` varchar(32)  NULL COMMENT '权限,mgmt/use',
    PRIMARY KEY (`guid`),
    CONSTRAINT `fore_custom_chart_permission_dashboard_chart` FOREIGN KEY (`dashboard_chart`) REFERENCES `custom_chart` (`guid`)
    )ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '公共图表权限表';



CREATE TABLE IF NOT EXISTS `custom_chart_series_config` (
    `guid` varchar(64) NOT NULL,
    `dashboard_chart_config` varchar(64) NOT NULL COMMENT '图表配置表',
    `tags` text  NULL COMMENT '标签',
    `color` varchar(32)  NULL COMMENT '颜色',
    `series_name` varchar(255)  NULL COMMENT '指标+对象+标签值',
    PRIMARY KEY (`guid`),
    CONSTRAINT `fore_custom_chart_series_config_dashboard_chart_config` FOREIGN KEY (`dashboard_chart_config`) REFERENCES `custom_chart_series` (`guid`)
    )ENGINE= InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '自定义图表配置表';


CREATE TABLE IF NOT EXISTS `custom_chart_series_tag` (
    `guid` varchar(64) NOT NULL,
    `dashboard_chart_config` varchar(64) NOT NULL COMMENT '图表配置表',
    `name` varchar(64)  NOT NULL COMMENT '标签名',
    PRIMARY KEY (`guid`),
    CONSTRAINT `fore_custom_chart_series_tag_dashboard_chart_config` FOREIGN KEY (`dashboard_chart_config`) REFERENCES `custom_chart_series` (`guid`)
    )ENGINE = InnoDB DEFAULT CHARSET =utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '自定义图表标签配置表';



CREATE TABLE IF NOT EXISTS `custom_chart_series_tagvalue` (
    `id` int(11) NOT NULL AUTO_INCREMENT,
    `dashboard_chart_tag` varchar(64) NOT NULL COMMENT '所属图表标签',
    `value` varchar(255)  NULL COMMENT '标签值',
    PRIMARY KEY (`id`),
    CONSTRAINT `fore_custom_chart_series_tagvalue_dashboard_chart_tag` FOREIGN KEY (`dashboard_chart_tag`) REFERENCES `custom_chart_series_tag` (`guid`)
    )ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '自定义图表标签值配置表';


alter table custom_dashboard_role_rel MODIFY role_id varchar(64)  NOT NULL COMMENT '角色id';
alter table metric add column log_metric_group varchar(64)  default NULL COMMENT '业务日志指标组配置Id';
alter table metric add column create_time varchar(64)  default NULL COMMENT '创建时间';
alter table metric add column create_user varchar(64)  default NULL COMMENT '创建人';
alter table metric add column update_user varchar(64)  default NULL COMMENT '更新人';
alter table custom_chart add column pie_type varchar(32)  default NULL COMMENT '饼图类型';
alter table custom_chart_series add column pie_display_tag varchar(64)  default NULL COMMENT '饼图展示标签';
alter table custom_chart_series add column endpoint_type varchar(64)  default NULL COMMENT '对象类型';
alter table custom_chart_series add column metric_type varchar(64)  default NULL COMMENT '指标类型';
alter table custom_chart_series add column metric_guid varchar(64)  default NULL COMMENT '指标Id';
alter table custom_dashboard add column time_range int(11)  default -3600 COMMENT '时间段';
alter table custom_dashboard add column refresh_week int(5)  default 10 COMMENT '定时刷新';

update custom_dashboard_role_rel set role_id= (select name from role where id = role_id) where role_id in (select id from role);
insert into main_dashboard (guid,role_id,custom_dashboard)(select uuid(),name,main_dashboard from `role` where main_dashboard > 0 and name!='' and main_dashboard in (select id from custom_dashboard));
alter table log_keyword_config add column name varchar(64) default null;
alter table log_metric_template modify `agg_type` varchar(255) DEFAULT 'agg' COMMENT '聚合类型';
alter table log_metric_string_map modify `log_metric_config` varchar(64) DEFAULT null;
alter table alarm modify column s_expr varchar(4096) default null;
#@v2.0.8.1-end@;

#@v3.0.4.1-begin@;
CREATE TABLE  `metric_comparison` (
      `guid` varchar(64) NOT NULL,
      `comparison_type` varchar(64) default null COMMENT '对比类型: day 日环比 week 周环比 month 月环比',
      `calc_type` varchar(64) default null COMMENT '计算数值: diff 差值,diff_percent 差值百分比',
      `calc_method` varchar(64) default null COMMENT '计算方法: avg平均,sum求和',
      `calc_period` int(11) default 0 COMMENT '计算周期',
      `metric_id` varchar(128) default null COMMENT '指标id',
      `origin_metric_id` varchar(128) default null COMMENT '原始指标id',
      `origin_metric` varchar(64) default null COMMENT '原始指标名称',
      `origin_prom_expr` tinytext default null COMMENT '原始指标表达式',
      `create_user` varchar(64)  default null COMMENT '创建人',
      `create_time` datetime  default null COMMENT '创建时间',
      PRIMARY KEY (`guid`),
      CONSTRAINT `fore_metric_comparison_metric_id` FOREIGN KEY (`metric_id`) REFERENCES `metric` (`guid`),
      CONSTRAINT `fore_metric_comparison_origin_metric_id` FOREIGN KEY (`origin_metric_id`) REFERENCES `metric` (`guid`)
)ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '指标同环比';

alter table metric add column endpoint_group varchar(64)  default NULL COMMENT '对象组';

CREATE TABLE `log_keyword_notify_rel` (
          `guid` varchar(64) NOT NULL COMMENT '唯一标识',
          `log_keyword_monitor` varchar(64) DEFAULT NULL COMMENT '业务关键字监控',
          `log_keyword_config` varchar(64) DEFAULT NULL COMMENT '业务关键字配置',
          `notify` varchar(64) NOT NULL COMMENT '通知表',
          PRIMARY KEY (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `db_keyword_monitor` (
      `guid` varchar(64) NOT NULL COMMENT '唯一标识',
      `service_group` varchar(64) NOT NULL COMMENT '业务监控组',
      `name` varchar(64) NOT NULL COMMENT '名称',
      `query_sql` text DEFAULT NULL COMMENT '查询sql',
      `priority` varchar(16) DEFAULT NULL COMMENT '告警级别',
      `content` text DEFAULT NULL COMMENT '告警内容',
      `notify_enable` tinyint(4) DEFAULT NULL COMMENT '是否通知',
      `active_window` varchar(255) DEFAULT NULL COMMENT '生效时间段',
      `step` int(11) DEFAULT NULL COMMENT '采集间隔',
      `monitor_type` varchar(32) DEFAULT NULL COMMENT '监控类型',
      `create_user` varchar(64) DEFAULT NULL COMMENT '创建人',
      `update_user` varchar(64) DEFAULT NULL COMMENT '更新人',
      `create_time` datetime DEFAULT NULL COMMENT '创建时间',
      `update_time` datetime DEFAULT NULL COMMENT '更新时间',
      PRIMARY KEY (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `db_keyword_endpoint_rel` (
           `guid` varchar(64) NOT NULL COMMENT '唯一标识',
           `db_keyword_monitor` varchar(64) NOT NULL COMMENT '数据库关键字监控',
           `source_endpoint` varchar(160) DEFAULT NULL COMMENT '源对象',
           `target_endpoint` varchar(160) DEFAULT NULL COMMENT '目标对象',
           PRIMARY KEY (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `db_keyword_notify_rel` (
         `guid` varchar(64) NOT NULL COMMENT '唯一标识',
         `db_keyword_monitor` varchar(64) NOT NULL COMMENT '数据库关键字监控',
         `notify` varchar(64) NOT NULL COMMENT '通知表',
         PRIMARY KEY (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

alter table log_keyword_monitor add column create_time varchar(32) default null;
alter table log_keyword_monitor add column create_user varchar(64) default null;
alter table log_keyword_monitor add column update_user varchar(64) default null;
alter table log_keyword_config add column active_window varchar(255) default null;
alter table log_keyword_config add column create_time varchar(32) default null;
alter table endpoint_new add column create_user varchar(64) default null COMMENT '创建人';
alter table endpoint_new add column update_user varchar(64) default null COMMENT '更新人';

alter table endpoint_group add column create_user varchar(64) default null COMMENT '创建人';
alter table endpoint_group add column update_user varchar(64) default null COMMENT '更新人';

alter table role ADD UNIQUE (name);

alter table log_metric_string_map modify column log_metric_config varchar(64) default null;

SET FOREIGN_KEY_CHECKS=0;
update chart set metric=replace(metric,'.','_') where metric like '%.%';
update alarm_strategy set metric=replace(metric,'.','_') where metric in (select guid from metric where metric like '%.%' and service_group is null);
update alarm_strategy_metric set metric=replace(metric,'.','_') where metric in (select guid from metric where metric like '%.%' and service_group is null);
update custom_chart_series set metric=replace(metric,'.','_'),metric_guid=replace(metric_guid,'.','_');
update metric set metric=replace(metric,'.','_'),guid=replace(guid,'.','_') where metric like '%.%' and service_group is null;
update prom_metric set metric=replace(metric,'.','_') where metric like '%.%';
SET FOREIGN_KEY_CHECKS=1;
#@v3.0.4.1-end@;

#@v3.1.4-begin@;
alter table monitor_type add column system_type tinyint(1) default 0 COMMENT '系统类型,0为非系统类型,1为系统类型';
alter table monitor_type add column create_user varchar(64) default null COMMENT '创建人';
alter table monitor_type add column create_time varchar(32) default null COMMENT '创建时间';
alter table custom_dashboard_chart_rel add column group_display_config text default null COMMENT '组里视图位置长与宽';
alter table log_metric_string_map modify column target_value varchar(128) default null COMMENT '目标值';
alter table db_metric_monitor add column update_user varchar(64) default null COMMENT '更新人';

alter table service_group add column update_user varchar(64) default null COMMENT '更新人';
alter table log_keyword_config add column update_user varchar(64) default null COMMENT '更新人';

alter table alarm add index alarm_name (alarm_name);

update monitor_type set system_type = 1;
CREATE TABLE `db_keyword_alarm` (
    `id` int(11) unsigned NOT NULL auto_increment COMMENT '自增id',
    `alarm_id` int(11) NOT null COMMENT '告警id',
    `endpoint` varchar(255) NOT null COMMENT '告警对象',
    `status` varchar(20) NOT null COMMENT '状态',
    `db_keyword_monitor` varchar(64) NOT null COMMENT '数据库关键字配置',
    `content` text COMMENT '告警内容',
    `tags` varchar(1024) DEFAULT '' COMMENT '告警标签',
    `start_value` double DEFAULT null COMMENT '开始值',
    `end_value` double DEFAULT null COMMENT '结束值',
    `updated_time` datetime DEFAULT null COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_log_keyword_alarm_id` (`alarm_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

alter table log_keyword_alarm add column log_keyword_config varchar(64) default null;
alter table custom_chart_series_config modify column series_name varchar(512) default null;
alter table metric add column db_metric_monitor varchar(64) default null;
alter table alarm_strategy_metric add column monitor_engine tinyint(4) default 0;
create index idx_alarm_strategy_monitor_engine on alarm_strategy_metric(monitor_engine);
alter table alarm_strategy_metric add column monitor_engine_expr text default null;

CREATE TABLE `alarm_firing` (
    `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
    `endpoint` varchar(255) NOT null COMMENT '告警对象',
    `metric` varchar(255) NOT NULL COMMENT '告警指标',
    `tags` varchar(1024) DEFAULT '' COMMENT '告警标签',
    `alarm_name` varchar(64) DEFAULT NULL COMMENT '告警名称',
    `alarm_strategy` varchar(64) DEFAULT NULL COMMENT '告警策略配置',
    `notify_id` varchar(64) DEFAULT NULL COMMENT '告警通知配置',
    `expr` varchar(4096) DEFAULT NULL COMMENT '告警表达式',
    `cond` varchar(50) NOT NULL COMMENT '告警条件',
    `last` varchar(50) NOT NULL COMMENT '告警持续时间',
    `priority` varchar(50) NOT NULL COMMENT '告警级别',
    `content` text COMMENT '告警描述内容',
    `start_value` double DEFAULT null COMMENT '告警发生值',
    `start` datetime DEFAULT null COMMENT '告警发生时间',
    `custom_message` varchar(500) DEFAULT NULL COMMENT '告警人工备注',
    `unique_hash` varchar(64) DEFAULT NULL COMMENT '告警唯一标识(对象+指标+标签+配置)',
    `alarm_id` int(11) DEFAULT NULL COMMENT '告警历史表id',
    PRIMARY KEY (`id`),
    UNIQUE KEY `unique_alarm_firing_hash` (`unique_hash`),
    KEY `alarm_firing_endpoint_idx` (`endpoint`),
    KEY `alarm_firing_metric_idx` (`metric`),
    KEY `alarm_firing_priority_idx` (`priority`),
    KEY `alarm_firing_name` (`alarm_name`),
    KEY `alarm_firing_alarm_id` (`alarm_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

alter table log_monitor_template add column success_code varchar(200) default null COMMENT '成功码';
alter table log_metric_template add column color_group varchar(32) default null COMMENT '默认色系';
alter table log_metric_template add column auto_alarm tinyint(1) default 0 COMMENT '自动告警,1表示自动告警';
alter table log_metric_template add column range_config tinytext default null COMMENT '阈值配置json';
alter table log_metric_group add column template_snapshot text default null COMMENT '模版快照';
alter table log_metric_group add column ref_template_version varchar(32) default null COMMENT '引用模版版本';
alter table alarm_strategy add column log_metric_group varchar(64) default null COMMENT '业务日志指标id';
alter table custom_dashboard add column log_metric_group varchar(64) default null COMMENT '业务日志指标id';
alter table custom_chart add column log_metric_group varchar(64) default null COMMENT '业务日志指标id';
alter table alarm_strategy_tag add column equal varchar(32) default 'in' COMMENT 'in/not in';
alter table custom_chart_series_tag add column equal varchar(32) default 'in' COMMENT 'in/not in';
alter table alarm_strategy add column create_user varchar(64) default null COMMENT '创建人';

-- 新增明细失败量
insert into log_metric_template(guid,log_monitor_template,log_param_name,metric,display_name,step,agg_type,tag_config,create_user,update_user,create_time,update_time) select concat(guid,'_1'),log_monitor_template,log_param_name,'req_fail_count_detail' as metric,'分类失败量' as display_name,step,agg_type,'["code","retcode"]' as tag_config,create_user,update_user,create_time,update_time from log_metric_template where metric='req_fail_count';
insert into metric(guid,metric,monitor_type,prom_expr,update_time,service_group,workspace,log_metric_config,log_metric_template,log_metric_group,create_time,create_user,update_user,endpoint_group,db_metric_monitor) select replace(guid,'req_fail_count','req_fail_count_detail') as guid,replace(metric,'req_fail_count','req_fail_count_detail') as metric,monitor_type,prom_expr,update_time,service_group,workspace,log_metric_config,concat(log_metric_template,'_1'),log_metric_group,create_time,create_user,update_user,endpoint_group,db_metric_monitor from metric where log_metric_template in (select guid from log_metric_template where metric='req_fail_count');
-- 把失败量表达式改成汇总失败量
update log_metric_template set tag_config='["code"]',agg_type='{req_count}-{req_suc_count}' where metric='req_fail_count';
update metric t1,metric t2,metric t3 set t1.prom_expr=concat(replace(t2.prom_expr,'(key,agg,service_group,code)','(service_group,code)'),'-',replace(t3.prom_expr,'(key,agg,service_group,code,retcode)','(service_group,code)')) where t1.log_metric_group=t2.log_metric_group and t1.log_metric_group=t3.log_metric_group and t1.log_metric_template in (select guid from log_metric_template where metric='req_fail_count') and t2.metric like '%req_count' and t3.metric like '%req_suc_count';
-- 去掉成功量返回码
update log_metric_template set tag_config='["code"]' where metric='req_suc_count' and tag_config='["code","retcode"]';
-- 把新视图和阈值配置中req_fail_count改成req_fail_count_detail
update custom_chart_series set metric=replace(metric,'req_fail_count','req_fail_count_detail'),metric_guid=replace(metric_guid,'req_fail_count','req_fail_count_detail') where metric like '%req_fail_count';
update alarm_strategy_metric set metric=replace(metric,'req_fail_count','req_fail_count_detail') where metric in (select guid from metric where log_metric_template in (select guid from log_metric_template where metric='req_fail_count'));
update metric set prom_expr=replace(prom_expr,',code="$t_code"',',retcode="$t_retcode",code="$t_code"') where log_metric_template in (select guid from log_metric_template where metric='req_fail_count_detail');

INSERT INTO sys_parameter (guid,param_key,param_value) VALUES ('metric_template_13','service_metric_template','{"name":"(a+b)/2","prom_expr":"(@a+@b)/2","param":"@a,@b"}');
INSERT INTO sys_parameter (guid,param_key,param_value) VALUES ('metric_template_14','metric_template','{"name":"(a+b)/2","prom_expr":"(@a+@b)/2","param":"@a,@b"}');
INSERT INTO sys_parameter (guid,param_key,param_value) VALUES ('metric_template_15','service_metric_template','{"name":"a+b","prom_expr":"@a+@b","param":"@a,@b"}');
INSERT INTO sys_parameter (guid,param_key,param_value) VALUES ('metric_template_16','metric_template','{"name":"a+b","prom_expr":"@a+@b)","param":"@a,@b"}');
INSERT INTO sys_parameter (guid,param_key,param_value) VALUES ('metric_template_17','service_metric_template','{"name":"avg(a+b)","prom_expr":"avg(@a)+avg(@b)","param":"@a,@b"}');
INSERT INTO sys_parameter (guid,param_key,param_value) VALUES ('metric_template_18','metric_template','{"name":"avg(a+b)","prom_expr":"avg(@a)+avg(@b)","param":"@a,@b"}');
#@v3.1.4-end@;

#@v3.2.3-begin@;
alter table alarm_strategy modify column active_window varchar(255) default null;
#@v3.2.3-end@;

#@v3.2.3.8-begin@;
create index `alarm_strategy_status_idx` on alarm (`status`,`alarm_strategy`);
#@v3.2.3.8-end@;

#@v3.2.7-begin@;
alter table log_metric_string_map add column log_monitor_template varchar(64) default null COMMENT '业务日志模版';
alter table log_metric_string_map add index log_metric_string_map_template (log_monitor_template);
#@v3.2.7-end@;

#@v3.2.8.18-begin@;
alter table alarm_strategy_metric add column log_type varchar(32) default null COMMENT '日志类型,custom自定义';
alter table log_metric_config add column auto_alarm tinyint(1) default 0 COMMENT '自动告警,1表示自动告警';
alter table log_metric_config add column range_config tinytext default null COMMENT '阈值配置json';
alter table log_metric_config add column color_group varchar(32) default null COMMENT '默认色系';
alter table alarm add index start_time(start);
alter table alarm_custom add index alarm_custom_update_at(update_at);
alter table endpoint_new add index endpoint_new_update_time(update_time);
alter table metric add index metric_update_time(update_time);
#@v3.2.8.18-end@;
#@v3.2.8.23-begin@;
alter table log_metric_group add column auto_alarm tinyint(1) default 0 COMMENT '自动告警,1表示自动告警';
alter table log_metric_group add column auto_dashboard tinyint(1) default 0 COMMENT '自动生成自定义看板,1表示自动生成';
#@v3.2.8.23-end@;
#@v3.3.2-begin@;
alter table service_group add index service_group_update_time(update_time);
alter table custom_chart_permission add index custom_chart_permission_role(role_id);
alter table custom_chart add index dashboard_chart_public(public);
#@v3.3.2-end@;
#@v3.4.3-begin@;
alter table alarm_custom add column create_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间';
alter table alarm_custom add column alarm_total int(11) DEFAULT 1 COMMENT '告警次数';
CREATE TABLE `history_alarm_custom`
(
    `id`             int(11) unsigned NOT NULL AUTO_INCREMENT,
    `alert_info`     varchar(2048)      DEFAULT '',
    `alert_ip`       varchar(50)        DEFAULT '',
    `alert_level`    int(11) NOT NULL,
    `alert_obj`      varchar(50)        DEFAULT '',
    `alert_title`    varchar(1024)      DEFAULT '',
    `alert_reciver`  varchar(500)       DEFAULT '',
    `remark_info`    varchar(256)       DEFAULT '',
    `sub_system_id`  varchar(10)        DEFAULT '',
    `closed`         int(10) unsigned NOT NULL DEFAULT '1',
    `create_at`      timestamp NOT NULL,
    `update_at`      timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `use_umg_policy` varchar(50)        DEFAULT '',
    `alert_way`      varchar(50)        DEFAULT '',
    `custom_message` varchar(500)       DEFAULT NULL,
    `alarm_total` int(11) DEFAULT '1' COMMENT '告警次数',
    PRIMARY KEY (`id`),
    index `update_at_index` (`update_at`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;
#@v3.4.3-end@;
#@v3.5.7-begin@;
ALTER TABLE custom_dashboard_chart_rel MODIFY COLUMN `group` VARCHAR(500) DEFAULT NULL;
#@v3.5.7-end@;
#@v3.5.8-begin@;
alter table log_monitor_template add column auto_alarm tinyint(1) default 0 COMMENT '自动告警,1表示自动告警';
alter table log_monitor_template add column auto_dashboard tinyint(1) default 0 COMMENT '自动生成自定义看板,1表示自动生成';
alter table log_metric_group add column status varchar(20) default 'enable' COMMENT '是否启用,默认启用';
alter table alarm modify column alarm_name varchar(150) default null comment '告警名称';
#@v3.5.8-end@;
#@v3.6.5-begin@;
alter table alarm_condition modify column expr varchar(2000) default null COMMENT '指标表达式';
ALTER TABLE alarm_condition_rel ADD INDEX idx_alarm (alarm);
ALTER TABLE alarm_condition_rel ADD INDEX idx_alarm_condition (alarm_condition);
#@v3.6.5-end@;
#@v3.6.8-begin@;
alter table custom_chart_series modify column metric_guid varchar(128) default NULL COMMENT '指标Id';
update metric set prom_expr='node_filesystem_files_free{instance="$address",mountpoint ="/"} / node_filesystem_files{instance="$address",mountpoint ="/"} * 100' where guid='file_handler_free_percent__host';
alter table log_keyword_config modify column name varchar(150) default NULL COMMENT '告警名称';
alter table db_keyword_monitor modify column name varchar(150) default NULL COMMENT '告警名称';
#@v3.6.8-end@;