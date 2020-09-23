current_dir=$(shell pwd)
version=$(PLUGIN_VERSION)
project_name=$(shell basename "${current_dir}")

clean:
	rm -rf monitor-server/monitor-server
	rm -rf monitor-agent/agent_manager/agent_manager
	rm -rf monitor-agent/archive_mysql_tool/archive_mysql_tool
	rm -rf monitor-agent/node_exporter/node_exporter_new
	rm -rf monitor-agent/ping_exporter/ping_exporter
	rm -rf monitor-agent/db_data_exporter/db_data_exporter
	rm -rf monitor-agent/transgateway/transgateway
	rm -rf monitor-ui/dist
	rm -rf monitor-ui/plugin

build: clean
	chmod +x ./build/*.sh
	docker run --rm -v $(current_dir):/go/src/github.com/WeBankPartners/$(project_name) --name build_monitor_server golang:1.12.5 /bin/bash /go/src/github.com/WeBankPartners/$(project_name)/build/build-server.sh
	./build/build-ui.sh $(current_dir)

image: build
	docker build -t $(project_name):$(version) .

agent:
	chmod +x ./build/build-agent.sh
	chmod +x ./build/conf/node_exporter/*
	./build/build-agent.sh

package: image agent
	mkdir -p plugin
	cp -r monitor-ui/plugin/* plugin/
	zip -r ui.zip plugin
	rm -rf plugin
	cp build/register.xml ./
	cp wiki/db/monitor_sql_01_struct.sql ./init.sql
	cat wiki/db/monitor_sql_02_base_data_en.sql >> ./init.sql
	sed -i "s~{{PLUGIN_VERSION}}~$(version)~g" ./register.xml
	docker save -o image.tar $(project_name):$(version)
	zip  wecube-plugins-monitor-$(version).zip image.tar init.sql register.xml ui.zip node_exporter.tar.gz
	rm -f register.xml
	rm -f init.sql
	rm -f ui.zip
	rm -rf ./*.tar
	rm -f node_exporter.tar.gz
	docker rmi $(project_name):$(version)

upload: package
	$(eval container_id:=$(shell docker run -v $(current_dir):/package -itd --entrypoint=/bin/sh minio/mc))
	docker exec $(container_id) mc config host add wecubeS3 $(s3_server_url) $(s3_access_key) $(s3_secret_key) wecubeS3
	docker exec $(container_id) mc cp /package/wecube-plugins-monitor-$(version).zip wecubeS3/wecube-plugin-package-bucket
	docker rm -f $(container_id)
	rm -rf wecube-plugins-monitor-$(version).zip