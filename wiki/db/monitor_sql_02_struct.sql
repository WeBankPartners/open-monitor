DROP TABLE IF EXISTS `main_dashboard`;
CREATE TABLE IF NOT EXISTS `main_dashboard` (
`guid` varchar(64) NOT NULL,
`role_id` varchar(64) NOT NULL,
`custom_dashboard` int(11)  unsigned not null COMMENT '首页看板表',
PRIMARY KEY (`guid`),
CONSTRAINT `fore_main_dashboard_custom_dashboard` FOREIGN KEY (`custom_dashboard`) REFERENCES `custom_dashboard` (`id`)
)ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '首页看板表';

DROP TABLE IF EXISTS `custom_chart`;
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


DROP TABLE IF EXISTS `custom_dashboard_chart_rel`;
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


DROP TABLE IF EXISTS `custom_chart_series`;
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


DROP TABLE IF EXISTS `custom_chart_permission`;
CREATE TABLE IF NOT EXISTS `custom_chart_permission` (
`guid` varchar(64) NOT NULL,
`dashboard_chart` varchar(64) NOT NULL COMMENT '所属看板图表',
`role_id` varchar(64) NOT NULL COMMENT '角色名',
`permission` varchar(32)  NULL COMMENT '权限,mgmt/use',
PRIMARY KEY (`guid`),
CONSTRAINT `fore_custom_chart_permission_dashboard_chart` FOREIGN KEY (`dashboard_chart`) REFERENCES `custom_chart` (`guid`)
)ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '公共图表权限表';


DROP TABLE IF EXISTS `custom_chart_series_config`;
CREATE TABLE IF NOT EXISTS `custom_chart_series_config` (
`guid` varchar(64) NOT NULL,
`dashboard_chart_config` varchar(64) NOT NULL COMMENT '图表配置表',
`tags` text  NULL COMMENT '标签',
`color` varchar(32)  NULL COMMENT '颜色',
`series_name` varchar(255)  NULL COMMENT '指标+对象+标签值',
PRIMARY KEY (`guid`),
CONSTRAINT `fore_custom_chart_series_config_dashboard_chart_config` FOREIGN KEY (`dashboard_chart_config`) REFERENCES `custom_chart_series` (`guid`)
)ENGINE= InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '自定义图表配置表';

DROP TABLE IF EXISTS `custom_chart_series_tag`;
CREATE TABLE IF NOT EXISTS `custom_chart_series_tag` (
`guid` varchar(64) NOT NULL,
`dashboard_chart_config` varchar(64) NOT NULL COMMENT '图表配置表',
`name` varchar(64)  NOT NULL COMMENT '标签名',
PRIMARY KEY (`guid`),
CONSTRAINT `fore_custom_chart_series_tag_dashboard_chart_config` FOREIGN KEY (`dashboard_chart_config`) REFERENCES `custom_chart_series` (`guid`)
)ENGINE = InnoDB DEFAULT CHARSET =utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '自定义图表标签配置表';


DROP TABLE IF EXISTS `custom_chart_series_tagvalue`;
CREATE TABLE IF NOT EXISTS `custom_chart_series_tagvalue` (
`id` int(11) NOT NULL AUTO_INCREMENT,
`dashboard_chart_tag` varchar(64) NOT NULL COMMENT '所属图表标签',
`value` varchar(255)  NULL COMMENT '标签值',
PRIMARY KEY (`id`),
CONSTRAINT `fore_custom_chart_series_tagvalue_dashboard_chart_tag` FOREIGN KEY (`dashboard_chart_tag`) REFERENCES `custom_chart_series_tag` (`guid`)
)ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '自定义图表标签值配置表';


alter table custom_dashboard_role_rel MODIFY role_id varchar(64)  NOT NULL COMMENT '角色id';
alter table metric add column log_metric_group varchar(64)  default NULL COMMENT '业务日志指标组配置Id';
alter table metric ADD CONSTRAINT  `fore_metric_log_metric_group` FOREIGN KEY (log_metric_group) REFERENCES log_metric_group(guid);
alter table custom_chart add column pie_type varchar(32)  default NULL COMMENT '饼图类型';
alter table custom_chart_series add column pie_display_tag varchar(64)  default NULL COMMENT '饼图展示标签';