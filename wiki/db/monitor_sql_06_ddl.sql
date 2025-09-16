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
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE = utf8mb4_unicode_ci;