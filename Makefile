# Sync Tool Makefile

# 변수 설정
BINARY_NAME=sync-tool
BUILD_DIR=build
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION=$(shell go version | awk '{print $$3}')

# LDFLAGS
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GoVersion=$(GO_VERSION)"

# 기본 타겟
.PHONY: all
all: clean build

# 의존성 설치
.PHONY: deps
deps:
	@echo "의존성 설치 중..."
	go mod download
	go mod tidy

# 빌드 (현재 플랫폼)
.PHONY: build
build: deps
	@echo "빌드 중... ($(GO_VERSION))"
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

# 크로스 플랫폼 빌드
.PHONY: build-all
build-all: deps
	@echo "크로스 플랫폼 빌드 중..."
	@mkdir -p $(BUILD_DIR)
	
	# Linux AMD64
	@echo "Linux AMD64 빌드 중..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	
	# Linux ARM64
	@echo "Linux ARM64 빌드 중..."
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	
	# Windows AMD64
	@echo "Windows AMD64 빌드 중..."
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	
	# Windows ARM64
	@echo "Windows ARM64 빌드 중..."
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-arm64.exe .
	
	# macOS AMD64
	@echo "macOS AMD64 빌드 중..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	
	# macOS ARM64
	@echo "macOS ARM64 빌드 중..."
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	
	@echo "모든 빌드 완료!"

# 설치 (현재 플랫폼)
.PHONY: install
install: build
	@echo "설치 중..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "설치 완료: /usr/local/bin/$(BINARY_NAME)"

# 개발용 실행
.PHONY: run
run: deps
	@echo "개발 모드로 실행 중..."
	go run . $(ARGS)

# 테스트
.PHONY: test
test:
	@echo "테스트 실행 중..."
	go test -v ./...

# 벤치마크
.PHONY: bench
bench:
	@echo "벤치마크 실행 중..."
	go test -bench=. -benchmem ./...

# 코드 커버리지
.PHONY: coverage
coverage:
	@echo "코드 커버리지 측정 중..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "커버리지 리포트: coverage.html"

# 린팅
.PHONY: lint
lint:
	@echo "린팅 실행 중..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint가 설치되지 않았습니다. 설치 중..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi

# 포맷팅
.PHONY: fmt
fmt:
	@echo "코드 포맷팅 중..."
	go fmt ./...
	goimports -w .

# 보안 검사
.PHONY: security
security:
	@echo "보안 검사 실행 중..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec가 설치되지 않았습니다. 설치 중..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
		gosec ./...; \
	fi

# 정리
.PHONY: clean
clean:
	@echo "빌드 파일 정리 중..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# 도움말
.PHONY: help
help:
	@echo "Sync Tool Makefile"
	@echo ""
	@echo "사용 가능한 타겟:"
	@echo "  all         - 정리 후 빌드"
	@echo "  deps        - 의존성 설치"
	@echo "  build       - 현재 플랫폼용 빌드"
	@echo "  build-all   - 모든 플랫폼용 빌드"
	@echo "  install     - 시스템에 설치"
	@echo "  run         - 개발 모드로 실행 (ARGS=인수 전달 가능)"
	@echo "  test        - 테스트 실행"
	@echo "  bench       - 벤치마크 실행"
	@echo "  coverage    - 코드 커버리지 측정"
	@echo "  lint        - 린팅 실행"
	@echo "  fmt         - 코드 포맷팅"
	@echo "  security    - 보안 검사"
	@echo "  clean       - 빌드 파일 정리"
	@echo "  help        - 이 도움말 표시"
	@echo ""
	@echo "예시:"
	@echo "  make run ARGS=\"--help\""
	@echo "  make run ARGS=\"sync aunes_ins --dry-run\""
