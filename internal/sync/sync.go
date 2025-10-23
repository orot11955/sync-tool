package sync

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"sync-tool/internal/config"
	"sync-tool/internal/logger"
)

// ChangeType 파일 변경 타입
type ChangeType string

const (
	ChangeTypeNew       ChangeType = "new"
	ChangeTypeModified  ChangeType = "modified"
	ChangeTypeDeleted   ChangeType = "deleted"
	ChangeTypeUnchanged ChangeType = "unchanged"
)

// FileChange 파일 변경 정보
type FileChange struct {
	Type     ChangeType
	Path     string
	Size     string
	Checksum string
}

// SyncResult 동기화 결과
type SyncResult struct {
	Changes      []FileChange
	Deletions    []string
	Error        error
	HasChanges   bool
	HasDeletions bool
}

// SyncEngine 동기화 엔진
type SyncEngine struct {
	config *config.Config
}

// NewSyncEngine 새로운 동기화 엔진 생성
func NewSyncEngine(cfg *config.Config) *SyncEngine {
	return &SyncEngine{
		config: cfg,
	}
}

// DryRun 실제 동기화 없이 변경사항만 확인
func (s *SyncEngine) DryRun(profile *config.SyncProfile) (*SyncResult, error) {
	logger.Debugf("드라이런 시작: 프로필=%s, 서버경로=%s, 로컬경로=%s",
		profile.Name, profile.ServerPath, profile.LocalPath)

	// rsync 명령어 구성
	cmd := s.buildRsyncCommand(profile, true)

	logger.Debugf("실행할 rsync 명령어: %s", strings.Join(cmd.Args, " "))

	// 명령어 실행 (드라이런은 CombinedOutput 사용)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Errorf("rsync 드라이런 실행 실패: %v", err)
		logger.Errorf("rsync 출력: %s", string(output))
		return nil, fmt.Errorf("rsync 드라이런 실행 실패: %w", err)
	}

	// 디버깅: 원본 rsync 출력 확인
	logger.Debugf("원본 rsync 출력: %s", string(output))

	// 출력 파싱
	result := s.parseRsyncOutput(string(output))

	logger.Infof("드라이런 완료: 변경파일=%d개, 삭제파일=%d개",
		len(result.Changes), len(result.Deletions))

	return result, nil
}

// Sync 실제 동기화 실행
func (s *SyncEngine) Sync(profile *config.SyncProfile, changes *SyncResult) error {
	logger.Infof("동기화 시작: 프로필=%s", profile.Name)

	// 복사할 파일이 있는 경우
	if len(changes.Changes) > 0 {
		if err := s.syncFiles(profile, changes.Changes); err != nil {
			return fmt.Errorf("파일 동기화 실패: %w", err)
		}
	}

	// 삭제할 파일이 있는 경우
	if len(changes.Deletions) > 0 {
		if err := s.deleteFiles(profile, changes.Deletions); err != nil {
			return fmt.Errorf("파일 삭제 실패: %w", err)
		}
	}

	logger.Info("동기화 완료")
	return nil
}

// buildRsyncCommand rsync 명령어 구성
func (s *SyncEngine) buildRsyncCommand(profile *config.SyncProfile, dryRun bool) *exec.Cmd {
	args := []string{}

	// 기본 옵션
	options := profile.GetSyncOptions(s.config.Sync.Options)
	args = append(args, options...)

	// 드라이런 옵션
	if dryRun {
		args = append(args, "--dry-run")
	}

	// 권한 관련 옵션
	args = append(args, "--no-perms", "--no-owner", "--no-group")

	// 안정성을 위한 추가 옵션
	args = append(args, "--partial", "--partial-dir=.rsync-partial")
	args = append(args, "--timeout=300") // 5분 타임아웃

	// 삭제 옵션 (드라이런에서도 삭제 확인)
	args = append(args, "--delete")

	// SSH 옵션
	sshArgs := fmt.Sprintf("ssh -p %d", s.config.Server.Port)
	if s.config.Server.KeyPath != "" {
		sshArgs += fmt.Sprintf(" -i %s", s.config.Server.KeyPath)
	}
	args = append(args, "-e", sshArgs)

	// 제외/포함 패턴
	excludes := profile.GetExcludes(s.config.Sync.DefaultExcludes)
	logger.Debugf("기본 제외 패턴: %v", s.config.Sync.DefaultExcludes)
	logger.Debugf("프로필 제외 패턴: %v", profile.Excludes)
	logger.Debugf("최종 제외 패턴: %v", excludes)
	for _, exclude := range excludes {
		args = append(args, "--exclude", exclude)
	}

	for _, include := range profile.Includes {
		args = append(args, "--include", include)
	}

	// 소스와 대상
	source := fmt.Sprintf("%s@%s:%s/", s.config.Server.User, s.config.Server.Host, profile.ServerPath)
	target := fmt.Sprintf("%s/", profile.LocalPath)
	args = append(args, source, target)

	// 디버그: 생성된 rsync 명령어 로깅
	logger.Debugf("생성된 rsync 명령어: rsync %s", strings.Join(args, " "))

	return exec.Command("rsync", args...)
}

