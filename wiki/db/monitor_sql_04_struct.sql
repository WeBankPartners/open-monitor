alter table monitor_type add column system_type tinyint(1) default 0 COMMENT '系统类型,0为非系统类型,1为系统类型';
alter table monitor_type add column create_user varchar(64) default null COMMENT '创建人';
alter table monitor_type add column create_time varchar(32) default null COMMENT '创建时间';

update monitor_type set system_type = 1;
