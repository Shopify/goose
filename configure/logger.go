package configure

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/Shopify/goose/v2/logger"
)

func Logger(config *Config) {
	logger.RegisterStandardLoggerHook()
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)

	if env := config.Environment; env != "" {
		logger.GlobalFields["environment"] = env
	}
	if service := config.Service; service != "" {
		logger.GlobalFields["service"] = service
	}
	if version := config.Version; version != "" {
		logger.GlobalFields["version"] = version
	}
}