// parseRsyncOutput rsync 출력 파싱
func (s *SyncEngine) parseRsyncOutput(output string) *SyncResult {
	result := &SyncResult{
		Changes:   []FileChange{},
		Deletions: []string{},
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// rsync 출력 형식 파싱
		// 예: >f+++++++++ file.txt 또는 *deleting file.txt

		// 삭제 파일 처리
		if strings.HasPrefix(line, "*deleting") {
			path := strings.TrimSpace(line[9:]) // "*deleting " 제거
			if path != "" {
				result.Deletions = append(result.Deletions, path)
			}
			continue
		}

		// 일반 파일 변경 처리 - rsync 출력 형식: [타입][권한][크기][날짜] 파일명
		// 예: >f+++++++ test_text, >fcsT.... package/service/file.tar.gz
		if len(line) >= 12 {
			// rsync 통계 라인 제외 (sent, received, total size 등)
			if strings.Contains(line, "sent ") ||
				strings.Contains(line, "received ") ||
				strings.Contains(line, "total size") ||
				strings.Contains(line, "bytes/sec") ||
				strings.Contains(line, "speedup") ||
				strings.Contains(line, "Transfer starting") {
				continue
			}

			// 공백으로 분리하여 마지막 부분이 파일 경로
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				changeType := parts[0]
				path := parts[len(parts)-1] // 마지막 부분이 파일 경로

				// 유효한 변경 타입인지 확인
				if (strings.HasPrefix(changeType, "<") || strings.HasPrefix(changeType, ">") ||
					strings.HasPrefix(changeType, "c") || strings.HasPrefix(changeType, "h") ||
					strings.HasPrefix(changeType, "d") || strings.HasPrefix(changeType, "s")) &&
					path != "" &&
					!strings.Contains(path, "bytes") && !strings.Contains(path, "sec") &&
					!strings.Contains(path, "sent") && !strings.Contains(path, "received") {

					change := FileChange{
						Path: path,
					}

					// 변경 타입 결정
					if strings.Contains(changeType, "+++++++") {
						change.Type = ChangeTypeNew
					} else if strings.Contains(changeType, "csT") || strings.Contains(changeType, "..T") {
						change.Type = ChangeTypeModified
					} else {
						change.Type = ChangeTypeModified
					}

					result.Changes = append(result.Changes, change)
				}
			}
		}
	}

	result.HasChanges = len(result.Changes) > 0
	result.HasDeletions = len(result.Deletions) > 0

	return result
}

