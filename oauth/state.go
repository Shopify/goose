package oauth

import (
	"crypto/rand"

	"github.com/gorilla/securecookie"
)

type StateManager interface {
	Create() (*State, error)
	Encode(state *State) (string, error)
	Decode(encState string) (*State, error)
}

type stateManager struct {
	codecs []securecookie.Codec
}

func NewStateManager(codecs ...securecookie.Codec) StateManager {
	return &stateManager{
		codecs: codecs,
	}
}

type State struct {
	RandBytes []byte
	Origin    string
}

const tokenNumBytes = 32
const stateSessionName = "state"

func (m *stateManager) Create() (*State, error) {
	randBytes := make([]byte, tokenNumBytes)
	if _, err := rand.Read(randBytes); err != nil {
		return nil, err
	}

	return &State{RandBytes: randBytes}, nil
}

func (m *stateManager) Encode(state *State) (string, error) {
	return securecookie.EncodeMulti(stateSessionName, state, m.codecs...)
}

func (m *stateManager) Decode(encState string) (*State, error) {
	state := &State{}

	err := securecookie.DecodeMulti(stateSessionName, encState, state, m.codecs...)
	if err != nil {
		return nil, err
	}

	return state, nil
}
