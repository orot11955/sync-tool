package logger

import (
	"fmt"
	"io"
	"os"
	"strings"

	"sync-tool/internal/config"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Init 로거 초기화
func Init(cfg *config.LoggingConfig) error {
	log = logrus.New()

	// 로그 레벨 설정
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		return fmt.Errorf("잘못된 로그 레벨: %s", cfg.Level)
	}
	log.SetLevel(level)

	// 로그 포맷 설정
	switch strings.ToLower(cfg.Format) {
	case "json":
		log.SetFormatter(&logrus.JSONFormatter{})
	default:
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
			DisableColors: false,
		})
	}

	// 로그 출력 설정
	var output io.Writer = os.Stdout
	if cfg.File != "" {
		file, err := os.OpenFile(cfg.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("로그 파일 열기 실패: %w", err)
		}
		output = io.MultiWriter(os.Stdout, file)
	}
	log.SetOutput(output)

	return nil
}

// GetLogger 로거 인스턴스 반환
func GetLogger() *logrus.Logger {
	if log == nil {
		// 기본 설정으로 초기화
		log = logrus.New()
		log.SetLevel(logrus.InfoLevel)
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
			DisableColors: false,
		})
	}
	return log
}

// Debug 디버그 로그
func Debug(args ...interface{}) {
	GetLogger().Debug(args...)
}

// Debugf 포맷된 디버그 로그
func Debugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...)
}

// Info 정보 로그
func Info(args ...interface{}) {
	GetLogger().Info(args...)
}

// Infof 포맷된 정보 로그
func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}

// Warn 경고 로그
func Warn(args ...interface{}) {
	GetLogger().Warn(args...)
}

// Warnf 포맷된 경고 로그
func Warnf(format string, args ...interface{}) {
	GetLogger().Warnf(format, args...)
}

// Error 오류 로그
func Error(args ...interface{}) {
	GetLogger().Error(args...)
}

// Errorf 포맷된 오류 로그
func Errorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}

// Fatal 치명적 오류 로그 (프로그램 종료)
func Fatal(args ...interface{}) {
	GetLogger().Fatal(args...)
}

// Fatalf 포맷된 치명적 오류 로그 (프로그램 종료)
func Fatalf(format string, args ...interface{}) {
	GetLogger().Fatalf(format, args...)
}
