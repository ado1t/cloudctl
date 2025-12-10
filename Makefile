.PHONY: build clean install test lint fmt help run dev

# 变量定义
BINARY_NAME=cloudctl
VERSION?=0.2.0
BUILD_DIR=_output
MAIN_PATH=./cmd/cloudctl
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(shell date -u '+%Y-%m-%d_%H:%M:%S')"

# 默认目标
.DEFAULT_GOAL := help

## help: 显示帮助信息
help:
	@echo "可用的 make 命令:"
	@echo ""
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: 编译项目
build:
	@echo "正在编译 $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "编译完成: $(BUILD_DIR)/$(BINARY_NAME)"

## build-all: 编译所有平台版本
build-all:
	@echo "正在编译所有平台版本..."
	@mkdir -p $(BUILD_DIR)
	# Linux AMD64
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	# Linux ARM64
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	# Windows AMD64
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "所有平台编译完成"

## install: 安装到 GOPATH/bin
install:
	@echo "正在安装 $(BINARY_NAME)..."
	$(GO) install $(LDFLAGS) $(MAIN_PATH)
	@echo "安装完成"

## clean: 清理编译文件
clean:
	@echo "正在清理..."
	@rm -rf $(BUILD_DIR)
	@$(GO) clean
	@echo "清理完成"

## test: 运行测试
test:
	@echo "正在运行测试..."
	$(GO) test -v -race -coverprofile=coverage.out $(shell go list ./... | grep -v /examples)
	@echo "测试完成"

## test-coverage: 运行测试并生成覆盖率报告
test-coverage: test
	@echo "生成覆盖率报告..."
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

## lint: 运行代码检查
lint:
	@echo "正在运行代码检查..."
	@which golangci-lint > /dev/null || (echo "请先安装 golangci-lint: https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...
	@echo "代码检查完成"

## fmt: 格式化代码
fmt:
	@echo "正在格式化代码..."
	$(GO) fmt ./...
	@echo "代码格式化完成"

## vet: 运行 go vet
vet:
	@echo "正在运行 go vet..."
	$(GO) vet ./...
	@echo "go vet 完成"

## tidy: 整理依赖
tidy:
	@echo "正在整理依赖..."
	$(GO) mod tidy
	@echo "依赖整理完成"

## run: 运行程序
run: build
	@echo "正在运行 $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

## dev: 开发模式运行
dev:
	@echo "开发模式运行..."
	$(GO) run $(MAIN_PATH)

## deps: 下载依赖
deps:
	@echo "正在下载依赖..."
	$(GO) mod download
	@echo "依赖下载完成"

## check: 运行所有检查 (fmt, vet, lint, test)
check: fmt vet lint test
	@echo "所有检查完成"

## release: 构建发布版本
release: clean
	@echo "正在构建发布版本 $(VERSION)..."
	@$(MAKE) build-all
	@echo "发布版本构建完成"

## version: 显示版本信息
version:
	@echo "Version: $(VERSION)"
	@echo "Go Version: $(shell $(GO) version)"
