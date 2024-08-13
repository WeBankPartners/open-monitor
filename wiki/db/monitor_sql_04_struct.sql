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
