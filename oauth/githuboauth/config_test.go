package githuboauth_test

import (
	"github.com/Shopify/goose/oauth"
	"github.com/Shopify/goose/oauth/githuboauth"
)

func ExampleNewConfig() {
	clientID := ""
	clientSecret := ""

	paths := &oauth.Paths{
		RootURL:      "https://example.com",
		LoginPath:    "/oauth/login",
		CallbackPath: "/oauth/callback",
		RedirectPath: "/homepage",
	}

	scopes := []string{"repo"}
	config := githuboauth.NewConfig(clientID, clientSecret, paths, scopes)
	authorizer := githuboauth.NewOrgAuthorizer("example")

	_ = oauth.NewManager(config, githuboauth.Authenticator, authorizer)
}
