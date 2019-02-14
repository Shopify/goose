package githuboauth

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

	"github.com/Shopify/goose/oauth"
)

func NewConfig(clientID string, clientSecret string, paths *oauth.Paths, scopes []string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  paths.CallbackURL().String(),
		Scopes:       scopes,
		Endpoint:     github.Endpoint,
	}
}
