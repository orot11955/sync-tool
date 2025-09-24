# Sync Tool

Git과 유사한 동작 방식을 가진 서버-USB 파일 동기화 도구입니다.

## 주요 기능

- 🔄 **자동 동기화**: 서버와 로컬 간 파일 변경사항 자동 감지
- 📁 **프로필 시스템**: 다양한 동기화 설정을 프로필로 관리
- 🔍 **Dry-run 모드**: 실제 동기화 전 변경사항 미리보기
- 🎨 **TUI 인터페이스**: 직관적인 터미널 사용자 인터페이스
- 📝 **상세 로깅**: 디버깅을 위한 이해하기 쉬운 로그
- 🌍 **크로스 플랫폼**: Windows, macOS, Linux 지원
- ⚙️ **유연한 설정**: YAML 기반 설정 파일

## 설치

### 바이너리 다운로드

[Releases](https://github.com/your-repo/sync-tool/releases) 페이지에서 플랫폼에 맞는 바이너리를 다운로드하세요.

### 소스에서 빌드

```bash
git clone https://github.com/your-repo/sync-tool.git
cd sync-tool
make build
```

## 사용법

### 초기 설정

```bash
# 설정 파일 초기화
./sync-tool init

# 사용 가능한 프로필 확인
./sync-tool profiles

# 동기화 상태 확인
./sync-tool status
```

### 동기화 실행

```bash
# 대화형 프로필 선택으로 동기화
./sync-tool sync

# 특정 프로필로 동기화
./sync-tool sync aunes_ins

# 드라이런 모드 (변경사항만 확인)
./sync-tool sync aunes_ins --dry-run

# 확인 없이 자동 실행
./sync-tool sync aunes_ins --yes
```

### TUI 모드

```bash
# TUI 인터페이스로 실행
./sync-tool sync --tui
```

## 설정 파일 (config.yaml)

```yaml
# 서버 설정
server:
  host: "10.10.30.237"
  user: "root"
  port: 22
  key_path: ""  # SSH 키 경로

# 동기화 기본 설정
sync:
  options:
    - "-r"           # recursive
    - "-z"           # compress
    - "-c"           # checksum
    - "--itemize-changes"
  
  default_excludes:
    - ".DS_Store"
    - "._*"
    - ".Spotlight-V100"
    - ".Trashes"
    - ".TemporaryItems"
    - "System Volume Information"
    - ".fseventsd"
    - ".Trash-1000"

# 동기화 프로필들
profiles:
  aunes_ins:
    name: "AUNES_INS"
    description: "AUNES INS 폴더 동기화"
    server_path: "/stor2/USB_SYNC/AUNES_INS"
    local_path: "/Volumes/AUNES_INS"
    
  ventoy:
    name: "Ventoy (KICKSTART, config)"
    description: "Ventoy 폴더 동기화 (ISO 파일 제외)"
    server_path: "/stor2/USB_SYNC/Ventoy"
    local_path: "/Volumes/Ventoy"
    excludes:
      - "_iso/*.iso"
      
  iso_only:
    name: "ISO Files Only"
    description: "ISO 파일만 동기화"
    server_path: "/stor2/USB_SYNC/Ventoy"
    local_path: "/Volumes/Ventoy"
    includes:
      - "*.iso"
    excludes: ["*"]

# 로깅 설정
logging:
  level: "info"  # debug, info, warn, error
  format: "text"  # text, json
  file: ""  # 로그 파일 경로 (빈 문자열이면 콘솔만)

# UI 설정
ui:
  theme: "default"  # default, dark, light
  show_progress: true
  confirm_actions: true
```

## 프로필 설정

각 프로필은 다음과 같은 속성을 가집니다:

- `name`: 프로필 이름
- `description`: 프로필 설명
- `server_path`: 서버의 동기화 대상 경로
- `local_path`: 로컬의 동기화 대상 경로
- `options`: rsync 옵션 (선택사항, 기본값 사용 시 생략)
- `includes`: 포함할 파일 패턴 (선택사항)
- `excludes`: 제외할 파일 패턴 (선택사항)

## 개발

### 빌드

```bash
# 현재 플랫폼용 빌드
make build

# 모든 플랫폼용 빌드
make build-all

# 개발 모드로 실행
make run ARGS="--help"
```

### 테스트

```bash
# 테스트 실행
make test

# 코드 커버리지
make coverage

# 벤치마크
make bench
```

### 코드 품질

```bash
# 린팅
make lint

# 코드 포맷팅
make fmt

# 보안 검사
make security
```

## 기존 sync.sh 스크립트와의 차이점

| 기능 | sync.sh | sync-tool |
|------|---------|-----------|
| 설정 관리 | 하드코딩 | YAML 설정 파일 |
| 프로필 관리 | case문으로 분기 | 구조화된 프로필 시스템 |
| UI | 텍스트 기반 | TUI 인터페이스 |
| 로깅 | echo 출력 | 구조화된 로깅 |
| 플랫폼 지원 | macOS 전용 | 크로스 플랫폼 |
| 확장성 | 제한적 | 모듈화된 구조 |

## 라이선스

MIT License

## 기여

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request
