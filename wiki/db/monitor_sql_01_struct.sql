create database if not exists `monitor`;
USE `monitor`;

DROP TABLE IF EXISTS `button`;

CREATE TABLE `button` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `group_id` int(11) NOT NULL DEFAULT '0',
  `name` varchar(50) NOT NULL DEFAULT '',
  `b_type` varchar(50) NOT NULL DEFAULT '',
  `b_text` varchar(50) NOT NULL DEFAULT '',
  `refresh_panels` tinyint(1) NOT NULL DEFAULT '0',
  `refresh_charts` tinyint(1) NOT NULL DEFAULT '0',
  `option_group` int(11) DEFAULT '0',
  `refresh_button` int(10) DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `dashboard_button_index` (`group_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE = utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `chart`;

CREATE TABLE `chart` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `group_id` int(11) NOT NULL DEFAULT '0',
  `endpoint` varchar(2000) DEFAULT '',
  `metric` varchar(2000) DEFAULT '',
  `col` int(11) DEFAULT '6',
  `url` varchar(200) DEFAULT '',
  `unit` varchar(50) DEFAULT '',
  `title` varchar(100) DEFAULT '',
  `grid_type` varchar(20) DEFAULT 'line',
  `series_name` varchar(50) DEFAULT 'metric',
  `rate` tinyint(1) DEFAULT '0',
  `agg_type` varchar(20) DEFAULT 'avg',
  `legend` varchar(50) DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `dashboard_chart_index` (`group_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE = utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `dashboard`;

CREATE TABLE `dashboard` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `dashboard_type` varchar(50) NOT NULL,
  `search_enable` tinyint(1) NOT NULL DEFAULT '0',
  `search_id` int(11) DEFAULT '0',
  `button_enable` tinyint(1) NOT NULL DEFAULT '0',
  `button_group` int(11) DEFAULT '0',
  `message_enable` tinyint(1) NOT NULL DEFAULT '0',
  `message_group` int(11) DEFAULT '0',
  `message_url` varchar(200) DEFAULT '',
  `panels_enable` tinyint(1) NOT NULL DEFAULT '1',
  `panels_type` varchar(20) DEFAULT 'tabs',
  `panels_group` int(11) DEFAULT '0',
  `panels_param` varchar(50) DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `dashboard_type_index` (`dashboard_type`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE = utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `endpoint`;

CREATE TABLE `endpoint` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `guid` varchar(255) NOT NULL,
  `name` varchar(255) NOT NULL DEFAULT '',
  `ip` varchar(50) NOT NULL DEFAULT '',
  `endpoint_version` varchar(255) NOT NULL DEFAULT '',
  `export_type` varchar(50) NOT NULL DEFAULT 'host',
  `export_version` varchar(50) NOT NULL DEFAULT '',
  `step` int(11) DEFAULT '10',
  `address` varchar(50) NOT NULL DEFAULT '',
  `os_type` varchar(50) NOT NULL DEFAULT 'linux',
  `create_at` varchar(50) NOT NULL DEFAULT '',
  `stop_alarm` tinyint(4) DEFAULT '0',
  `address_agent` varchar(50) NOT NULL DEFAULT '',
  `update_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `monitor_endpoint_guid` (`guid`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

DROP TABLE IF EXISTS `endpoint_metric`;

CREATE TABLE `endpoint_metric` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `endpoint_id` int(11) NOT NULL,
  `metric` varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `monitor_endpoint_metric` (`endpoint_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE = utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `option`;

CREATE TABLE `option` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `group_id` int(11) NOT NULL DEFAULT '0',
  `option_text` varchar(100) NOT NULL DEFAULT '',
  `option_value` varchar(100) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `dashboard_option_index` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE = utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `panel`;

CREATE TABLE `panel` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `group_id` int(11) NOT NULL DEFAULT '0',
  `title` varchar(50) NOT NULL DEFAULT '',
  `tags_enable` tinyint(1) NOT NULL DEFAULT '0',
  `tags_url` varchar(200) NOT NULL DEFAULT '',
  `tags_key` varchar(50) DEFAULT '',
  `chart_group` int(11) DEFAULT '0',
  `auto_display` int(11) DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `dashboard_panel_index` (`group_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE = utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `prom_metric`;

CREATE TABLE `prom_metric` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `metric` varchar(50) NOT NULL DEFAULT '',
  `metric_type` varchar(50) NOT NULL DEFAULT '',
  `prom_ql` varchar(500) NOT NULL DEFAULT '',
  `prom_main` varchar(255) DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE = utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `search`;

CREATE TABLE `search` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(50) NOT NULL DEFAULT '',
  `search_url` varchar(200) NOT NULL DEFAULT '',
  `search_col` varchar(50) NOT NULL DEFAULT '',
  `refresh_panels` tinyint(1) NOT NULL DEFAULT '0',
  `refresh_message` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE = utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `alarm`;

CREATE TABLE `alarm` (
  `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
  `strategy_id` INT(11) NOT NULL,
  `endpoint` VARCHAR(255) NOT NULL,
  `status` VARCHAR(20) NOT NULL,
  `s_metric` VARCHAR(255) NOT NULL,
  `s_expr` VARCHAR(500) NOT NULL,
  `s_cond` VARCHAR(50) NOT NULL,
  `s_last` VARCHAR(50) NOT NULL,
  `s_priority` VARCHAR(50) NOT NULL,
  `content` VARCHAR(500) DEFAULT '',
  `tags` varchar(255) DEFAULT '',
  `start_value` DOUBLE,
  `start` DATETIME,
  `end_value` DOUBLE,
  `end` DATETIME,
  `close_type` VARCHAR(50) DEFAULT 'auto',
  `close_msg` VARCHAR(255),
  `close_user` VARCHAR(50),
  PRIMARY KEY (`id`),
  UNIQUE KEY `alarm_unique_index` (`strategy_id`,`endpoint`,`status`,`tags`,`start`)
) ENGINE=INNODB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

DROP TABLE IF EXISTS `grp`;

CREATE TABLE `grp` (
  `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(255) NOT NULL,
  `description` VARCHAR(255) DEFAULT '',
  `create_user` VARCHAR(50) DEFAULT '',
  `update_user` VARCHAR(50) DEFAULT '',
  `parent` int(11) DEFAULT '0',
  `create_at` DATETIME,
  `update_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=INNODB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE = utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `tpl`;

CREATE TABLE `tpl` (
  `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
  `grp_id` INT(11) DEFAULT 0,
  `endpoint_id` INT(11) DEFAULT 0,
  `notify_url` VARCHAR(255) DEFAULT '',
  `create_user` VARCHAR(50) DEFAULT '',
  `update_user` VARCHAR(50) DEFAULT '',
  `action_user` VARCHAR(255) DEFAULT '',
  `action_role` VARCHAR(255) DEFAULT '',
  `extra_mail` text,
  `extra_phone` text,
  `create_at` DATETIME,
  `update_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=INNODB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE = utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `grp_endpoint`;

CREATE TABLE `grp_endpoint` (
  `grp_id` INT(11) UNSIGNED NOT NULL,
  `endpoint_id` INT(11) UNSIGNED NOT NULL,
  UNIQUE KEY `idx_grp_endpoint` (`grp_id`,`endpoint_id`)
) ENGINE=INNODB DEFAULT CHARSET=utf8mb4 COLLATE = utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `strategy`;

CREATE TABLE `strategy` (
  `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
  `tpl_id` INT(11) NOT NULL,
  `metric` VARCHAR(255) NOT NULL,
  `expr` VARCHAR(500) NOT NULL,
  `cond` VARCHAR(50) NOT NULL,
  `last` VARCHAR(50) NOT NULL,
  `priority` VARCHAR(50) NOT NULL,
  `content` VARCHAR(500) NOT NULL,
  `config_type` varchar(50) default 'default',
  PRIMARY KEY (`id`),
  KEY `idx_strategy_tpl_id` (`tpl_id`)
) ENGINE=INNODB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE = utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `log_monitor`;

CREATE TABLE `log_monitor` (
  `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
  `strategy_id` INT(11) NOT NULL,
  `path` VARCHAR(255) NOT NULL,
  `keyword` VARCHAR(255) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=INNODB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE = utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `maintain`;

CREATE TABLE `maintain` (
  `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
  `endpoint_id` INT(11) NOT NULL,
  `maintain_start` DATETIME,
  `maintain_end` DATETIME,
  `maintain_user` VARCHAR(50) NOT NULL,
  `update_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=INNODB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE = utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `custom_dashboard`;

CREATE TABLE `custom_dashboard` (
  `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(50) DEFAULT NULL,
  `panels_group` INT(11) DEFAULT NULL,
  `cfg` MEDIUMTEXT,
  `main` tinyint(4) default 0,
  `create_user` VARCHAR(50) DEFAULT '',
  `update_user` VARCHAR(50) DEFAULT '',
  `create_at` DATETIME DEFAULT NULL,
  `update_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=INNODB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE `process_monitor` (
  `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
  `endpoint_id` INT(11) NOT NULL,
  `name` VARCHAR(100) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=INNODB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE `business_monitor` (
  `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
  `endpoint_id` INT(11) NOT NULL,
  `path` VARCHAR(100) NOT NULL,
  `owner_endpoint` varchar(255) DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=INNODB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE `alarm_custom` (
  `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
  `alert_info` VARCHAR(2048) DEFAULT '',
  `alert_ip` VARCHAR(50) DEFAULT '',
  `alert_level` INT NOT NULL,
  `alert_obj` VARCHAR(50) DEFAULT '',
  `alert_title` VARCHAR(100) DEFAULT '',
  `alert_reciver` VARCHAR(500) DEFAULT '',
  `remark_info` VARCHAR(256) DEFAULT '',
  `sub_system_id` VARCHAR(10) DEFAULT '',
  `closed` INT(10) UNSIGNED NOT NULL DEFAULT '0',
  `update_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `closed_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=INNODB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE `user` (
  `id` INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(64) NOT NULL,
  `passwd` VARCHAR(64) NOT NULL DEFAULT '',
  `display_name` VARCHAR(128) NOT NULL DEFAULT '',
  `email` VARCHAR(255) NOT NULL DEFAULT '',
  `phone` VARCHAR(16) NOT NULL DEFAULT '',
  `ext_contact_one` VARCHAR(255) NOT NULL DEFAULT '',
  `ext_contact_two` VARCHAR(255) NOT NULL DEFAULT '',
  `created` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_name` (`name`)
) ENGINE=INNODB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE `role` (
  `id` INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(64) NOT NULL,
  `display_name` VARCHAR(255) NOT NULL DEFAULT '',
  `email` VARCHAR(255) NOT NULL DEFAULT '',
  `parent` INT(10) UNSIGNED NOT NULL DEFAULT '0',
  `owner_user` VARCHAR(64) NOT NULL DEFAULT '',
  `creator` VARCHAR(255) DEFAULT '',
  `created` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=INNODB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE `rel_role_user` (
  `id` INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `role_id` INT(10) UNSIGNED NOT NULL,
  `user_id` INT(10) UNSIGNED NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=INNODB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE `panel_recursive` (
  `guid` VARCHAR(50) NOT NULL,
  `display_name` VARCHAR(50) NOT NULL DEFAULT '',
  `parent` TEXT,
  `endpoint` TEXT,
  `email` TEXT,
  `phone` TEXT,
  `role` TEXT,
  `firing_callback_name` VARCHAR(100) DEFAULT '',
  `firing_callback_key` VARCHAR(50) DEFAULT '',
  `recover_callback_name` VARCHAR(100) DEFAULT '',
  `recover_callback_key` VARCHAR(50) DEFAULT '',
  `obj_type` varchar(50) DEFAULT 'object',
  `update_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`guid`)
) ENGINE=INNODB DEFAULT CHARSET=utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE `rel_role_grp` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `role_id` int(10) unsigned NOT NULL,
  `grp_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE `endpoint_telnet` (
  `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
  `endpoint_guid` VARCHAR(255) NOT NULL,
  `port` VARCHAR(255) DEFAULT '',
  `note` VARCHAR(255) DEFAULT '',
  `update_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=INNODB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE `endpoint_http` (
  `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
  `endpoint_guid` VARCHAR(255) NOT NULL,
  `method` VARCHAR(255) DEFAULT '',
  `url` VARCHAR(255) DEFAULT '',
  `update_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=INNODB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE `agent_manager` (
  `endpoint_guid` VARCHAR(255) NOT NULL,
  `name` VARCHAR(255) DEFAULT '',
  `user` VARCHAR(255) DEFAULT '',
  `password` VARCHAR(255) DEFAULT '',
  `instance_address` VARCHAR(255) DEFAULT '',
  `agent_address` VARCHAR(255) DEFAULT '',
  `config_file` VARCHAR(255) DEFAULT '',
  `bin_path` VARCHAR(255) DEFAULT '',
  PRIMARY KEY (`endpoint_guid`)
) ENGINE=INNODB DEFAULT CHARSET=utf8;