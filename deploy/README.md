# GitOps MCP Server 部署指南

## 目录结构

```
deploy/
├── Dockerfile              # 多阶段构建镜像
├── README.md               # 本文档
├── docker/
│   ├── docker-compose.yml  # Docker Compose 配置
│   └── .env.example        # 环境变量模板
└── k8s/
    ├── namespace.yaml      # 命名空间
    ├── secret.yaml         # 敏感配置（GitHub Token）
    ├── configmap.yaml      # 非敏感配置
    ├── deployment.yaml     # 部署清单
    └── service.yaml        # 服务暴露
```

---

## 方式一：本地二进制（开发）

```bash
# 编译
make build

# stdio 模式（Claude Code 直接使用）
./bin/gitops-mcp

# SSE 模式
MCP_TRANSPORT=sse GITHUB_TOKEN=ghp_xxx ./bin/gitops-mcp
```

---

## 方式二：Docker 部署

### 构建镜像

```bash
make docker
```

### 运行

```bash
# 环境变量方式
docker run -d \
  --name gitops-mcp \
  -p 18080:18080 \
  -e GITHUB_TOKEN=ghp_xxx \
  -e MCP_TRANSPORT=sse \
  gitops-mcp-server:latest

# 或使用 docker compose
cd deploy/docker
cp .env.example .env
# 编辑 .env 填入 GITHUB_TOKEN
docker compose up -d
```

### 查看日志

```bash
docker logs -f gitops-mcp-server
```

---

## 方式三：Kubernetes 部署

### 1. 准备 Secret

复制模板并填入真实的 base64 编码 Token：

```bash
cp deploy/k8s/secret.yaml.example deploy/k8s/secret.yaml

# 生成 base64 编码的 Token
echo -n "ghp_your_token_here" | base64
# 输出：Z2hwX3lvdXJfdG9rZW5faGVyZQ==
```

将输出填入 `secret.yaml` 的 `github-token` 字段。

### 2. （可选）自定义配置

编辑 `deploy/k8s/configmap.yaml` 中的非敏感配置项。

### 3. 部署

```bash
make k8s-deploy
```

或手动执行：

```bash
kubectl apply -f deploy/k8s/namespace.yaml
kubectl apply -f deploy/k8s/secret.yaml
kubectl apply -f deploy/k8s/configmap.yaml
kubectl apply -f deploy/k8s/deployment.yaml
kubectl apply -f deploy/k8s/service.yaml
```

### 4. 验证

```bash
# 查看 Pod 状态
kubectl -n gitops-mcp get pods

# 查看日志
kubectl -n gitops-mcp logs -f deploy/gitops-mcp-server

# 测试连接（集群内）
kubectl -n gitops-mcp run curl --rm -it --image=curlimages/curl -- \
  curl http://gitops-mcp-server:18080/health
```

### 5. 清理

```bash
make k8s-delete
```

---

## 安全说明

| 部署方式 | 密钥存储 |
|---|---|
| 本地二进制 | 环境变量或配置文件（明文） |
| Docker | 环境变量（`docker inspect` 可见），生产环境建议用 Docker Secrets |
| Kubernetes | **Secret 资源**（base64 编码），推荐配合 External Secrets Operator 对接 Vault 等 |

> ⚠️ `secret.yaml` 中的 base64 仅是编码，不是加密。请勿将真实 Token 提交到 Git。
> 建议将 `secret.yaml` 加入 `.gitignore`，或使用 CI/CD 流水线动态生成。
