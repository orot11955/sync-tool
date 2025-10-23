#!/bin/bash

# Sync Tool 빌드 스크립트

set -e

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 로그 함수들
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 변수 설정
BINARY_NAME="sync-tool"
BUILD_DIR="build"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION=$(go version | awk '{print $3}')

# LDFLAGS
LDFLAGS="-ldflags '-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GoVersion=${GO_VERSION}'"

# 플랫폼별 빌드 설정
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
    "windows/arm64"
    "darwin/amd64"
    "darwin/arm64"
)

# 함수들
check_dependencies() {
    log_info "의존성 확인 중..."
    
    if ! command -v go &> /dev/null; then
        log_error "Go가 설치되지 않았습니다."
        exit 1
    fi
    
    log_success "Go 버전: $(go version)"
}

setup_build_dir() {
    log_info "빌드 디렉토리 설정 중..."
    mkdir -p ${BUILD_DIR}
    log_success "빌드 디렉토리 생성: ${BUILD_DIR}"
}

download_dependencies() {
    log_info "의존성 다운로드 중..."
    go mod download
    go mod tidy
    log_success "의존성 다운로드 완료"
}

build_current_platform() {
    log_info "현재 플랫폼용 빌드 중..."
    
    local output_file="${BUILD_DIR}/${BINARY_NAME}"
    if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
        output_file="${output_file}.exe"
    fi
    
    go build ${LDFLAGS} -o ${output_file} .
    
    if [ $? -eq 0 ]; then
        log_success "빌드 완료: ${output_file}"
    else
        log_error "빌드 실패"
        exit 1
    fi
}

build_all_platforms() {
    log_info "모든 플랫폼용 빌드 중..."
    
    for platform in "${PLATFORMS[@]}"; do
        IFS='/' read -r os arch <<< "$platform"
        
        log_info "${os}/${arch} 빌드 중..."
        
        local output_name="${BINARY_NAME}-${os}-${arch}"
        if [[ "$os" == "windows" ]]; then
            output_name="${output_name}.exe"
        fi
        
        local output_file="${BUILD_DIR}/${output_name}"
        
        GOOS=${os} GOARCH=${arch} go build ${LDFLAGS} -o ${output_file} .
        
        if [ $? -eq 0 ]; then
            log_success "${os}/${arch} 빌드 완료: ${output_file}"
        else
            log_error "${os}/${arch} 빌드 실패"
            exit 1
        fi
    done
    
    log_success "모든 플랫폼 빌드 완료"
}

create_checksums() {
    log_info "체크섬 파일 생성 중..."
    
    cd ${BUILD_DIR}
    if command -v shasum &> /dev/null; then
        shasum -a 256 * > checksums.txt
    elif command -v sha256sum &> /dev/null; then
        sha256sum * > checksums.txt
    else
        log_warning "체크섬 생성 도구를 찾을 수 없습니다."
        return
    fi
    
    cd ..
    log_success "체크섬 파일 생성 완료: ${BUILD_DIR}/checksums.txt"
}

show_build_info() {
    log_info "빌드 정보:"
    echo "  버전: ${VERSION}"
    echo "  빌드 시간: ${BUILD_TIME}"
    echo "  Go 버전: ${GO_VERSION}"
    echo "  플랫폼: $(uname -s)/$(uname -m)"
}

# 메인 함수
main() {
    log_info "Sync Tool 빌드 시작"
    
    # 인수 처리
    case "${1:-build}" in
        "build")
            check_dependencies
            setup_build_dir
            download_dependencies
            build_current_platform
            show_build_info
            ;;
        "build-all")
            check_dependencies
            setup_build_dir
            download_dependencies
            build_all_platforms
            create_checksums
            show_build_info
            ;;
        "clean")
            log_info "빌드 파일 정리 중..."
            rm -rf ${BUILD_DIR}
            log_success "정리 완료"
            ;;
        "info")
            show_build_info
            ;;
        *)
            echo "사용법: $0 {build|build-all|clean|info}"
            echo ""
            echo "명령어:"
            echo "  build     - 현재 플랫폼용 빌드"
            echo "  build-all - 모든 플랫폼용 빌드"
            echo "  clean     - 빌드 파일 정리"
            echo "  info      - 빌드 정보 표시"
            exit 1
            ;;
    esac
    
    log_success "빌드 스크립트 완료"
}

# 스크립트 실행
main "$@"
