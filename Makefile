current_dir=$(shell pwd)
version=$(PLUGIN_VERSION)
project_name=$(shell basename "${current_dir}")

clean:
	rm -rf monitor-server/monitor-server
	rm -rf monitor-agent/agent_manager/agent_manager
	rm -rf monitor-agent/archive_mysql_tool/archive_mysql_tool
	rm -rf monitor-agent/node_exporter/node_exporter_new
	rm -rf monitor-agent/ping_exporter/ping_exporter
	rm -rf monitor-agent/transgateway/transgateway
	rm -rf monitor-ui/dist
	rm -rf monitor-ui/plugin

build: clean
	chmod +x ./build/*.sh
	docker run --rm -v $(current_dir):/go/src/github.com/WeBankPartners/$(project_name) --name build_monitor_server golang:1.12.5 /bin/bash /go/src/github.com/WeBankPartners/$(project_name)/build/build-server.sh
	./build/build-ui.sh $(current_dir)

image: build
	docker build -t $(project_name):$(version) .

