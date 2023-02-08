package errors

import (
	"github.com/Shopify/goose/logger"
	"github.com/pkg/errors"
)

type LoggableError interface {
	error
	logger.Loggable
}

func FieldsFromError(err error) Fields {
	var loggable LoggableError
	if errors.As(err, &loggable) {
		return Fields(loggable.LogFields())
	}

	return Fields{}
}

func mergeFields(ctx logger.Valuer, err error, fieldsList ...Fields) Fields {
	// context > fields > parent error
	fields := Fields{}
	fieldsList = append([]Fields{FieldsFromError(err)}, fieldsList...)
	fieldsList = append(fieldsList, Fields(logger.GetLoggableValues(ctx)))

	for _, fs := range fieldsList {
		for k, v := range fs {
			fields[k] = v
		}
	}
	return fields
}
