package configure

import (
	"cloud.google.com/go/profiler"
	"github.com/sirupsen/logrus"
)

func Profiling(config *Config) {
	if !config.Profiler.CloudProfiler.Enabled {
		logrus.Warn("skipping profiler configuration")
	}

	logrus.Info("configuring profiler")

	err := profiler.Start(profiler.Config{
		Service:        config.Runtime(),
		ServiceVersion: config.Version,
		ProjectID:      config.Profiler.CloudProfiler.ProjectID,
	})
	if err != nil {
		logrus.WithError(err).Error("failed to start profiler")
	}
}