// syncFiles 파일 동기화
func (s *SyncEngine) syncFiles(profile *config.SyncProfile, changes []FileChange) error {
	logger.Infof("파일 복사 시작: %d개 파일", len(changes))

	// 변경된 파일 목록을 임시 파일로 저장
	tmpFile, err := os.CreateTemp("", "sync-files-*.txt")
	if err != nil {
		return fmt.Errorf("임시 파일 생성 실패: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	for _, change := range changes {
		if change.Type == ChangeTypeNew || change.Type == ChangeTypeModified {
			fmt.Fprintln(tmpFile, change.Path)
		}
	}
	tmpFile.Close()

	// rsync 명령어 구성 (파일 목록 사용)
	cmd := s.buildRsyncCommandWithFileList(profile, tmpFile.Name())

	logger.Debugf("파일 복사 명령어: %s", strings.Join(cmd.Args, " "))

	// 진행률 표시기 생성
	progress := NewSimpleProgress(len(changes))

	// 실시간 출력을 위해 stdout/stderr을 터미널에 연결
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 시작 메시지
	fmt.Printf("\n🔄 파일 동기화 진행 중...\n")
	fmt.Printf("📁 대상: %s\n", profile.LocalPath)
	fmt.Printf("📊 총 파일: %d개\n\n", len(changes))

	// 진행률 표시 시작
	progress.Update(0, "시작...")

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("파일 복사 시작 실패: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		// rsync exit status 코드에 따른 에러 메시지
		if exitError, ok := err.(*exec.ExitError); ok {
			switch exitError.ExitCode() {
			case 23:
				return fmt.Errorf("rsync 부분 실패 (일부 파일 전송 실패)")
			case 24:
				return fmt.Errorf("rsync 일시적 실패 (재시도 필요)")
			default:
				return fmt.Errorf("rsync 실행 실패 (exit code %d)", exitError.ExitCode())
			}
		}
		return fmt.Errorf("파일 복사 실행 실패: %w", err)
	}

	// 진행률 완료 표시
	progress.Complete()

	fmt.Printf("📁 대상 경로: %s\n", profile.LocalPath)
	logger.Infof("파일 복사 완료: %s", profile.LocalPath)

	return nil
}

// deleteFiles 파일 삭제
func (s *SyncEngine) deleteFiles(profile *config.SyncProfile, deletions []string) error {
	logger.Infof("로컬 파일 삭제 시작: %d개 파일", len(deletions))

	for _, filePath := range deletions {
		fullPath := filepath.Join(profile.LocalPath, filePath)

		// 파일 존재 확인
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			logger.Warnf("삭제할 파일이 존재하지 않음: %s", fullPath)
			continue
		}

		// 파일/디렉토리 삭제
		if err := os.RemoveAll(fullPath); err != nil {
			logger.Errorf("파일 삭제 실패: %s, 오류: %v", fullPath, err)
			continue
		}

		logger.Infof("파일 삭제됨: %s", fullPath)
	}

	logger.Info("로컬 파일 삭제 완료")
	return nil
}

// buildRsyncCommandWithFileList 파일 목록을 사용한 rsync 명령어 구성
func (s *SyncEngine) buildRsyncCommandWithFileList(profile *config.SyncProfile, fileListPath string) *exec.Cmd {
	args := []string{}

	// 기본 옵션
	options := profile.GetSyncOptions(s.config.Sync.Options)
	args = append(args, options...)

	// 파일 목록 옵션
	args = append(args, "--files-from", fileListPath)

	// 권한 관련 옵션
	args = append(args, "--no-perms", "--no-owner", "--no-group")

	// 안정성을 위한 추가 옵션
	args = append(args, "--partial", "--partial-dir=.rsync-partial")
	args = append(args, "--timeout=300") // 5분 타임아웃

	// SSH 옵션
	sshArgs := fmt.Sprintf("ssh -p %d", s.config.Server.Port)
	if s.config.Server.KeyPath != "" {
		sshArgs += fmt.Sprintf(" -i %s", s.config.Server.KeyPath)
	}
	args = append(args, "-e", sshArgs)

	// 소스와 대상
	source := fmt.Sprintf("%s@%s:%s/", s.config.Server.User, s.config.Server.Host, profile.ServerPath)
	target := fmt.Sprintf("%s/", profile.LocalPath)
	args = append(args, source, target)

	return exec.Command("rsync", args...)
}

// ValidateProfile 프로필 유효성 검사
func (s *SyncEngine) ValidateProfile(profile *config.SyncProfile) error {
	// 로컬 경로 확인
	if _, err := os.Stat(profile.LocalPath); os.IsNotExist(err) {
		return fmt.Errorf("로컬 경로가 존재하지 않습니다: %s", profile.LocalPath)
	}

	// 서버 정보 확인
	if s.config.Server.Host == "" {
		return fmt.Errorf("서버 호스트가 설정되지 않았습니다")
	}

	if s.config.Server.User == "" {
		return fmt.Errorf("서버 사용자가 설정되지 않았습니다")
	}

	return nil
}
