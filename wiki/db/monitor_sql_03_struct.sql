CREATE TABLE IF NOT EXISTS `metric_comparison` (
    `guid` varchar(64) NOT NULL,
    `comparison_type` varchar(64) default null COMMENT '对比类型: diff 差值,diff_percent 差值百分比',
    `calc_type` varchar(64) default null COMMENT '计算数值: diff 差值,diff_percent 差值百分比',
    `calc_method` varchar(64) default null COMMENT '计算方法: avg平均,sum求和',
    `calc_period` int(11) default 0 COMMENT '计算周期',
    `metric_id` varchar(128) default null COMMENT '指标id',
    `origin_metric_id` varchar(128) default null COMMENT '原始指标id',
    `create_user` varchar(64)  default null COMMENT '创建人',
    `create_time` datetime  default null COMMENT '创建时间',
    PRIMARY KEY (`guid`),
    CONSTRAINT `fore_metric_comparison_metric_id` FOREIGN KEY (`metric_id`) REFERENCES `metric` (`guid`),
    CONSTRAINT `fore_metric_comparison_origin_metric_id` FOREIGN KEY (`origin_metric_id`) REFERENCES `metric` (`guid`)
)ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '指标同环比';