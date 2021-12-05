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

insert into monitor_type(guid,display_name) value ('host','host'),('mysql','mysql'),('redis','redis'),('java','java'),('nginx','nginx'),('ping','ping'),('telnet','telnet'),('http','http'),('windows','windows'),('snmp','snmp'),('process','process'),('pod','pod');

CREATE TABLE `endpoint_new` (
  `guid` varchar(128) NOT NULL PRIMARY KEY,
  `name` varchar(64) NOT NULL,
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
  `endpoint` varchar(128) NOT NULL,
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
  `source_endpoint` varchar(128),
  `target_endpoint` varchar(128),
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
  `source_endpoint` varchar(128),
  `target_endpoint` varchar(128),
  CONSTRAINT `db_monitor_endpoint_metric` FOREIGN KEY (`db_metric_monitor`) REFERENCES `db_metric_monitor` (`guid`),
  CONSTRAINT `db_monitor_endpoint_source` FOREIGN KEY (`source_endpoint`) REFERENCES `endpoint_new` (`guid`),
  CONSTRAINT `db_monitor_endpoint_target` FOREIGN KEY (`target_endpoint`) REFERENCES `endpoint_new` (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

insert into grp(name,endpoint_type) value ('default_process_group','process');
insert  into `dashboard`(`dashboard_type`,`search_enable`,`search_id`,`button_enable`,`button_group`,`message_enable`,`message_group`,`message_url`,`panels_enable`,`panels_type`,`panels_group`,`panels_param`) values ('process',1,1,1,1,0,0,'',1,'tabs',11,'endpoint={endpoint}');
insert  into `panel`(`group_id`,`title`,`tags_enable`,`tags_url`,`tags_key`,`chart_group`,`auto_display`) values (11,'process',0,'','',18,0);
insert  into `chart`(`group_id`,`endpoint`,`metric`,`col`,`url`,`unit`,`title`,`grid_type`,`series_name`,`rate`,`agg_type`,`legend`) values (18,'','process_cpu_used_percent',6,'/dashboard/chart','','process.cpu.used','line','metric',0,'avg','$metric'),(18,'','process_mem_byte',6,'/dashboard/chart','','process.mem.used','line','metric',0,'avg','$metric');
insert  into `prom_metric`(`metric`,`metric_type`,`prom_ql`,`prom_main`) values ('process_alive_count','process','node_process_monitor_count_current{instance=\"$address\"}',''),('process_cpu_used_percent','process','node_process_monitor_cpu{instance=\"$address\"}',''),('process_mem_byte','process','node_process_monitor_mem{instance=\"$address\"}','');
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
  `endpoint` varchar(128) NOT NULL,
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
insert into alarm_strategy select CONCAT('old_',t1.id),t3.name,t4.guid ,t1.cond,t1.`last`,t1.priority,t1.content,t1.notify_enable,t1.notify_delay,'' as update_time from strategy t1 left join tpl t2 on t1.tpl_id=t2.id left join grp t3 on t2.grp_id=t3.id left join metric t4 on (t1.metric=t4.metric and t3.endpoint_type=t4.monitor_type) where t2.grp_id>0;
#@v1.13.0.2-end@;