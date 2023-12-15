# Open-Monitor

<p align="left">
    <a href="https://opensource.org/licenses/Apache-2.0" alt="License">
        <img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" /></a>
    <a href="https://github.com/WeBankPartners/open-monitor/tree/v1.0.1" alt="release">
        <img src="https://img.shields.io/github/v/release/WeBankPartners/open-monitor.svg" /></a>
    <a href="#" alt="Code Size">
        <img src="https://img.shields.io/github/languages/code-size/WeBankPartners/open-monitor.svg" /></a>
    <a href="#" alt="Java">
        <img src="https://img.shields.io/badge/language-go-orange.svg" /></a>
    <a href="#" alt="Vue">
        <img src="https://img.shields.io/badge/language-vue-green.svg" /></a>
    <a href="https://github.com/WeBankPartners/open-monitor/graphs/contributors" alt="Contributors">
        <img src="https://img.shields.io/github/contributors/WeBankPartners/open-monitor" /></a>
    <a href="https://github.com/WeBankPartners/open-monitor/pulse" alt="Activity">
        <img src="https://img.shields.io/github/commit-activity/m/WeBankPartners/open-monitor" /></a>
</p>

English / [中文](README.md)

## Experience

<img src="./wiki/images/wecube-monitor01.gif" />  
if you want to get better experience, please set up your private environment refer to: [Open-Monitor_Deployment_Guide](wiki/install_guide.md)

## Introduction

Prometheus is an open-source monitor-alarm system and time series database (TSDB) developed by SoundCloud. It is an open-source version of the Google BorgMon monitoring system in Go language.

Open-Monitor encapsulates the functionality of Prometheus without intrusion and provides better alarm management and graphic dashboard, as well as interaction with other systems.  

Open-Monitor consists of several components: `Prometheus`, `Alert Manager`, `Monitor`, `Agent_manager`, `Ping_exporter`, `Archive_mysql_tool`.

## System Architecture

The overall architecture diagram is as follows:

![Open-Monitor_Architecture](wiki/images/Architecture-new.svg)

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

Open-Monitor is deployed in a docker container.

Please refer to the [Open-Monitor_Compiling_Guide](wiki/compile_guide_new.md) on how to compile Open-Monitor.

Please refer to the [Open-Monitor_Deployment_Guide](wiki/install_guide.md) on how to install Open-Monitor.

## User Manuals

Please refer to [Open-Monitor User Guide](wiki/user_guide.md) for usage and operations

## Developer Guide
Develop Open-Monitor in Normal Mode  
Please refer to the [Open-Monitor Develop Doc](wiki/develop_local_guide.md) for setting up local environment quickly.

## License

Open-Monitor is licensed under the Apache License Version 2.0 , please refer to the [license](LICENSE) for details.

## Community

- For quick response, please [raise_an_issue](https://github.com/WeBankPartners/open-monitor/issues/new/choose) to us, or you can also scan the following QR code to join our community, we will provide feedback as quickly as we can.

	<div align="left">
	<img src="wiki/images/wecube_qr_code.png"  height="200" width="200">
	</div>

* Contact us: fintech@webank.com
