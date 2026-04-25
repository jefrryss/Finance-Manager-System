package logger

import (
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

var global *zap.Logger

func Init(env string, logDir string) error {
	var (
		l   *zap.Logger
		err error
	)

	if strings.TrimSpace(logDir) == "" {
		logDir = "./logs"
	}
	if err = os.MkdirAll(logDir, 0o755); err != nil {
		return err
	}

	logFile := filepath.Join(logDir, "app.log")

	if strings.EqualFold(env, "local") || strings.EqualFold(env, "dev") || strings.EqualFold(env, "development") {
		cfg := zap.NewDevelopmentConfig()
		cfg.OutputPaths = []string{"stdout", logFile}
		cfg.ErrorOutputPaths = []string{"stderr", logFile}
		l, err = cfg.Build()
	} else {
		cfg := zap.NewProductionConfig()
		cfg.OutputPaths = []string{"stdout", logFile}
		cfg.ErrorOutputPaths = []string{"stderr", logFile}
		l, err = cfg.Build()
	}
	if err != nil {
		return err
	}

	global = l
	zap.ReplaceGlobals(l)
	return nil
}

func L() *zap.Logger {
	if global != nil {
		return global
	}
	return zap.L()
}

func Sync() {
	if global != nil {
		_ = global.Sync()
	}
}
