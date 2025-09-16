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
)ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '指标同环比';

alter table metric add column endpoint_group varchar(64)  default NULL COMMENT '对象组';

CREATE TABLE `log_keyword_notify_rel` (
          `guid` varchar(64) NOT NULL COMMENT '唯一标识',
          `log_keyword_monitor` varchar(64) DEFAULT NULL COMMENT '业务关键字监控',
          `log_keyword_config` varchar(64) DEFAULT NULL COMMENT '业务关键字配置',
          `notify` varchar(64) NOT NULL COMMENT '通知表',
          PRIMARY KEY (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4mb4;

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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4mb4;

CREATE TABLE `db_keyword_endpoint_rel` (
           `guid` varchar(64) NOT NULL COMMENT '唯一标识',
           `db_keyword_monitor` varchar(64) NOT NULL COMMENT '数据库关键字监控',
           `source_endpoint` varchar(160) DEFAULT NULL COMMENT '源对象',
           `target_endpoint` varchar(160) DEFAULT NULL COMMENT '目标对象',
           PRIMARY KEY (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4mb4;

CREATE TABLE `db_keyword_notify_rel` (
         `guid` varchar(64) NOT NULL COMMENT '唯一标识',
         `db_keyword_monitor` varchar(64) NOT NULL COMMENT '数据库关键字监控',
         `notify` varchar(64) NOT NULL COMMENT '通知表',
         PRIMARY KEY (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4mb4;

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