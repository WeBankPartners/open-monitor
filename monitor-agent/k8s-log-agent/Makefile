.PHONY: test build image deploy-template-check
IMAGE ?= monitor-k8s-log-agent:dev
ALPINE_MIRROR ?= mirrors.aliyun.com

test:
	go test -race ./...
build:
	CGO_ENABLED=0 go build -trimpath -o bin/k8s-log-agent ./cmd/k8s-log-agent
image:
	docker build --build-arg ALPINE_MIRROR=$(ALPINE_MIRROR) -t $(IMAGE) .

deploy-template-check:
	python3 -B -m unittest discover -s ../../build/tests -p 'test_k8s_log_agent_deployment.py' -v
