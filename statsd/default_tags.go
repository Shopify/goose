package statsd

import (
	"os"
	"strings"
)

const (
	StatsdDefaultTags = "STATSD_DEFAULT_TAGS"
	DefaultSep        = ","
)

func defaultTagsFromEnv() []string {
	statsdTags, found := os.LookupEnv(StatsdDefaultTags)
	if !found || statsdTags == "" {
		return nil
	}
	return strings.Split(statsdTags, DefaultSep)
}
