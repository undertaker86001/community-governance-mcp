# Higress社区治理Agent Makefile

# 变量定义
BINARY_NAME=higress-agent
BUILD_DIR=build
DOCKER_IMAGE=higress-agent
DOCKER_TAG=latest

# Go相关变量
GO=go
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

# 版本信息
VERSION?=1.0.0
COMMIT_HASH=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# 构建标志
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.CommitHash=$(COMMIT_HASH) -X main.BuildTime=$(BUILD_TIME)"

# 默认目标
.PHONY: all
all: clean build

# 清理构建目录
.PHONY: clean
clean:
	@echo "清理构建目录..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)

# 安装依赖
.PHONY: deps
deps:
	@echo "安装依赖..."
	@$(GO) mod download
	@$(GO) mod tidy

# 构建二进制文件
.PHONY: build
build: deps
	@echo "构建二进制文件..."
	@mkdir -p $(BUILD_DIR)
	@$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/agent

# 构建特定平台的二进制文件
.PHONY: build-linux
build-linux:
	@echo "构建Linux二进制文件..."
	@GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/agent

.PHONY: build-darwin
build-darwin:
	@echo "构建macOS二进制文件..."
	@GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/agent

.PHONY: build-windows
build-windows:
	@echo "构建Windows二进制文件..."
	@GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/agent

# 构建所有平台
.PHONY: build-all
build-all: build-linux build-darwin build-windows

# 运行服务
.PHONY: run
run: build
	@echo "启动服务..."
	@$(BUILD_DIR)/$(BINARY_NAME)

# 开发模式运行
.PHONY: dev
dev:
	@echo "开发模式运行..."
	@$(GO) run ./cmd/agent

# 测试
.PHONY: test
test:
	@echo "运行测试..."
	@$(GO) test -v ./...

.PHONY: test-coverage
test-coverage:
	@echo "运行测试并生成覆盖率报告..."
	@$(GO) test -v -coverprofile=coverage.out ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

# 代码检查
.PHONY: lint
lint:
	@echo "运行代码检查..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint未安装，跳过代码检查"; \
	fi

# 格式化代码
.PHONY: fmt
fmt:
	@echo "格式化代码..."
	@$(GO) fmt ./...

# 生成文档
.PHONY: docs
docs:
	@echo "生成文档..."
	@if command -v godoc >/dev/null 2>&1; then \
		godoc -http=:6060; \
	else \
		echo "godoc未安装，跳过文档生成"; \
	fi

# Docker相关
.PHONY: docker-build
docker-build:
	@echo "构建Docker镜像..."
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

.PHONY: docker-run
docker-run: docker-build
	@echo "运行Docker容器..."
	@docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-push
docker-push:
	@echo "推送Docker镜像..."
	@docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

# 发布
.PHONY: release
release: clean build-all
	@echo "创建发布包..."
	@mkdir -p release
	@cp $(BUILD_DIR)/* release/
	@cp configs/config.yaml release/
	@cp .env.example release/
	@cp README.md release/
	@tar -czf release/higress-agent-$(VERSION).tar.gz -C release .
	@echo "发布包已创建: release/higress-agent-$(VERSION).tar.gz"

# 安装
.PHONY: install
install: build
	@echo "安装到系统..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "安装完成，可以使用 'higress-agent' 命令"

# 卸载
.PHONY: uninstall
uninstall:
	@echo "卸载..."
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "卸载完成"

# 帮助
.PHONY: help
help:
	@echo "可用的Make目标:"
	@echo "  all          - 清理并构建"
	@echo "  clean        - 清理构建目录"
	@echo "  deps         - 安装依赖"
	@echo "  build        - 构建二进制文件"
	@echo "  build-linux  - 构建Linux版本"
	@echo "  build-darwin - 构建macOS版本"
	@echo "  build-windows- 构建Windows版本"
	@echo "  build-all    - 构建所有平台版本"
	@echo "  run          - 运行服务"
	@echo "  dev          - 开发模式运行"
	@echo "  test         - 运行测试"
	@echo "  test-coverage- 运行测试并生成覆盖率报告"
	@echo "  lint         - 代码检查"
	@echo "  fmt          - 格式化代码"
	@echo "  docs         - 生成文档"
	@echo "  docker-build - 构建Docker镜像"
	@echo "  docker-run   - 运行Docker容器"
	@echo "  docker-push  - 推送Docker镜像"
	@echo "  release      - 创建发布包"
	@echo "  install      - 安装到系统"
	@echo "  uninstall    - 卸载"
	@echo "  help         - 显示此帮助信息"

# 检查环境
.PHONY: check-env
check-env:
	@echo "检查环境..."
	@if [ ! -f .env ]; then \
		echo "警告: .env文件不存在，请复制.env.example并配置"; \
	fi
	@if [ -z "$$OPENAI_API_KEY" ]; then \
		echo "警告: OPENAI_API_KEY环境变量未设置"; \
	fi
	@if [ -z "$$GITHUB_TOKEN" ]; then \
		echo "警告: GITHUB_TOKEN环境变量未设置"; \
	fi

# 初始化项目
.PHONY: init
init: check-env
	@echo "初始化项目..."
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "已创建.env文件，请编辑配置"; \
	fi
	@mkdir -p logs
	@mkdir -p data/knowledge
	@echo "项目初始化完成"

# 显示版本信息
.PHONY: version
version:
	@echo "版本: $(VERSION)"
	@echo "提交: $(COMMIT_HASH)"
	@echo "构建时间: $(BUILD_TIME)"