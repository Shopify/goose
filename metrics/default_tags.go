package metrics

import (
	"os"
	"strings"
)

const (
	StatsdDefaultTags = "STATSD_DEFAULT_TAGS"
)

func defaultTagsFromEnv() Tags {
	return TagsFromEnv(StatsdDefaultTags)
}

func TagsFromEnv(name string) Tags {
	return SplitTags(os.Getenv(name))
}

func SplitTags(csv string) Tags {
	kvs := strings.Split(csv, ",")
	tags := make(Tags, len(kvs))
	for _, kv := range kvs {
		if kv == "" {
			continue
		}
		parts := strings.SplitN(kv, ":", 2)
		tags[parts[0]] = parts[1] // TODO syntax error handling
	}
	return tags
}
