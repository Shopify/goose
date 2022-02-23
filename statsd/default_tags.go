package statsd

import (
	"os"
	"strings"
)

const StatsdDefaultTags = "STATSD_DEFAULT_TAGS"

func defaultTagsFromEnv() []string {
	statsdTags, found := os.LookupEnv(StatsdDefaultTags)
	if !found || statsdTags == "" {
		return nil
	}

	const defaultSep = ","
	return strings.Split(statsdTags, defaultSep)
}
