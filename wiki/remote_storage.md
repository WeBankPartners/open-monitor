# Remote Storage Guide

说明： Prometheus推荐的远程存储方式有Opentsdb、Influxdb等，因为Influxdb的集群模式是商业收费版，所以该文档主要示例以Opentsdb为主，并且Open-Monitor自身也集成了Opentsdb的数据读取。  
Prometheus官方文档：https://prometheus.io/docs/operating/integrations/#remote-endpoints-and-storage  

### Opentsdb安装
说明： 因为Opentsdb依赖于hdfs和hbase，所以该文档会描述分布式安装hadoop、hbase、opentsdb的步骤，如果使用者已有相关大数据平台，请直接拉到最下面的opentsdb安装即可。  

