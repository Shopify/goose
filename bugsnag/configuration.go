package bugsnag

import (
	logrusbugsnag "github.com/Shopify/logrus-bugsnag"
	bugsnaggo "github.com/bugsnag/bugsnag-go/v2"
	"github.com/sirupsen/logrus"

	"github.com/Shopify/goose/v2/logger"
)

func AutoConfigure(apiKey string, appVersion string, releaseStage string, packages []string) {
	config := NewConfiguration(apiKey, appVersion, releaseStage, packages)
	bugsnaggo.Configure(config)
	RegisterLogger()
	RegisterDefaultHooks()
}

func RegisterLogger() {
	defaultLogger := logrus.StandardLogger()
	// Duplicate to avoid creating a report loop
	bugsnagLogger := logger.DuplicateLogger(defaultLogger)
	logger.RegisterHook(bugsnagLogger)
	defaultLogger.AddHook(&logrusbugsnag.Hook{})
	bugsnaggo.Config.Logger = bugsnagLogger
}

func NewConfiguration(apiKey string, appVersion string, releaseStage string, packages []string) bugsnaggo.Configuration {
	packages = append(packages, "main*")

	return bugsnaggo.Configuration{
		APIKey:          apiKey,
		AppVersion:      appVersion,
		ProjectPackages: packages,
		ReleaseStage:    releaseStage,
		Synchronous:     true,
	}
}
