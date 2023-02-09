package configure

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/Shopify/goose/v2/bugsnag"
)

// Bugsnag should be the very first thing done in an application.
func Bugsnag(config *Config) {
	if config.Bugsnag.APIKey == "" {
		if apiKey := os.Getenv("BUGSNAG_API_KEY"); apiKey != "" {
			config.Bugsnag.APIKey = apiKey
		}

		if config.Bugsnag.APIKey == "" {
			logrus.Warn("skipping bugsnag configuration")
		}
		return
	}

	logrus.Info("configuring bugsnag")

	bugsnag.AutoConfigure(config.Bugsnag.APIKey, config.Version, config.Runtime(), config.Bugsnag.ProjectPackages)
}
