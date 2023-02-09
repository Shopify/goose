package errors

import (
	stderrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnwrapJoinedErrors(t *testing.T) {
	joined := stderrors.Join(New("first"), New("second"))

	unwrapped := Unwrap(joined)
	assert.Nil(t, unwrapped)
}
