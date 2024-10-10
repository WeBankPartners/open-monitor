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


