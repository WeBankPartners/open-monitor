# Open-Monitor安装指引

## docker运行：
  
#### 如果已存在docker镜像，请忽略1、2、4步
#### 1. 下载release上最新的open-monitor.zip压缩包，装好所依赖docker和mysql
#### 2. 解压open-monitor.zip压缩包，里面会有image.tar镜像和init.sql
#### 3. 在mysql上导入init.sql初始化数据库，如果是zip包解压的会有init.sql，如果是项目编译的，sql文件在 wiki/db/ 下，monitor_sql_01_struct.sql是表结构文件，monitor_sql_02_base_data_*.sql是对应的中英版本的初始化数据
#### 4. 导入docker镜像
```
docker load --input image.tar
```
#### 5. docker运行(docker或docker-compose)：
docker run运行：  
创建本地目录，例如/app/test(如果是其它目录请替换下面命令中的/app/test)，替换如下命令的mysql连接参数  
```
MONITOR_DB_HOST=127.0.0.1 -> 把127.0.0.1替换成mysql地址
MONITOR_DB_PORT=3306  -> 把3306替换成mysql端口
MONITOR_DB_USER=root -> 把root替换成mysql用户
MONITOR_DB_PWD=defalut -> 把defalut替换成mysql用户密码
monitor:v1.3.0 -> 把后面的版本号改成所导入镜像的版本号
```
```
mkdir -p /app/test
docker run --name open-monitor --volume /app/test/prometheus/logs:/app/monitor/prometheus/logs --volume /app/test/prometheus/data:/app/monitor/prometheus/data --volume /app/test/prometheus/rules:/app/monitor/prometheus/rules  --volume /app/test/alertmanager/logs:/app/monitor/alertmanager/logs --volume /app/test/alertmanager/data:/app/monitor/alertmanager/data --volume /app/test/consul/data:/app/monitor/consul/data --volume /app/test/consul/logs:/app/monitor/consul/logs --volume /app/test/monitor/logs:/app/monitor/monitor/logs --volume /app/test/deploy:/app/deploy --volume /app/test/transgateway/logs:/app/monitor/transgateway/logs --volume /app/test/transgateway/data:/app/monitor/transgateway/data --volume /etc/localtime:/etc/localtime -d -p 8080:8080 -p 8500:8500 -p 8300:8300 -p 9090:9090 -p 19091:19091 -e MONITOR_SERVER_PORT=8080 -e MONITOR_DB_HOST=127.0.0.1 -e MONITOR_DB_PORT=3306 -e MONITOR_DB_USER=root -e MONITOR_DB_PWD=wecube -e MONITOR_SESSION_ENABLE=true -e MONITOR_LOG_LEVEL=info -e MONITOR_ALARM_ALIVE_MAX_DAY=30  monitor:v1.3.0
```
docker-compose运行：  
去github上拷贝一份open-monitor的docker-compose.tpl，重命名为docker-compose.yml   
修改docker-compose.yml里的一些参数
```
{{version}} -> open-monitor镜像版本
{{path}} -> 容器文件映射到主机的目录
{{port}} -> http对外服务端口
{{db_host}} -> mysql连接主机
{{db_port}} -> mysql连接端口
{{db_user}} -> mysql连接用户
{{db_password}} -> mysql连接密码
{{monitor_host_ip}} -> 容器运行所在主机ip
```
最后运行docker-compose
```
docker-compose -f docker-compose.yml up -d
```
容器运行起来后打开 http://127.0.0.1:8080/monitor/ 登录界面

#### 6. 注册agent
open-monitor.zip解压后里面有exporter_host.tar.gz文件，执行如下安装命令
```
tar zxf exporter_host.tar.gz
cd exporter_host
chmod +x start.sh
./start.sh
```
最后在界面里的 配置->对象->新增 把刚才agent的主机ip地址填入并保存，提示成功后可在 视图->对象视图中搜索查看该主机的性能图表
