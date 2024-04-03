# Monitor 独立版本界面部署文档

monitor的独立前端镜像

### docker-compose部署

应用包解压后会有镜像文件：monitor-standalone-ui.tar
```bash
# 导入docker镜像
docker load --input monitor-standalone-ui.tar
```
#### 环境变量整理

**请按需修正一下环境变量值**

```bash
# 独立版本界面版本
VERSION='v1.0.1'
# 部署的主机ip
HOSTIP='10.0.0.1'
# 网页访问的端口
PORT='8901'
# wecube platform-gateway的ip地址
WECUBE_GATEWAY_HOST='10.0.0.2'
# wecube platform-gateway的端口
WECUBE_GATEWAY_PORT='8005'
```

#### 准备yaml内容

在本地准备以下yaml文件

monitor-standalone-ui.yml

```yaml
version: '3'
services:
  monitor-standalone:
    image: monitor-standalone-ui:{{VERSION}}
    container_name: monitor-standalone-{{VERSION}}
    restart: always
    volumes:
      - /etc/localtime:/etc/localtime
      - /data/app/monitor-ui/log:/var/log/nginx/
    ports:
      - "{{HOSTIP}}:{{PORT}}:8080"
    environment:
      - GATEWAY_HOST={{WECUBE_GATEWAY_HOST}}
      - GATEWAY_PORT={{WECUBE_GATEWAY_PORT}}
      - PUBLIC_DOMAIN={{HOSTIP}}:{{PORT}}
      - TZ=Asia/Shanghai
    command: /bin/bash -c "/etc/nginx/start_monitor.sh"
```

#### 修正yaml内容的值

```bash
# 请修改以下变量为正确值
# 粘贴以上整理的环境变量


sed -i "s/{{VERSION}}/$VERSION/g" monitor-standalone-ui.yml
sed -i "s/{{HOSTIP}}/$HOSTIP/g" monitor-standalone-ui.yml
sed -i "s/{{PORT}}/$PORT/g" monitor-standalone-ui.yml
sed -i "s/{{WECUBE_GATEWAY_HOST}}/$WECUBE_GATEWAY_HOST/g" monitor-standalone-ui.yml
sed -i "s/{{WECUBE_GATEWAY_PORT}}/$WECUBE_GATEWAY_PORT/g" monitor-standalone-ui.yml
```

#### 启动docker容器
- 启动monitor-standalone-ui服务
  ```bash
  # 建映射出来的日志路径
  mkdir -p /data/app/monitor-ui/log
  chmod 777 /data/app/monitor-ui/log
  # 启动容器
  docker-compose -f monitor-standalone-ui.yml up -d
  ```

- 验证服务，访问 http://{{HOSTIP}}:{{PORT}}

