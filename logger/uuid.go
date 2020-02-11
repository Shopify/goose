package logger

import (
	"context"

	uuid "github.com/nu7hatch/gouuid"
)

const UUIDKey = "uuid"

var log = New("logger")

func WithUUID(ctx context.Context) (context.Context, string) {
	if id := GetLoggableValue(ctx, UUIDKey); id != nil {
		return ctx, id.(string)
	}

	requestID := newUUID()
	return WithField(ctx, UUIDKey, requestID), requestID
}

func newUUID() string {
	u, err := uuid.NewV4()
	if err != nil {
		log(nil, err).Error("unable to generate uuid v4")
		return "00000000-0000-0000-0000-000000000000"
	}
	return u.String()
}
