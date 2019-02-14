package oauth

import (
	"context"

	"golang.org/x/oauth2"
)

// Authenticator creates a User from a token
type Authenticator func(ctx context.Context, m Manager, token *oauth2.Token) (*User, error)
