package oauth

import (
	"context"
	"crypto/sha256"
	"encoding/gob"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/Shopify/goose/logger"
)

var log = logger.New("oauth")

const domain = "shopify.com"
const cookieName = "oauthSession"
const tokenCookieKey = "token"

// Manager wraps all the logic and configuration of the OAuth flow.
type Manager interface {
	// GetLoginURL builds a URL pointing to the login page to start the auth flow
	// `origin` is the URL the user should redirected to once the login completes
	GetLoginURL(ctx context.Context, origin string) (string, error)

	DecodeState(encState string) (*State, error)

	ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error)

	// GetClient returns an http client to make authenticated requests
	GetClient(ctx context.Context, token *oauth2.Token) *http.Client

	// SaveToken saves the token in the user's session to be used for futures authentications
	SaveToken(w http.ResponseWriter, r *http.Request, token *oauth2.Token) error

	// GetSavedToken returns the token saved in the current session
	GetSavedToken(r *http.Request) (*oauth2.Token, bool)

	// AuthorizeToken authenticates and authorizes the user
	AuthorizeToken(ctx context.Context, token *oauth2.Token) (*User, error)
}

func init() {
	// For (de)serialization
	gob.Register(oauth2.Token{})
	gob.Register(State{})
}

func NewManager(oauthConfig *oauth2.Config, authenticator Authenticator, authorizer Authorizer) Manager {
	authKey := sha256.Sum256([]byte(oauthConfig.ClientSecret + "authKey"))
	encKey := sha256.Sum256([]byte(oauthConfig.ClientSecret + "encKey"))
	codecs := securecookie.CodecsFromPairs(authKey[:], encKey[:])

	var tokenStore = &sessions.CookieStore{
		Codecs: codecs,
		Options: &sessions.Options{
			Path:     "/",
			MaxAge:   int(time.Hour.Seconds()), // Google OAuth tokens are only valid for an hour
			Secure:   strings.Index(oauthConfig.RedirectURL, "https://") == 0,
			HttpOnly: true,
		},
	}

	return &manager{
		oauthConfig:   oauthConfig,
		tokenStore:    tokenStore,
		stateManager:  NewStateManager(codecs...),
		authenticator: authenticator,
		authorizer:    authorizer,
	}
}

type manager struct {
	oauthConfig   *oauth2.Config
	stateManager  StateManager
	tokenStore    sessions.Store
	authenticator Authenticator
	authorizer    Authorizer
}

// Parameters come from Google redirecting to /oauth?state=STATE&code=CODE
func (m *manager) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return m.oauthConfig.Exchange(ctx, code)
}

// Parameters come from Google redirecting to /oauth?state=STATE&code=CODE
func (m *manager) DecodeState(encState string) (*State, error) {
	return m.stateManager.Decode(encState)
}

func (m *manager) GetSavedToken(r *http.Request) (*oauth2.Token, bool) {
	session, err := m.tokenStore.Get(r, cookieName)
	if err != nil {
		log(r.Context(), err).Warn("unable to decode cookie session")
		return nil, false
	}

	token, ok := session.Values[tokenCookieKey].(oauth2.Token)
	return &token, ok
}

func (m *manager) SaveToken(w http.ResponseWriter, r *http.Request, token *oauth2.Token) error {
	session, err := m.tokenStore.Get(r, cookieName)
	if err != nil {
		log(r.Context(), err).Warn("unable to decode cookie session")
		return err
	}

	session.Values[tokenCookieKey] = *token
	return session.Save(r, w)
}

func (m *manager) GetLoginURL(ctx context.Context, origin string) (string, error) {
	state, err := m.stateManager.Create()
	if err != nil {
		log(ctx, err).Warn("unable to create a state")
		return "", err
	}

	state.Origin = origin
	encState, err := m.stateManager.Encode(state)
	if err != nil {
		log(ctx, err).Warn("unable to encode the state")
		return "", err
	}

	return m.oauthConfig.AuthCodeURL(encState, oauth2.SetAuthURLParam("hd", domain)), nil
}

func (m *manager) GetClient(ctx context.Context, token *oauth2.Token) *http.Client {
	return m.oauthConfig.Client(ctx, token)
}

func (m *manager) AuthorizeToken(ctx context.Context, token *oauth2.Token) (*User, error) {
	user, err := m.authenticator(ctx, m, token)
	if err != nil {
		return user, errors.Wrap(err, "user is not authenticated")
	}

	if m.authorizer != nil {
		if err := m.authorizer(ctx, m, token, user); err != nil {
			return user, errors.Wrap(err, "user is not authorized")
		}
	}

	return user, nil
}
