version: '2'
services:
  open-monitor:
    image: open-monitor:{{version}}
    container_name: open-monitor-{{version}}
    restart: always
    volumes:
      - /etc/localtime:/etc/localtime
      - {{path}}/monitor/logs:/app/monitor/monitor/logs
      - {{path}}/prometheus/logs:/app/monitor/prometheus/logs
      - {{path}}/prometheus/data:/app/monitor/prometheus/data
      - {{path}}/prometheus/rules:/app/monitor/prometheus/rules
      - {{path}}/alertmanager/logs:/app/monitor/alertmanager/logs
      - {{path}}/alertmanager/data:/app/monitor/alertmanager/data
      - {{path}}/transgateway/data:/app/monitor/transgateway/data
      - {{path}}/deploy:/app/deploy
    ports:
      - "{{port}}:8080"
      - "9090:9090"
      - "9093:9093"
      - "19091:19091"
      - "14241:14241"
    environment:
      - MONITOR_LOG_LEVEL=info
      - MONITOR_DB_HOST={{db_host}}
      - MONITOR_DB_PORT={{db_port}}
      - MONITOR_DB_USER={{db_user}}
      - MONITOR_DB_PWD={{db_password}}
      - MONITOR_SESSION_ENABLE=true
      - MONITOR_HOST_IP={{monitor_host_ip}}
      - DEMO_FLAG=false
      - MONITOR_ALARM_ALIVE_MAX_DAY=1
      - MONITOR_ARCHIVE_ENABLE=N

