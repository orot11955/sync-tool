package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config 전체 설정 구조체
type Config struct {
	Server   ServerConfig           `yaml:"server"`
	Sync     SyncConfig             `yaml:"sync"`
	Profiles map[string]SyncProfile `yaml:"profiles"`
	Logging  LoggingConfig          `yaml:"logging"`
	UI       UIConfig               `yaml:"ui"`
}

// ServerConfig 서버 연결 설정
type ServerConfig struct {
	Host    string `yaml:"host"`
	User    string `yaml:"user"`
	Port    int    `yaml:"port"`
	KeyPath string `yaml:"key_path"`
}

// SyncConfig 동기화 기본 설정
type SyncConfig struct {
	Options         []string `yaml:"options" mapstructure:"options"`
	DefaultExcludes []string `yaml:"default_excludes" mapstructure:"default_excludes"`
}

// SyncProfile 동기화 프로필
type SyncProfile struct {
	Name        string   `yaml:"name" mapstructure:"name"`
	Description string   `yaml:"description" mapstructure:"description"`
	ServerPath  string   `yaml:"server_path" mapstructure:"server_path"`
	LocalPath   string   `yaml:"local_path" mapstructure:"local_path"`
	Options     []string `yaml:"options,omitempty" mapstructure:"options"`
	Includes    []string `yaml:"includes,omitempty" mapstructure:"includes"`
	Excludes    []string `yaml:"excludes,omitempty" mapstructure:"excludes"`
}

// LoggingConfig 로깅 설정
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	File   string `yaml:"file"`
}

// UIConfig UI 설정
type UIConfig struct {
	Theme          string `yaml:"theme"`
	ShowProgress   bool   `yaml:"show_progress"`
	ConfirmActions bool   `yaml:"confirm_actions"`
}

// InitializeConfig 설정 파일 초기화
func InitializeConfig(configPath string) error {
	// 설정 파일이 이미 존재하는지 확인
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("설정 파일이 이미 존재합니다: %s", configPath)
	}

	// 기본 설정 생성
	defaultConfig := &Config{
		Server: ServerConfig{
			Host:    "10.10.30.237",
			User:    "root",
			Port:    22,
			KeyPath: "",
		},
		Sync: SyncConfig{
			Options: []string{"-r", "-z", "-c", "--itemize-changes"},
			DefaultExcludes: []string{
				".DS_Store",
				"._*",
				".Spotlight-V100",
				".Trashes",
				".TemporaryItems",
				"System Volume Information",
				"System Volume Information/*",
				".fseventsd",
				".fseventsd/*",
				".Trash-1000",
				".Trash-1000/*",
			},
		},
		Profiles: map[string]SyncProfile{
			"aunes_ins": {
				Name:        "AUNES_INS",
				Description: "AUNES INS 폴더 동기화",
				ServerPath:  "/stor2/USB_SYNC/AUNES_INS",
				LocalPath:   "/Volumes/AUNES_INS",
				Excludes:    []string{},
			},
			"ventoy": {
				Name:        "Ventoy (KICKSTART, config)",
				Description: "Ventoy 폴더 동기화 (ISO 파일 제외)",
				ServerPath:  "/stor2/USB_SYNC/Ventoy",
				LocalPath:   "/Volumes/Ventoy",
				Excludes:    []string{"_iso/*.iso"},
			},
			"iso_only": {
				Name:        "ISO Files Only",
				Description: "ISO 파일만 동기화",
				ServerPath:  "/stor2/USB_SYNC/Ventoy",
				LocalPath:   "/Volumes/Ventoy",
				Includes:    []string{"*.iso"},
				Excludes:    []string{"*"},
				Options:     []string{"-r", "-z", "--itemize-changes"},
			},
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
			File:   "",
		},
		UI: UIConfig{
			Theme:          "default",
			ShowProgress:   true,
			ConfirmActions: true,
		},
	}

	// 디렉토리 생성
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("디렉토리 생성 실패: %w", err)
	}

	// YAML 파일로 저장
	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		return fmt.Errorf("설정 마샬링 실패: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("설정 파일 저장 실패: %w", err)
	}

	fmt.Printf("설정 파일이 생성되었습니다: %s\n", configPath)
	return nil
}

// GetSyncOptions 프로필의 동기화 옵션 반환
func (p *SyncProfile) GetSyncOptions(baseOptions []string) []string {
	if len(p.Options) > 0 {
		return p.Options
	}
	return baseOptions
}

// GetExcludes 프로필의 제외 패턴 반환 (기본 제외 패턴 포함)
func (p *SyncProfile) GetExcludes(defaultExcludes []string) []string {
	excludes := make([]string, 0, len(defaultExcludes)+len(p.Excludes))
	excludes = append(excludes, defaultExcludes...)
	excludes = append(excludes, p.Excludes...)
	return excludes
}
