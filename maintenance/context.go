package maintenance

import (
	"context"
)

// BackgroundContextWithValues returns a new context that propagate values but not cancellation
func BackgroundContextWithValues(contextForValues context.Context) context.Context {
	return &contextWithValues{context.Background(), contextForValues}
}

type contextWithValues struct {
	context.Context
	valueCtx context.Context
}

func (c *contextWithValues) Value(key interface{}) interface{} {
	return c.valueCtx.Value(key)
}
