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
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE = utf8mb4_unicode_ci;

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
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE = utf8mb4_unicode_ci;


alter table log_metric_config add column auto_alarm tinyint(1) default 0 COMMENT '自动告警,1表示自动告警';
alter table log_metric_config add column range_config tinytext default null COMMENT '阈值配置json';
alter table log_metric_config add column color_group varchar(32) default null COMMENT '默认色系';
alter table alarm add index start_time(start);
alter table alarm_custom add index alarm_custom_update_at(update_at);
alter table endpoint_new add index endpoint_new_update_time(update_time);
alter table metric add index metric_update_time(update_time);



