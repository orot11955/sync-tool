package ui

import (
	"fmt"
	"strings"

	"sync-tool/internal/sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// UIState UI 상태
type UIState int

const (
	StateSelectProfile UIState = iota
	StateConfirmSync
	StateSyncProgress
	StateCompleted
)

// SyncModel TUI 모델
type SyncModel struct {
	state           UIState
	profiles        map[string]string
	selectedProfile string
	changes         *sync.SyncResult
	confirmSync     bool
	confirmDelete   bool
	err             error
	width           int
	height          int
}

// Init 초기화
func (m SyncModel) Init() tea.Cmd {
	return nil
}

// Update 업데이트
func (m SyncModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case StateSelectProfile:
			return m.updateProfileSelection(msg)
		case StateConfirmSync:
			return m.updateSyncConfirmation(msg)
		case StateCompleted:
			if msg.String() == "q" {
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

// View 뷰 렌더링
func (m SyncModel) View() string {
	switch m.state {
	case StateSelectProfile:
		return m.renderProfileSelection()
	case StateConfirmSync:
		return m.renderSyncConfirmation()
	case StateSyncProgress:
		return m.renderSyncProgress()
	case StateCompleted:
		return m.renderCompleted()
	default:
		return "알 수 없는 상태입니다."
	}
}

// updateProfileSelection 프로필 선택 업데이트
func (m SyncModel) updateProfileSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		// 이전 프로필 선택 로직
	case "down", "j":
		// 다음 프로필 선택 로직
	case "enter":
		m.state = StateConfirmSync
		return m, nil
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

// updateSyncConfirmation 동기화 확인 업데이트
func (m SyncModel) updateSyncConfirmation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		m.confirmSync = true
		m.state = StateSyncProgress
		return m, nil
	case "n":
		return m, tea.Quit
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

// renderProfileSelection 프로필 선택 화면 렌더링
func (m SyncModel) renderProfileSelection() string {
	var content strings.Builder

	// 제목
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Render("Sync Tool - 프로필 선택")

	content.WriteString(title + "\n\n")

	// 프로필 목록
	content.WriteString("동기화할 프로필을 선택하세요:\n\n")

	for name, description := range m.profiles {
		profileStyle := lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.Color("#7D56F4"))

		if name == m.selectedProfile {
			profileStyle = profileStyle.
				Background(lipgloss.Color("#7D56F4")).
				Foreground(lipgloss.Color("#FAFAFA"))
		}

		content.WriteString(fmt.Sprintf("%s %s\n",
			profileStyle.Render("→"),
			profileStyle.Render(fmt.Sprintf("%s - %s", name, description))))
	}

	// 도움말
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render("\n↑/↓: 이동  Enter: 선택  q: 종료")

	content.WriteString(help)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content.String())
}

// renderSyncConfirmation 동기화 확인 화면 렌더링
func (m SyncModel) renderSyncConfirmation() string {
	var content strings.Builder

	// 제목
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Render("동기화 확인")

	content.WriteString(title + "\n\n")

	// 변경사항 요약
	content.WriteString("변경사항 요약:\n")
	content.WriteString(fmt.Sprintf("• 복사할 파일: %d개\n", len(m.changes.Changes)))
	content.WriteString(fmt.Sprintf("• 삭제할 파일: %d개\n\n", len(m.changes.Deletions)))

	// 복사할 파일 목록
	if len(m.changes.Changes) > 0 {
		content.WriteString("복사할 파일 목록:\n")
		content.WriteString("────────────────────\n")
		for i, change := range m.changes.Changes {
			if i >= 10 { // 최대 10개만 표시
				content.WriteString(fmt.Sprintf("... 및 %d개 더\n", len(m.changes.Changes)-10))
				break
			}
			content.WriteString(fmt.Sprintf("  %s %s\n",
				getChangeTypeIcon(change.Type), change.Path))
		}
		content.WriteString("\n")
	}

	// 삭제할 파일 목록
	if len(m.changes.Deletions) > 0 {
		content.WriteString("삭제할 파일 목록:\n")
		content.WriteString("────────────────────\n")
		for i, deletion := range m.changes.Deletions {
			if i >= 10 { // 최대 10개만 표시
				content.WriteString(fmt.Sprintf("... 및 %d개 더\n", len(m.changes.Deletions)-10))
				break
			}
			content.WriteString(fmt.Sprintf("  🗑️  %s\n", deletion))
		}
		content.WriteString("\n")
	}

	// 확인 메시지
	content.WriteString("위 파일들을 동기화하시겠습니까?\n\n")
	content.WriteString("y: 예, n: 아니오, q: 종료")

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content.String())
}

// renderSyncProgress 동기화 진행 화면 렌더링
func (m SyncModel) renderSyncProgress() string {
	content := "동기화 진행 중...\n\n"
	content += "⏳ 파일을 처리하고 있습니다.\n"
	content += "\n잠시만 기다려주세요."

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// renderCompleted 완료 화면 렌더링
func (m SyncModel) renderCompleted() string {
	content := "✅ 동기화가 완료되었습니다!\n\n"

	if m.err != nil {
		content = fmt.Sprintf("❌ 동기화 중 오류가 발생했습니다:\n\n%s\n\n", m.err.Error())
	}

	content += "q: 종료"

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// getChangeTypeIcon 변경 타입에 따른 아이콘 반환
func getChangeTypeIcon(changeType sync.ChangeType) string {
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

// NewSyncModel 새로운 동기화 모델 생성
func NewSyncModel(profiles map[string]string, changes *sync.SyncResult) SyncModel {
	return SyncModel{
		state:    StateSelectProfile,
		profiles: profiles,
		changes:  changes,
	}
}

// ShowSyncUI 동기화 UI 표시
func ShowSyncUI(profiles map[string]string, changes *sync.SyncResult) error {
	// TUI 기능은 현재 개발 중입니다.
	// 임시로 기본 텍스트 출력으로 대체
	fmt.Println("=== 동기화 프로필 선택 ===")
	for name, description := range profiles {
		fmt.Printf("• %s: %s\n", name, description)
	}

	fmt.Println("\n=== 변경사항 ===")
	fmt.Printf("복사할 파일: %d개\n", len(changes.Changes))
	fmt.Printf("삭제할 파일: %d개\n", len(changes.Deletions))

	if len(changes.Changes) > 0 {
		fmt.Println("\n복사할 파일 목록:")
		for _, change := range changes.Changes {
			fmt.Printf("  %s %s\n", getChangeTypeIcon(change.Type), change.Path)
		}
	}

	if len(changes.Deletions) > 0 {
		fmt.Println("\n삭제할 파일 목록:")
		for _, deletion := range changes.Deletions {
			fmt.Printf("  🗑️  %s\n", deletion)
		}
	}

	return nil
}
