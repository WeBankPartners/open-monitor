version: '2'
services:
  consul:
    image: consul
    container_name: consul
    restart: always
    volumes:
      - ~/data/docker/consul:/consul/data
    ports:
      - "8300:8300"
      - "8400:8400"
      - "8500:8500"
    networks:
      - monitor
  alertmanager:
    image: prom/alertmanager
    container_name: alertmanager
    restart: always
    volumes:
      - ~/data/docker/alertmanager:/alertmanager
      - ~/app/docker/alertmanager:/etc/alertmanager
    ports:
      - "9093:9093"
    networks:
      - monitor
    command:
      - --config.file=/etc/alertmanager/alertmanager.yml
      - --web.listen-address=:9093
      - --cluster.listen-address=:9094
  prometheus:
    image: prom/prometheus
    container_name: prometheus
    restart: always
    volumes:
      - ~/data/docker/prometheus:/prometheus
      - ~/app/docker/prometheus:/etc/prometheus
    ports:
      - "9090:9090"
    networks:
      - monitor
    command:
      - --config.file=/etc/prometheus/prometheus.yml
      - --web.enable-lifecycle
  monitor-db:
    image: {{MONITOR_DATABASE_IMAGE_NAME}}
    restart: always
    command: [
            '--character-set-server=utf8mb4',
            '--collation-server=utf8mb4_unicode_ci',
            '--default-time-zone=+8:00',
            '--max_allowed_packet=4M'
    ]
    volumes:
      - /etc/localtime:/etc/localtime
      - ~/data/docker/monitor-db:/var/lib/mysql
    environment:
      - MYSQL_ROOT_PASSWORD={{MYSQL_ROOT_PASSWORD}}
    ports:
      - 3306:3306
    networks:
      - monitor
  monitor-server:
    image: {{MONITOR_IMAGE_NAME}}
    restart: always
    volumes:
      - /etc/localtime:/etc/localtime
      - ~/app/docker/monitor/conf:/app/monitor/conf
      - ~/app/docker/monitor/logs:/app/monitor/logs
      - ~/app/docker/prometheus/rules:/app/monitor/conf/rules
    ports:
      - {{MONITOR_SERVER_PORT}}:{{MONITOR_SERVER_PORT}}
    networks:
      - monitor

networks:
  monitor:
    driver: bridge

