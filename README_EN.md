# WeCube-plugins-prometheus

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![](https://img.shields.io/badge/language-go-orange.svg)
![](https://img.shields.io/badge/language-vue-green.svg)

English / [中文](README.md)

## Link for Trial

[Click_To_Try_WeCube-plugins-prometheus](https://sandbox.webank.com/wecube-monitor)

## Demo
<img src="./wiki/images/wecube-monitor01.gif" />

## Introduction

Prometheus is an open-source monitor-alarm system and time series database (TSDB) developed by SoundCloud. It is an open-source version of the Google BorgMon monitoring system in Go language.

The Prometheus Monitoring Plugin consists of several components: `Prometheus Server`, `Consul`, `Alert Manager`, and `Monitoring Applications`.

WeCube-plugins-prometheus encapsulates the functionality of Prometheus without intrusion and provides better alarm management and graphic dashboard, as well as interaction with other systems.

## System Architecture

The overall architecture diagram is as follows:

![WeCube-plugins-prometheus_Architecture](wiki/images/Architecture.svg)

## Summary

WeCube monitors, alerts resources and applications through monitoring plugins.

The plugin is based on Prometheus. The upper layer `Monitor` encapsulates the configuration management and chart display of Prometheus. The `Monitor` backend technology is written by Go + Gin + Xorm, and the front-end technology is written by Vue + ECharts.

**Monitor has the following features：**

- Endpoint Management

  The monitor supplies register and de-register endpoint functions. It synchronizes endpoint from CMDB when it connects to CMDB. It supports group management of endpoints and customization of the the alarm configuration.

- Friendly Dashboard

  The monitor supports mainstream monitoring types, including the host, MySQL, Redis, Tomcat, etc.  
   It also supports Prometheus' native `PromQL` query and its metric configuration.
  The dashboard customization is friendly, too.

- Alarm Management

  It provides the persistence and the distribution of Prometheus alert rules, and it also supports the manifestation of un-recovered alarm panels and historical alarms.
  User can customize endpoint alarm configuration, group alert configuration and receiver management of alarms.

## Main Features

- Endpoint Management: register, start and stop functions.
- Data Management: data collection configuration and data query functions.
- Alarm Management: threshold configuration, log monitoring, alarm triggering functions.
- Dashboard Management: graphical configuration and custom dashboard functions.

## Quick Start

WeCube-plugins-prometheus is deployed in a docker container.

Please refer to the [WeCube-plugins-prometheus_Compiling_Guide](wiki/compile_guide.md) on how to compile WeCube-plugins-prometheus.

Please refer to the [WeCube-plugins-prometheus_Deployment_Guide](wiki/install_guide.md) on how to install WeCube-plugins-prometheus.

## License

WeCube-plugins-prometheus is licensed under the Apache License Version 2.0 , please refer to the [license](LICENSE) for details.

## Community

- For quick response, please [raise_an_issue](https://github.com/WeBankPartners/wecube-plugins-prometheus/issues/new/choose) to us, or you can also scan the following QR code to join our community, we will provide feedback as quickly as we can.

	<div align="left">
	<img src="wiki/images/wecube_qr_code.png"  height="200" width="200">
	</div>

* Contact us: fintech@webank.com
