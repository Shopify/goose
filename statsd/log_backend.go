package statsd

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

// NewLogBackend creates a new Backend that emits statsd metrics to the logger.
//
// namespace is a global namespace to apply to every metric emitted.
// tags is a global set of tags that will be added to every metric emitted.
func NewLogBackend(namespace string, tags []string) Backend {
	lb := &logBackend{
		namespace: namespace,
		tags:      tags,
	}
	return NewForwardingBackend(lb.log)
}

type logBackend struct {
	namespace string
	tags      []string
}

func (b *logBackend) log(ctx context.Context, mType string, name string, value interface{}, tags []string, rate float64) error {
	log(ctx, nil).WithFields(logrus.Fields{
		"metric": fmt.Sprintf("%s%s", b.namespace, name),
		"type":   mType,
		"tags":   append(b.tags, tags...),
		"value":  value,
		"rate":   rate,
	}).Debug("emit statsd")
	return nil
}
