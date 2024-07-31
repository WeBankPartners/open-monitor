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