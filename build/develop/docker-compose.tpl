version: '2'
services:
  consul:
    image: consul
    container_name: consul
    restart: always
    volumes:
      - {{MONITOR_BASE_PATH}}/consul/data:/consul/data
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
      - alertmanager-data:/alertmanager
      - {{MONITOR_BASE_PATH}}/alertmanager/conf:/etc/alertmanager
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
      - prometheus-data:/prometheus
      - {{MONITOR_BASE_PATH}}/prometheus/conf:/etc/prometheus
    ports:
      - "9090:9090"
    networks:
      - monitor
    command:
      - --config.file=/etc/prometheus/prometheus.yml
      - --web.enable-lifecycle
  node_exporter:
    image: prom/node-exporter
    container_name: node_exporter
    restart: always
    ports:
      - "9100:9100"
    networks:
      - monitor
  monitor-db:
    image: monitor-db:dev
    restart: always
    command: [
            '--character-set-server=utf8mb4',
            '--collation-server=utf8mb4_unicode_ci',
            '--default-time-zone=+8:00',
            '--max_allowed_packet=4M'
    ]
    volumes:
      - {{MONITOR_BASE_PATH}}/monitor-db:/var/lib/mysql
    environment:
      - MYSQL_ROOT_PASSWORD=wecube
    ports:
      - 3306:3306
    networks:
      - monitor

networks:
  monitor:
    driver: bridge

volumes:
  prometheus-data:
    driver: local
    driver_opts:
      o: bind
      type: none
      device: {{MONITOR_BASE_PATH}}/prometheus/data
  alertmanager-data:
    driver: local
    driver_opts:
      o: bind
      type: none
      device: {{MONITOR_BASE_PATH}}/alertmanager/data