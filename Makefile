export GOPATH=$(PWD)

current_dir=$(shell pwd)
version=$(shell bash ./build/version.sh)
project_name=$(shell basename "${current_dir}" )


APP_HOME=src/github.com/WeBankPartners/wecube-plugins-prometheus

ifndef RUN_MODE
  RUN_MODE=dev
endif

clean-server:
	rm -rf monitor-server/monitor-server

clean-ui:
	rm -rf monitor-ui/dist

build-server: clean-server
	chmod +x ./build/*.sh
	docker run --rm -v $(current_dir):/go/src/github.com/WeBankPartners/$(project_name) --name build_$(project_name)_server golang:1.12.5 /bin/bash /go/src/github.com/WeBankPartners/$(project_name)/build/build-server.sh

build-ui: clean-ui
	chmod +x ./build/*.sh
	docker run --rm -v $(current_dir):/go/src/github.com/WeBankPartners/$(project_name) --name build_$(project_name)_ui node:12.10.0 /bin/bash /go/src/github.com/WeBankPartners/$(project_name)/build/build-ui.sh

image: build-server build-ui
	docker build -t monitor:$(version) .

