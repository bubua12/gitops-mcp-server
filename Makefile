BINARY_NAME := gitops-mcp
IMAGE_NAME := gitops-mcp-server
VERSION := 1.0.0

.PHONY: build clean docker docker-push

## build: 编译二进制
build:
	go build -ldflags="-s -w" -o bin/$(BINARY_NAME) ./cmd/gitops-mcp

## clean: 清理构建产物
clean:
	rm -rf bin/

## docker: 构建 Docker 镜像
docker:
	docker build -t $(IMAGE_NAME):$(VERSION) -f deploy/Dockerfile .
	docker tag $(IMAGE_NAME):$(VERSION) $(IMAGE_NAME):latest

## docker-push: 推送 Docker 镜像（需要先设置 IMAGE_REPO）
docker-push:
	docker tag $(IMAGE_NAME):$(VERSION) $(IMAGE_REPO)/$(IMAGE_NAME):$(VERSION)
	docker tag $(IMAGE_NAME):$(VERSION) $(IMAGE_REPO)/$(IMAGE_NAME):latest
	docker push $(IMAGE_REPO)/$(IMAGE_NAME):$(VERSION)
	docker push $(IMAGE_REPO)/$(IMAGE_NAME):latest

## k8s-deploy: 部署到 K8s（先创建 Secret）
k8s-deploy:
	@if [ ! -f deploy/k8s/secret.yaml ]; then \
		echo ">>> Error: deploy/k8s/secret.yaml not found!"; \
		echo ">>> Run: cp deploy/k8s/secret.yaml.example deploy/k8s/secret.yaml"; \
		echo ">>> Then edit it with your base64-encoded GitHub token."; \
		exit 1; \
	fi
	@echo ">>> Creating namespace..."
	kubectl apply -f deploy/k8s/namespace.yaml
	@echo ">>> Applying Secret..."
	kubectl apply -f deploy/k8s/secret.yaml
	@echo ">>> Applying ConfigMap..."
	kubectl apply -f deploy/k8s/configmap.yaml
	@echo ">>> Deploying..."
	kubectl apply -f deploy/k8s/deployment.yaml
	kubectl apply -f deploy/k8s/service.yaml
	@echo ">>> Done! Check status:"
	kubectl -n gitops-mcp get pods

## k8s-delete: 删除 K8s 部署
k8s-delete:
	kubectl delete -f deploy/k8s/ --ignore-not-found

## help: 显示帮助
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | column -t -s ':'
