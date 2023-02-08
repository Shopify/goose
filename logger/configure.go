package logger

import (
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	defaultLogger     *logrus.Logger
	defaultLoggerOnce sync.Once
)

func RegisterStandardLoggerHook() {
	RegisterHook(logrus.StandardLogger())
}

func RegisterHook(logger *logrus.Logger) {
	logger.AddHook(defaultHook)
}

func DefaultLogger() *logrus.Logger {
	defaultLoggerOnce.Do(func() {
		defaultLogger = DuplicateLogger(logrus.StandardLogger())
		defaultLogger.AddHook(defaultHook)
	})

	return defaultLogger
}

func DuplicateLogger(source *logrus.Logger) *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(source.Out)
	logger.SetFormatter(source.Formatter)
	logger.SetLevel(source.Level)
	logger.SetReportCaller(source.ReportCaller)

	for level, hooks := range logger.Hooks {
		logger.Hooks[level] = append(logger.Hooks[level], hooks...)
	}

	return logger
}
