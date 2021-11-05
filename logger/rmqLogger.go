package logger

import (
	"go.uber.org/zap"
)

type RmqLogger struct {
	logger *zap.SugaredLogger
}

func (r *RmqLogger) Debug(msg string, fields map[string]interface{}) {
	if msg == "" && len(fields) == 0 {
		return
	}
	r.logger.Debug(msg, fields)
}
func (r *RmqLogger) Info(msg string, fields map[string]interface{}) {
	if msg == "" && len(fields) == 0 {
		return
	}
	r.logger.Info(msg, fields)
}
func (r *RmqLogger) Warning(msg string, fields map[string]interface{}) {
	if msg == "" && len(fields) == 0 {
		return
	}
	r.logger.Warn(msg, fields)
}
func (r *RmqLogger) Error(msg string, fields map[string]interface{}) {
	if msg == "" && len(fields) == 0 {
		return
	}
	r.logger.Error(msg, fields)
}
func (r *RmqLogger) Fatal(msg string, fields map[string]interface{}) {
	if msg == "" && len(fields) == 0 {
		return
	}
	r.logger.Fatal(msg, fields)
}
func (r *RmqLogger) Init(sl *zap.SugaredLogger) {
	r.logger = sl
}
