package configure

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/Shopify/goose/v2/metrics"
)

func Metrics(config *Config) {
	if config.Metrics.Implementation != "" {
		mustSetEnvVar(metrics.DefaultImplementationEnvVar, config.Metrics.Implementation)
	}

	if config.Metrics.Statsd.Addr != "" {
		mustSetEnvVar(metrics.DefaultStatsdEndpoint, config.Metrics.Statsd.Addr)
	}

	b, err := metrics.NewBackendFromEnv()
	if err != nil {
		logrus.WithError(err).Error("unable to initialize metrics backend")
		return
	}

	b = metrics.BackendWithDefaultWrappers(b, config.MetricsPrefix())

	tags := config.Metrics.GlobalTags
	if env := config.Environment; env != "" {
		tags = tags.WithTag("environment", env)
	}
	if service := config.Service; service != "" {
		tags = tags.WithTag("service", service)
	}
	if len(tags) > 0 {
		b = metrics.NewTagsWrapper(b, tags)
	}

	metrics.SetDefaultBackend(b)
}

func mustSetEnvVar(key, val string) {
	err := os.Setenv(key, val)
	if err != nil {
		logrus.WithError(err).WithField("key", key).Panic("unable to set environment variable")
	}
}
