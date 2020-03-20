# Open-Monitor安装指引

## docker运行：

下载release上最新的open-monitor.zip压缩包，装好所依赖docker和mysql
#### 1. 解压open-monitor.zip压缩包，里面会有image.tar镜像和init.sql
#### 2. 在mysql上导入init.sql初始化数据库
#### 3. 导入docker镜像
```
docker load --input image.tar
```
#### 4. 创建本地目录，例如/app/test(如果是其它目录请替换下面命令中的/app/test)，替换如下命令的mysql连接参数
```
MONITOR_DB_HOST=127.0.0.1 -> 把127.0.0.1替换成mysql地址
***REMOVED*** -> 把root替换成mysql用户
***REMOVED*** -> 把wecube替换成mysql用户密码
wecube-monitor:v1.3.0 -> 把后面的版本号改成所导入镜像的版本号
```
```
mkdir -p /app/test
docker run --name open-monitor --volume /app/test/prometheus/logs:/app/monitor/prometheus/logs --volume /app/test/prometheus/data:/app/monitor/prometheus/data --volume /app/test/prometheus/rules:/app/monitor/prometheus/rules  --volume /app/test/alertmanager/logs:/app/monitor/alertmanager/logs --volume /app/test/alertmanager/data:/app/monitor/alertmanager/data --volume /app/test/consul/data:/app/monitor/consul/data --volume /app/test/consul/logs:/app/monitor/consul/logs --volume /app/test/monitor/logs:/app/monitor/monitor/logs --volume /app/test/deploy:/app/deploy --volume /app/test/transgateway/logs:/app/monitor/transgateway/logs --volume /app/test/transgateway/data:/app/monitor/transgateway/data --volume /etc/localtime:/etc/localtime -d -p 8080:8080 -p 8500:8500 -p 8300:8300 -p 9090:9090 -p 19091:19091 -e MONITOR_SERVER_PORT=8080 -e MONITOR_DB_HOST=127.0.0.1 -e ***REMOVED*** -e ***REMOVED*** -e MONITOR_SESSION_ENABLE=true wecube-monitor:v1.3.0
```
容器运行起来后打开 http://127.0.0.1:8080/wecube-monitor/ 登录界面

#### 5. 注册agent
open-monitor.zip解压后里面有exporter_host.tar.gz文件，执行如下安装命令
```
tar zxf exporter_host.tar.gz
cd exporter_host
chmod +x start.sh
./start.sh
```
最后在界面里的 配置->对象->新增 把刚才agent的主机ip地址填入并保存，提示成功后可在 视图->对象视图中搜索查看该主机的性能图表
