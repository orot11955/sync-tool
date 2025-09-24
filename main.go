package main

import (
	"fmt"
	"os"

	"sync-tool/internal/app"
	"sync-tool/internal/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	configFile string
	verbose    bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "sync-tool",
		Short: "서버와 USB 간 파일 동기화 도구",
		Long:  `Git과 유사한 동작 방식을 가진 서버-USB 파일 동기화 도구입니다.`,
	}

	// 글로벌 플래그
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "config.yaml", "설정 파일 경로")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "상세 로그 출력")

	// 서브 커맨드들
	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(syncCmd())
	rootCmd.AddCommand(profilesCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "오류 발생: %v\n", err)
		os.Exit(1)
	}
}

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "설정 파일 초기화",
		RunE: func(cmd *cobra.Command, args []string) error {
			return config.InitializeConfig(configFile)
		},
	}
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "동기화 상태 확인",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			return app.ShowStatus(cfg)
		},
	}
}

func syncCmd() *cobra.Command {
	var profile string
	var dryRun bool
	var autoConfirm bool
	var useTUI bool

	cmd := &cobra.Command{
		Use:   "sync [프로필명]",
		Short: "파일 동기화 실행",
		Long:  "지정된 프로필로 서버와 로컬 간 파일을 동기화합니다.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			if len(args) > 0 {
				profile = args[0]
			}

			if useTUI {
				return app.ShowTUI(cfg, profile)
			}

			return app.Sync(cfg, profile, dryRun, autoConfirm)
		},
	}

	cmd.Flags().StringVarP(&profile, "profile", "p", "", "사용할 프로필명")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "실제 동기화 없이 변경사항만 확인")
	cmd.Flags().BoolVar(&autoConfirm, "yes", false, "확인 없이 자동 실행")
	cmd.Flags().BoolVar(&useTUI, "tui", false, "TUI 인터페이스 사용")

	return cmd
}

func profilesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "profiles",
		Short: "사용 가능한 프로필 목록 보기",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			return app.ShowProfiles(cfg)
		},
	}
}

func loadConfig() (*config.Config, error) {
	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("설정 파일 읽기 실패: %w", err)
	}

	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("설정 파싱 실패: %w", err)
	}

	// 디버그: 설정 로딩 확인
	fmt.Printf("디버그 - 기본 제외 패턴: %v\n", cfg.Sync.DefaultExcludes)

	// verbose 플래그가 설정된 경우 로그 레벨 변경
	if verbose {
		cfg.Logging.Level = "debug"
	}

	return &cfg, nil
}
