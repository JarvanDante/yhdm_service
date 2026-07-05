.PHONY: tidy run dev build test

tidy:
	go mod tidy

# 生产配置启动
run:
	go run ./cmd/server -conf configs/config.yaml

# 开发配置启动（跳过谷歌 2FA / IP 白名单）
dev:
	go run ./cmd/server -conf configs/config.dev.yaml

build:
	go build -o bin/yhdm_service ./cmd/server

test:
	go test ./...
