# KCP CLI 프로젝트 루트 Makefile

.PHONY: all build clean test lint db-up db-down

# 전체 빌드
all: build

# Go 프로젝트 빌드
build:
	@echo "=== kcp-cli 빌드 ==="
	cd kcp-cli && go build -o ../bin/kcp ./cmd/cli/
	cd kcp-cli && go build -o ../bin/kcp-tui ./cmd/tui/
	@echo "=== kcp-gateway 빌드 ==="
	cd kcp-gateway && go build -o ../bin/kcp-gateway ./cmd/server/

# 전체 테스트
test:
	@echo "=== kcp-cli 테스트 ==="
	cd kcp-cli && go test ./...
	@echo "=== kcp-gateway 테스트 ==="
	cd kcp-gateway && go test ./...

# 린트
lint:
	cd kcp-cli && golangci-lint run ./...
	cd kcp-gateway && golangci-lint run ./...

# PostgreSQL 컨테이너 실행
db-up:
	bash db/nerdctl-postgres.sh

# PostgreSQL 컨테이너 중지
db-down:
	nerdctl rm -f kcp-postgres

# 클린
clean:
	rm -rf bin/
	cd kcp-cli && go clean
	cd kcp-gateway && go clean

# Gateway 실행
run-gateway:
	cd kcp-gateway && go run ./cmd/server/

# WebUI 개발 서버
run-webui:
	cd kcp-webui && npm run dev
