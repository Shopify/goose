package configure

import (
	"github.com/iancoleman/strcase"
	"github.com/sirupsen/logrus"

	"github.com/Shopify/goose/v2/metrics"
)

type Config struct {
	Environment string
	Service     string
	Version     string

	Bugsnag struct {
		APIKey          string
		ProjectPackages []string // Example: []string{"github.com/my-org/my-project*"}
	}

	Logger struct {
		GlobalFields logrus.Fields
	}

	Metrics struct {
		Implementation string
		Prefix         string       // Defaults to strcase.ToCamel(Service)
		GlobalTags     metrics.Tags // Tags in STATSD_DEFAULT_TAGS are already automatically parsed.

		Statsd struct {
			Addr string
		}
	}

	Profiler struct {
		CloudProfiler struct {
			Enabled   bool
			ProjectID string
		}
	}
}

func (c *Config) Runtime() string {
	r := c.Service
	if c.Environment != "" {
		if r != "" {
			r += "-"
		}
		r += c.Environment
	}
	return r
}

func (c *Config) MetricsPrefix() string {
	if c.Metrics.Prefix != "" {
		return c.Metrics.Prefix
	}
	return strcase.ToCamel(c.Service)
}
