version: '2'
services:
  consul:
    image: consul
    container_name: prom-consul
  	restart: always
    dns_search: .
    volumes:
      - consul-data:/consul/data
    ports:
      - "8300:8300"
      - "8400:8400"
      - "8500:8500"
    networks:
      - monitor
  alertmanager:
    image: prom/alertmanager
    container_name: prom-alertmanager
  	restart: always
    dns_search: .
    volumes:
      - alertmanager-data:/alertmanager
      - /app/docker/alertmanager:/etc/alertmanager
    ports:
      - "9300:9300"
    networks:
      - monitor
  prometheus:
    image: prom/prometheus
    container_name: prom-promethues
  	restart: always
    dns_search: .
    volumes:
      - prometheus-tsdb:/prometheus
      - /app/docker/prometheus:/etc/prometheus
    ports:
      - "9090:9090"
    networks:
      - monitor

networks:
  monitor:
    driver: bridge
