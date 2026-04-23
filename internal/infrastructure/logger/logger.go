package logger

import (
	"strings"

	"go.uber.org/zap"
)

var global *zap.Logger

func Init(env string) error {
	var (
		l   *zap.Logger
		err error
	)

	if strings.EqualFold(env, "local") || strings.EqualFold(env, "dev") || strings.EqualFold(env, "development") {
		l, err = zap.NewDevelopment()
	} else {
		l, err = zap.NewProduction()
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
