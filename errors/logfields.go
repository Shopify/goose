package errors

import (
	"github.com/pkg/errors"

	"github.com/Shopify/goose/v2/logger"
)

type LoggableError interface {
	error
	logger.Loggable
}

func FieldsFromError(err error) Fields {
	var loggable LoggableError
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		fs := []Fields{}
		for _, e := range joined.Unwrap() {
			if errors.As(e, &loggable) {
				fs = append(fs, Fields(loggable.LogFields()))
			}
		}
		return mergeFields(fs)
	}
	if errors.As(err, &loggable) {
		return Fields(loggable.LogFields())
	}

	return Fields{}
}

func mergeFieldsCtx(ctx logger.Valuer, err error, fieldsList ...Fields) Fields {
	fieldsList = append([]Fields{FieldsFromError(err)}, fieldsList...)
	fieldsList = append(fieldsList, Fields(logger.GetLoggableValues(ctx)))
	return mergeFields(fieldsList)
}

func mergeFields(fieldsList []Fields) Fields {
	fields := Fields{}
	for _, fs := range fieldsList {
		for k, v := range fs {
			fields[k] = v
		}
	}
	return fields
}
