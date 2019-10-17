# Wecube-plugins-prometheus 监控插件

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![](https://img.shields.io/badge/language-go-orange.svg)
![](https://img.shields.io/badge/language-vue-green.svg)

## 试用链接
[点此试用Wecube-plugins-prometheus](http://134.175.254.251)

## 引言
Prometheus是由SoundCloud开发的开源监控报警系统和时序列数据库(TSDB)。Prometheus使用Go语言开发，是Google BorgMon监控系统的开源版本。

Prometheus 监控插件包括几个组成部分： Prometheus Server、Consul、Alert Manager、监控应用程序。

Wecube-plugins-prometheus 无侵入式地封装了Prometheus的功能，并提供更好的告警管理和图表展示，以及与其它系统的交互等

## 技术实现
WeCube通过监控插件来对资源以及应用的监控及告警。

此插件底层引用Prometheus，上层Monitor封装了对Prometheus的配置管理和图表展示，Monitor后端技术选型为Go + Gin + Xorm, 前端技术选型为Vue + ECharts。


## 主要功能
监控插件包括以下功能：

- agent管理：注册、启动、停止；
- 数据管理：提供数据采集配置， 数据查询等功能；
- 告警管理：提供阈值配置、日志监控、告警触发等功能；
- 视图管理：提供图形配置和自定义视图功能；

## 快速入门
Wecube-plugins-prometheus采用容器化部署。

如何编译，请查看以下文档
[Wecube-plugins-prometheus编译文档](wiki/compile_guide.md)

如何安装， 请查看以下文档
[Wecube-plugins-prometheus部署文档](wiki/install_guide.md)
