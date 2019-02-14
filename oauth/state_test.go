package oauth

import (
	"crypto/sha256"
	"testing"

	"github.com/gorilla/securecookie"
	"github.com/stretchr/testify/assert"
)

func TestGenerateValidate(t *testing.T) {
	sv := newStateManager()

	state, err := sv.Create()
	assert.NoError(t, err)
	state.Origin = "foo"

	encState, err := sv.Encode(state)
	assert.NoError(t, err)

	state2, err := sv.Decode(encState)
	assert.NoError(t, err)

	assert.Equal(t, state, state2)
}

func TestGenerateDifferent(t *testing.T) {
	sv := newStateManager()

	state, err := sv.Create()
	assert.NoError(t, err)

	state2, err := sv.Create()
	assert.NoError(t, err)

	assert.NotEqual(t, state, state2, "should generate different states")
}

func newStateManager() StateManager {
	authKey := sha256.Sum256([]byte("authKey"))
	encKey := sha256.Sum256([]byte("encKey"))
	codecs := securecookie.CodecsFromPairs(authKey[:], encKey[:])
	return NewStateManager(codecs...)
}
