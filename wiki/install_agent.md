#Prometheus Agent 安装手册

Prometheus 提供了一个主动采集指标的标准，相应的软件只需要自己暴露/metrics接口即可，一些主流的软件Prometheus官方提供了相应的agent实现，可从Prometheus官方下载相应agent的最近release二进制文件运行即可。

## Host
Prometheus官方 [node_exporter](https://github.com/prometheus/node_exporter)  
- 下载对应操作系统的最新release版本
- 运行node_exporter
```bash
nohup ./node_exporter > app.log 2>&1 & 
```
- 查看默认提供的9100端口，能看到所暴露的指标数据

## Mysql
Prometheus官方 [mysqld_exporter](https://github.com/prometheus/mysqld_exporter)
- 下载对应操作系统的最新release版本
- 需要在被监控数据库中新建监控用户并授权
```bash
CREATE USER 'exporter'@'localhost' IDENTIFIED BY 'your_password';
GRANT PROCESS, REPLICATION CLIENT, SELECT ON *.* TO 'exporter'@'localhost';
```
- 新建my.cnf配置文件
```bash
[client]
user=exporter
password=your_password
host=127.0.0.1
```
- 运行exporter
```bash
nohub ./mysqld_exporter --config.my-cnf="my.cnf" > app.log 2>&1 &
```
- 查看提供的9104端口，能看到所暴露的指标数据

## Redis
官方并没有提供Redis的exporter，可使用较流行的[redis_exporter](https://github.com/oliver006/redis_exporter)  
- 下载对应操作系统的最新release版本
- 运行redis_exporter
```bash
nohup ./redis_exporter --redis.addr="redis://localhost:6379" --redis.password="prom_pwd" > app.log 2>&1 &
```
- 查看提供的9121端口，能看到所暴露的指标数据

## Tomcat & Jmx
Prometheus官方 [jmx_exporter](https://github.com/prometheus/jmx_exporter)
 下载最新的source并编译jar包，这里提供一个编译好的0.12.0版本的包 [jmx_prometheus_javaagent-0.12.0.jar](exporter/jmx_prometheus_javaagent-0.12.0.jar)  

 tomcat war 部署方式监控，这种方式直接在tomcat里面加上exporter的jar包修改catalina.sh，启动并暴露指标  
 1、新建config.yaml文件
```yaml
lowercaseOutputLabelNames: true
lowercaseOutputName: true
rules:
- pattern: 'Catalina<type=GlobalRequestProcessor, name=\"(\w+-\w+)-(\d+)\"><>(\w+):'
  name: tomcat_$3_total
  labels:
    port: "$2"
    protocol: "$1"
  help: Tomcat global $3
  type: COUNTER
- pattern: 'Catalina<j2eeType=Servlet, WebModule=//([-a-zA-Z0-9+&@#/%?=~_|!:.,;]*[-a-zA-Z0-9+&@#/%=~_|]), name=([-a-zA-Z0-9+/$%~_-|!.]*), J2EEApplication=none, J2EEServer=none><>(requestCount|maxTime|processingTime|errorCount):'
  name: tomcat_servlet_$3_total
  labels:
    module: "$1"
    servlet: "$2"
  help: Tomcat servlet $3 total
  type: COUNTER
- pattern: 'Catalina<type=ThreadPool, name="(\w+-\w+)-(\d+)"><>(currentThreadCount|currentThreadsBusy|keepAliveCount|pollerThreadCount|connectionCount):'
  name: tomcat_threadpool_$3
  labels:
    port: "$2"
    protocol: "$1"
  help: Tomcat threadpool $3
  type: GAUGE
- pattern: 'Catalina<type=Manager, host=([-a-zA-Z0-9+&@#/%?=~_|!:.,;]*[-a-zA-Z0-9+&@#/%=~_|]), context=([-a-zA-Z0-9+/$%~_-|!.]*)><>(processingTime|sessionCounter|rejectedSessions|expiredSessions):'
  name: tomcat_session_$3_total
  labels:
    context: "$2"
    host: "$1"
  help: Tomcat session $3 total
  type: COUNTER
```
 2、把config.yaml和jmx_prometheus_javaagent-0.12.0.jar放到 $TOMCAT_PATH/bin下面  
 3、修改$TOMCAT_PATH/bin下的catalina.sh文件，在 JAVA_OPTS="$JAVA_OPTS $JSSE_OPTS" 后面加上一句
 ```bash
JAVA_OPTS="$JAVA_OPTS -javaagent:$PWD/jmx_prometheus_javaagent-0.12.0.jar=9151:$PWD/config.yaml"
```
4、最后startup.sh启动tomcat，可在提供的9151端口上查到tomcat和jmx的相关指标