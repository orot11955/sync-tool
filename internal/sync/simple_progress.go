package sync

import (
	"fmt"
	"strings"
	"time"
)

// SimpleProgress 간단한 진행률 표시기
type SimpleProgress struct {
	TotalFiles  int
	CurrentFile int
	StartTime   time.Time
	LastUpdate  time.Time
}

// NewSimpleProgress 새로운 진행률 표시기 생성
func NewSimpleProgress(totalFiles int) *SimpleProgress {
	return &SimpleProgress{
		TotalFiles:  totalFiles,
		CurrentFile: 0,
		StartTime:   time.Now(),
		LastUpdate:  time.Now(),
	}
}

// Update 진행률 업데이트
func (p *SimpleProgress) Update(currentFile int, filename string) {
	p.CurrentFile = currentFile
	now := time.Now()

	// 1초마다 업데이트
	if now.Sub(p.LastUpdate) < time.Second && currentFile < p.TotalFiles {
		return
	}

	p.LastUpdate = now

	// 진행률 계산
	percent := 0.0
	if p.TotalFiles > 0 {
		percent = float64(currentFile) / float64(p.TotalFiles) * 100
	}

	// 진행률 바 생성
	barWidth := 30
	filled := int(float64(barWidth) * percent / 100)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	// 경과 시간 계산
	elapsed := now.Sub(p.StartTime)

	// 파일명 표시 (길이 제한)
	displayName := filename
	if len(displayName) > 40 {
		displayName = "..." + displayName[len(displayName)-37:]
	}

	// 진행률 출력
	fmt.Printf("\r📊 [%d/%d] %s |%s| %.1f%% (%v)",
		currentFile, p.TotalFiles, displayName, bar, percent, elapsed.Round(time.Second))

	// 완료 시 새 줄
	if currentFile >= p.TotalFiles {
		fmt.Println()
	}
}

// Complete 완료 메시지
func (p *SimpleProgress) Complete() {
	totalTime := time.Since(p.StartTime)
	fmt.Printf("✅ 동기화 완료! 총 %d개 파일, 소요 시간: %v\n",
		p.TotalFiles, totalTime.Round(time.Second))
}
