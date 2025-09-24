package sync

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ProgressInfo 진행률 정보
type ProgressInfo struct {
	File    string
	Percent int
	Speed   string
	ETA     string
	Size    string
}

// ProgressMonitor 진행률 모니터
type ProgressMonitor struct {
	TotalFiles  int
	CurrentFile int
	StartTime   time.Time
}

// NewProgressMonitor 새로운 진행률 모니터 생성
func NewProgressMonitor(totalFiles int) *ProgressMonitor {
	return &ProgressMonitor{
		TotalFiles:  totalFiles,
		CurrentFile: 0,
		StartTime:   time.Now(),
	}
}

// ParseProgressLine rsync 진행률 라인 파싱
func ParseProgressLine(line string) (*ProgressInfo, error) {
	// rsync 진행률 형식: filename
	//                  123,456,789 12%  1.23MB/s    0:00:12 (xfr#1, to-chk=0/1)

	// 파일명 추출 (첫 번째 줄)
	if !strings.Contains(line, "%") && !strings.Contains(line, "xfr#") {
		return &ProgressInfo{
			File: strings.TrimSpace(line),
		}, nil
	}

	// 진행률 정보 파싱
	re := regexp.MustCompile(`(\d+)%\s+([\d.]+\w+/s)?\s*([\d:]+)?`)
	matches := re.FindStringSubmatch(line)

	if len(matches) < 2 {
		return nil, fmt.Errorf("진행률 파싱 실패: %s", line)
	}

	percent, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("퍼센트 파싱 실패: %s", matches[1])
	}

	info := &ProgressInfo{
		Percent: percent,
	}

	if len(matches) > 2 && matches[2] != "" {
		info.Speed = matches[2]
	}

	if len(matches) > 3 && matches[3] != "" {
		info.ETA = matches[3]
	}

	return info, nil
}

// MonitorRsyncProgress rsync 진행률 모니터링
func MonitorRsyncProgress(cmd *exec.Cmd, monitor *ProgressMonitor) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout 파이프 생성 실패: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr 파이프 생성 실패: %w", err)
	}

	// 진행률 출력을 stdout으로 리다이렉트
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 진행률 모니터링 시작
	go func() {
		scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
		var currentFile string

		for scanner.Scan() {
			line := scanner.Text()

			// 파일명 라인
			if !strings.Contains(line, "%") && !strings.Contains(line, "xfr#") && strings.TrimSpace(line) != "" {
				currentFile = strings.TrimSpace(line)
				monitor.CurrentFile++
				fmt.Printf("\r📄 [%d/%d] %s", monitor.CurrentFile, monitor.TotalFiles, currentFile)
				continue
			}

			// 진행률 라인
			if progress, err := ParseProgressLine(line); err == nil && progress.Percent > 0 {
				fmt.Printf("\r📊 [%d/%d] %s - %d%%",
					monitor.CurrentFile, monitor.TotalFiles, currentFile, progress.Percent)

				if progress.Speed != "" {
					fmt.Printf(" (%s", progress.Speed)
					if progress.ETA != "" {
						fmt.Printf(", ETA: %s", progress.ETA)
					}
					fmt.Printf(")")
				}
			}
		}
	}()

	return nil
}

// ShowProgressBar 간단한 진행률 바 표시
func ShowProgressBar(current, total int, filename string) {
	if total == 0 {
		return
	}

	percent := float64(current) / float64(total) * 100
	barWidth := 50
	filled := int(float64(barWidth) * percent / 100)

	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	fmt.Printf("\r📊 [%d/%d] %s |%s| %.1f%%",
		current, total, filename, bar, percent)

	if current == total {
		fmt.Println()
	}
}
