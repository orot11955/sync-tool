@echo off
REM Sync Tool Windows 빌드 스크립트

setlocal enabledelayedexpansion

REM 색상 정의 (Windows 10 이상)
for /f %%i in ('echo prompt $E ^| cmd') do set "ESC=%%i"
set "RED=%ESC%[31m"
set "GREEN=%ESC%[32m"
set "YELLOW=%ESC%[33m"
set "BLUE=%ESC%[34m"
set "NC=%ESC%[0m"

REM 변수 설정
set "BINARY_NAME=sync-tool"
set "BUILD_DIR=build"
set "VERSION=dev"
set "BUILD_TIME=%DATE% %TIME%"
set "GO_VERSION=unknown"

REM Git 버전 확인
git describe --tags --always --dirty >nul 2>&1
if %errorlevel% equ 0 (
    for /f "delims=" %%i in ('git describe --tags --always --dirty') do set "VERSION=%%i"
)

REM Go 버전 확인
go version >nul 2>&1
if %errorlevel% equ 0 (
    for /f "tokens=3" %%i in ('go version') do set "GO_VERSION=%%i"
)

REM LDFLAGS
set "LDFLAGS=-ldflags -X main.Version=%VERSION% -X main.BuildTime=%BUILD_TIME% -X main.GoVersion=%GO_VERSION%"

REM 플랫폼별 빌드 설정
set "PLATFORMS=linux/amd64 linux/arm64 windows/amd64 windows/arm64 darwin/amd64 darwin/arm64"

REM 함수들
:log_info
echo %BLUE%[INFO]%NC% %~1
goto :eof

:log_success
echo %GREEN%[SUCCESS]%NC% %~1
goto :eof

:log_warning
echo %YELLOW%[WARNING]%NC% %~1
goto :eof

:log_error
echo %RED%[ERROR]%NC% %~1
goto :eof

:check_dependencies
call :log_info "의존성 확인 중..."

go version >nul 2>&1
if %errorlevel% neq 0 (
    call :log_error "Go가 설치되지 않았습니다."
    exit /b 1
)

for /f "tokens=*" %%i in ('go version') do call :log_success "Go 버전: %%i"
goto :eof

:setup_build_dir
call :log_info "빌드 디렉토리 설정 중..."
if not exist "%BUILD_DIR%" mkdir "%BUILD_DIR%"
call :log_success "빌드 디렉토리 생성: %BUILD_DIR%"
goto :eof

:download_dependencies
call :log_info "의존성 다운로드 중..."
go mod download
go mod tidy
if %errorlevel% equ 0 (
    call :log_success "의존성 다운로드 완료"
) else (
    call :log_error "의존성 다운로드 실패"
    exit /b 1
)
goto :eof

:build_current_platform
call :log_info "현재 플랫폼용 빌드 중..."

set "output_file=%BUILD_DIR%\%BINARY_NAME%.exe"
go build %LDFLAGS% -o "%output_file%" .

if %errorlevel% equ 0 (
    call :log_success "빌드 완료: %output_file%"
) else (
    call :log_error "빌드 실패"
    exit /b 1
)
goto :eof

:build_all_platforms
call :log_info "모든 플랫폼용 빌드 중..."

for %%p in (%PLATFORMS%) do (
    for /f "tokens=1,2 delims=/" %%a in ("%%p") do (
        call :log_info "%%a/%%b 빌드 중..."
        
        set "output_name=%BINARY_NAME%-%%a-%%b.exe"
        set "output_file=%BUILD_DIR%\!output_name!"
        
        set "GOOS=%%a"
        set "GOARCH=%%b"
        
        go build %LDFLAGS% -o "!output_file!" .
        
        if !errorlevel! equ 0 (
            call :log_success "%%a/%%b 빌드 완료: !output_file!"
        ) else (
            call :log_error "%%a/%%b 빌드 실패"
            exit /b 1
        )
    )
)

call :log_success "모든 플랫폼 빌드 완료"
goto :eof

:create_checksums
call :log_info "체크섬 파일 생성 중..."

cd "%BUILD_DIR%"
certutil -hashfile * SHA256 > checksums.txt 2>nul
if %errorlevel% equ 0 (
    call :log_success "체크섬 파일 생성 완료: %BUILD_DIR%\checksums.txt"
) else (
    call :log_warning "체크섬 생성 실패"
)
cd ..
goto :eof

:show_build_info
call :log_info "빌드 정보:"
echo   버전: %VERSION%
echo   빌드 시간: %BUILD_TIME%
echo   Go 버전: %GO_VERSION%
echo   플랫폼: %OS%
goto :eof

:clean
call :log_info "빌드 파일 정리 중..."
if exist "%BUILD_DIR%" rmdir /s /q "%BUILD_DIR%"
call :log_success "정리 완료"
goto :eof

REM 메인 로직
call :log_info "Sync Tool 빌드 시작"

if "%1"=="" set "1=build"

if "%1"=="build" (
    call :check_dependencies
    call :setup_build_dir
    call :download_dependencies
    call :build_current_platform
    call :show_build_info
) else if "%1"=="build-all" (
    call :check_dependencies
    call :setup_build_dir
    call :download_dependencies
    call :build_all_platforms
    call :create_checksums
    call :show_build_info
) else if "%1"=="clean" (
    call :clean
) else if "%1"=="info" (
    call :show_build_info
) else (
    echo 사용법: %0 {build^|build-all^|clean^|info}
    echo.
    echo 명령어:
    echo   build     - 현재 플랫폼용 빌드
    echo   build-all - 모든 플랫폼용 빌드
    echo   clean     - 빌드 파일 정리
    echo   info      - 빌드 정보 표시
    exit /b 1
)

call :log_success "빌드 스크립트 완료"
endlocal
