# Prometheus 监控插件

Prometheus是由SoundCloud开发的开源监控报警系统和时序列数据库(TSDB)。Prometheus使用Go语言开发，是Google BorgMon监控系统的开源版本。

Prometheus 监控插件包括几个组成部分： Prometheus Server、Consul、Alert Manager、监控应用程序。

## 技术实现
WeCube通过监控插件来对资源以及应用的监控及告警。

此插件后端技术选型为Java + Spring boot, 前端技术选型为ECharts。


## 主要功能
监控插件包括以下功能：

- agent管理：注册、启动、停止；
- 数据管理：提供数据采集配置， 数据查询等功能；
- 告警管理：提供阈值配置、告警触发等功能；
- 视图管理：提供图形配置功能；

## 设计文档

[Prometheus安装文档](wiki/prometheus_deploy_guide.md)

## 编译打包
插件开发进行中...


## 插件运行
插件包制作完成后，需要通过WeCube的插件管理界面进行注册才能使用。运行插件的主机需提前安装好docker。
