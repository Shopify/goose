package googleoauth

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/Shopify/goose/oauth"
)

func NewConfig(clientID string, clientSecret string, paths *oauth.Paths) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  paths.CallbackURL().String(),
		Scopes: []string{
			// https://developers.google.com/identity/protocols/googlescopes#google_sign-in
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}
}

// PreferredDomain instructs Google to prefer this domain when prompting the user to login
func PreferredDomain(domain string) oauth2.AuthCodeOption {
	return oauth2.SetAuthURLParam("hd", domain)
}
