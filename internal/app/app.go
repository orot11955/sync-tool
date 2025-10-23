package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"sync-tool/internal/config"
	"sync-tool/internal/logger"
	"sync-tool/internal/sync"
	"sync-tool/internal/ui"
)

// ShowStatus 동기화 상태 표시
func ShowStatus(cfg *config.Config) error {
	logger.Info("동기화 상태 확인 시작")

	fmt.Println("=== Sync Tool 상태 ===")
	fmt.Printf("서버: %s@%s:%d\n", cfg.Server.User, cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("설정된 프로필: %d개\n", len(cfg.Profiles))
	fmt.Println()

	// 각 프로필별 상태 확인
	for name, profile := range cfg.Profiles {
		fmt.Printf("프로필: %s\n", name)
		fmt.Printf("  설명: %s\n", profile.Description)
		fmt.Printf("  서버 경로: %s\n", profile.ServerPath)
		fmt.Printf("  로컬 경로: %s\n", profile.LocalPath)

		// 로컬 경로 존재 여부 확인
		if _, err := os.Stat(profile.LocalPath); os.IsNotExist(err) {
			fmt.Printf("  상태: ❌ 로컬 경로 없음\n")
		} else {
			fmt.Printf("  상태: ✅ 로컬 경로 존재\n")
		}
		fmt.Println()
	}

	return nil
}

// ShowProfiles 사용 가능한 프로필 목록 표시
func ShowProfiles(cfg *config.Config) error {
	fmt.Println("=== 사용 가능한 프로필 ===")

	for name, profile := range cfg.Profiles {
		fmt.Printf("• %s\n", name)
		fmt.Printf("  %s\n", profile.Description)
		fmt.Printf("  서버: %s\n", profile.ServerPath)
		fmt.Printf("  로컬: %s\n", profile.LocalPath)
		fmt.Println()
	}

	return nil
}

// Sync 파일 동기화 실행
func Sync(cfg *config.Config, profileName string, dryRun bool, autoConfirm bool) error {
	// 로거 초기화
	if err := logger.Init(&cfg.Logging); err != nil {
		return fmt.Errorf("로거 초기화 실패: %w", err)
	}

	logger.Info("동기화 시작")

	// 프로필 선택
	var selectedProfile *config.SyncProfile
	if profileName != "" {
		profile, exists := cfg.Profiles[profileName]
		if !exists {
			return fmt.Errorf("프로필을 찾을 수 없습니다: %s", profileName)
		}
		selectedProfile = &profile
	} else {
		// 대화형 프로필 선택
		profile, err := selectProfileInteractively(cfg)
		if err != nil {
			return fmt.Errorf("프로필 선택 실패: %w", err)
		}
		selectedProfile = profile
	}

	logger.Infof("선택된 프로필: %s", selectedProfile.Name)

	// 동기화 엔진 생성
	syncEngine := sync.NewSyncEngine(cfg)

	// 프로필 유효성 검사
	if err := syncEngine.ValidateProfile(selectedProfile); err != nil {
		return fmt.Errorf("프로필 유효성 검사 실패: %w", err)
	}

	// 드라이런 실행
	logger.Info("변경사항 확인 중...")
	changes, err := syncEngine.DryRun(selectedProfile)
	if err != nil {
		return fmt.Errorf("드라이런 실행 실패: %w", err)
	}

	// 변경사항 표시
	showChanges(changes)

	// 변경사항이 없는 경우
	if !changes.HasChanges && !changes.HasDeletions {
		fmt.Println("✅ 동기화할 변경사항이 없습니다.")
		return nil
	}

	// 드라이런 모드인 경우 여기서 종료
	if dryRun {
		fmt.Println("드라이런 모드로 실행되었습니다. 실제 동기화는 수행되지 않았습니다.")
		return nil
	}

	// 사용자 확인
	if !autoConfirm && cfg.UI.ConfirmActions {
		if !confirmSync(changes) {
			fmt.Println("동기화가 취소되었습니다.")
			return nil
		}
	}

	// 실제 동기화 실행
	logger.Info("동기화 실행 중...")
	if err := syncEngine.Sync(selectedProfile, changes); err != nil {
		return fmt.Errorf("동기화 실행 실패: %w", err)
	}

	fmt.Println("✅ 동기화가 완료되었습니다.")
	return nil
}

// selectProfileInteractively 대화형 프로필 선택
func selectProfileInteractively(cfg *config.Config) (*config.SyncProfile, error) {
	fmt.Println("동기화할 프로필을 선택하세요:")
	fmt.Println()

	profiles := make([]config.SyncProfile, 0, len(cfg.Profiles))
	profileNames := make([]string, 0, len(cfg.Profiles))

	i := 1
	for name, profile := range cfg.Profiles {
		fmt.Printf("%d) %s - %s\n", i, name, profile.Description)
		fmt.Printf("   서버: %s\n", profile.ServerPath)
		fmt.Printf("   로컬: %s\n", profile.LocalPath)
		fmt.Println()

		profiles = append(profiles, profile)
		profileNames = append(profileNames, name)
		i++
	}

	fmt.Print("번호 선택: ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("입력 읽기 실패: %w", err)
	}

	input = strings.TrimSpace(input)
	var selectedIndex int
	if _, err := fmt.Sscanf(input, "%d", &selectedIndex); err != nil {
		return nil, fmt.Errorf("잘못된 입력: %s", input)
	}

	if selectedIndex < 1 || selectedIndex > len(profiles) {
		return nil, fmt.Errorf("잘못된 선택: %d", selectedIndex)
	}

	return &profiles[selectedIndex-1], nil
}

// showChanges 변경사항 표시
func showChanges(changes *sync.SyncResult) {
	fmt.Println()
	fmt.Println("=== 변경사항 요약 ===")
	fmt.Printf("복사할 파일: %d개\n", len(changes.Changes))
	fmt.Printf("삭제할 파일: %d개\n", len(changes.Deletions))
	fmt.Println()

	// 복사할 파일 목록
	if len(changes.Changes) > 0 {
		fmt.Println("복사할 파일 목록:")
		fmt.Println("────────────────────")
		for _, change := range changes.Changes {
			icon := getChangeIcon(change.Type)
			fmt.Printf("%s %s\n", icon, change.Path)
		}
		fmt.Println()
	}

	// 삭제할 파일 목록
	if len(changes.Deletions) > 0 {
		fmt.Println("삭제할 파일 목록:")
		fmt.Println("────────────────────")
		for _, deletion := range changes.Deletions {
			fmt.Printf("🗑️  %s\n", deletion)
		}
		fmt.Println()
	}
}

// confirmSync 동기화 확인
func confirmSync(changes *sync.SyncResult) bool {
	fmt.Print("위 파일들을 동기화하시겠습니까? (y/n): ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		logger.Errorf("입력 읽기 실패: %v", err)
		return false
	}

	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes"
}

// getChangeIcon 변경 타입에 따른 아이콘 반환
func getChangeIcon(changeType sync.ChangeType) string {
	switch changeType {
	case sync.ChangeTypeNew:
		return "📄"
	case sync.ChangeTypeModified:
		return "📝"
	case sync.ChangeTypeDeleted:
		return "🗑️"
	default:
		return "📁"
	}
}

// ShowTUI TUI 모드로 실행
func ShowTUI(cfg *config.Config, profileName string) error {
	// 로거 초기화
	if err := logger.Init(&cfg.Logging); err != nil {
		return fmt.Errorf("로거 초기화 실패: %w", err)
	}

	// TUI 표시 (프로필 선택은 TUI에서 처리)
	profiles := make(map[string]string)
	for name, profile := range cfg.Profiles {
		profiles[name] = profile.Description
	}

	// 임시 변경사항 (실제로는 TUI에서 프로필 선택 후 동기화)
	changes := &sync.SyncResult{
		Changes:   []sync.FileChange{},
		Deletions: []string{},
	}

	return ui.ShowSyncUI(profiles, changes)
}
