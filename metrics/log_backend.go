package metrics

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

// NewLogBackend creates a new Backend that emits statsd metrics to the logger.
//
// namespace is a global namespace to apply to every metric emitted.
// tags is a global set of tags that will be added to every metric emitted.
func NewLogBackend(namespace string, tags Tags) Backend {
	lb := &logBackend{
		namespace: namespace,
		tags:      defaultTagsFromEnv().Merge(tags),
	}
	return NewForwardingBackend(lb.log)
}

type logBackend struct {
	namespace string
	tags      Tags
}

func (b *logBackend) log(ctx context.Context, mType string, name string, value interface{}, tags Tags, rate float64) error {
	log(ctx, nil).WithFields(logrus.Fields{
		"metric": fmt.Sprintf("%s%s", b.namespace, name),
		"type":   mType,
		"tags":   b.tags.Merge(tags),
		"value":  value,
		"rate":   rate,
	}).Debug("emit metric")
	return nil
}
