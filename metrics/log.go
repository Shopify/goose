package metrics

import (
	"context"

	"github.com/sirupsen/logrus"
)

func loggerFromContext(ctx context.Context) fieldLogger {
	return logrus.StandardLogger().WithContext(ctx)
}

func logError(ctx context.Context, err error) {
	if err != nil {
		loggerFromContext(ctx).WithError(err).Error("submitting metric")
	}
}
