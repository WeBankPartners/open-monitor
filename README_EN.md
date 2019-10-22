# WeCube-plugins-prometheus 

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![](https://img.shields.io/badge/language-go-orange.svg)
![](https://img.shields.io/badge/language-vue-green.svg)

English / [中文](README.md)

## Link for Trial
[Click_To_Try_WeCube-plugins-prometheus](http://134.175.254.251)

## Introduction
Prometheus is a open-source monitor system develop by SoundCloud. It written by Go language, and is Google BorgMon open-source version.  

Prometheus contain these parts: Prometheus Server, Alertmanager, Exporters

WeCube-plugins-prometheus non-invasively package Prometheus, provide better alarm management and graph dashboard.


## System Architecture
Architecture is as follows:  

![WeCube-plugins-prometheus_Architecture](wiki/images/Architecture.svg)


## Summary
WeCube monitor endpoint and alarm with WeCube-plugins-prometheus.

WeCube-plugins-prometheus package the config manager and graph display from Prometheus with Monitor. 
Monitor-server select Go + Gin + Xorm technology, Monitor-ui select Vue + ECharts.

**Here are the characteristics of Monitor：**

- Endpoint Management

    Provide register and deregister endpoint function. It can synchronize endpoint from CMDB when it connect to CMDB.
    Support group management of endpoint and can be used for alarm configuration.
    
- Friendly Dashboard

    Provide the dashboard of main monitoring type, including host, mysql, redis, tomcat, etc.    
    Provide Prometheus native PromQL query and query metric configuration.
    Provide the ability to customize dashboard.
    
- Alarm Management

    Provide persistence and distribution of Prometheus alert rule configuration.
    Provide display of unrecovered alarm panel and historical alarm.
    Provide endpoint alarm configuration and group alarm configuration.
    Provide alarm receiver management.  
    

## Main Features

- Endpoint Management: Provide register, start and stop functions.  
- Data Management: Provide data collection configuration and data query functions.
- Alarm Management: Provide threshold configuration, log monitoring, alarm triggering functions.
- Dashboard Management: Provide graphic configuration and custom dashboard functions.


## Quick Start
Normal deployment with containers of WeCube-plugins-prometheus

Please refer to the [WeCube-plugins-prometheus_Compiling_Guide](wiki/compile_guide.md) on how to compile WeCube-plugins-prometheus.

Please refer to the [WeCube-plugins-prometheus_Deployment_Guide](wiki/install_guide.md) on how to install WeCube-plugins-prometheus.

## License
WeCube-plugins-prometheus is licensed under the Apache License Version 2.0 , please refer to the [license](LICENSE) for details.

## Community
- For quick response, please [raise_an_issue](https://github.com/WeBankPartners/wecube-plugins-prometheus/issues/new/choose) to us, or you can also scan the following QR code to join our community, we will provide feedback as quickly as we can.

	<div align="left">
	<img src="wiki/images/wecube_qr_code.png"  height="200" width="200">
	</div>


- Contact us: fintech@webank.com