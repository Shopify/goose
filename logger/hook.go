package logger

import (
	"errors"

	"github.com/sirupsen/logrus"
)

var defaultHook hook
var GlobalFields = logrus.Fields{}

type hook struct{}

func (h hook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h hook) Fire(entry *logrus.Entry) error {
	for k, v := range GlobalFields {
		entry.Data[k] = v
	}

	if entry.Context != nil {
		for k, v := range GetLoggableValues(entry.Context) {
			entry.Data[k] = v
		}
	}

	if err, ok := entry.Data[logrus.ErrorKey].(error); ok {
		var loggable loggableError
		if errors.As(err, &loggable) {
			for k, v := range loggable.LogFields() {
				entry.Data[k] = v
			}
		}
	}

	return nil
}
