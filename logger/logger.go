/* Copyright (c) 2018, joy.zhou <chowyu08@gmail.com>
 */

package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// env can be setup at build time with Go Linker. Value could be prod or whatever else for dev env
	Instance   *zap.Logger
	logCfg     zap.Config
	encoderCfg = zap.NewProductionEncoderConfig()
)

func init() {
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
}

// NewDevLogger return a logger for dev builds
func NewDevLogger() (*zap.Logger, error) {
	logCfg := zap.NewProductionConfig()
	logCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	// logCfg.DisableStacktrace = true
	logCfg.EncoderConfig = encoderCfg
	return logCfg.Build()
}

// NewProdLogger return a logger for production builds
func NewProdLogger(Level zapcore.Level) (*zap.Logger, error) {
	logCfg := zap.NewProductionConfig()
	logCfg.DisableStacktrace = true
	logCfg.Level = zap.NewAtomicLevelAt(Level)
	logCfg.EncoderConfig = encoderCfg
	return logCfg.Build()
}
