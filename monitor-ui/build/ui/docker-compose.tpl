version: '3'
services:
  monitor-standalone:
    image: monitor-standalone:{{version}}
    container_name: monitor-standalone-{{version}}
    restart: always
    volumes:
      - /etc/localtime:/etc/localtime
      - [#path]/monitor-standalone/log:/var/log/nginx/
    ports:
      - "[@HOSTIP]:[#http_port]:8080"
    environment:
      - GATEWAY_HOST=[#GATEWAY_HOST]
      - GATEWAY_PORT=[#GATEWAY_PORT]
      - PUBLIC_DOMAIN=[#PUBLIC_DOMAIN]
      - TZ=Asia/Shanghai
    command: /bin/bash -c "/etc/nginx/start_monitor.sh"
