package googleoauth_test

import (
	"github.com/Shopify/goose/oauth"
	"github.com/Shopify/goose/oauth/googleoauth"
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

	config := googleoauth.NewConfig(clientID, clientSecret, paths)
	authorizer := oauth.NewCompositeAuthorizer(
		oauth.EmailVerifiedAuthorizer,
		oauth.NewDomainAuthorizer("example.com"),
	)

	_ = oauth.NewManager(config, googleoauth.Authenticator, authorizer)
}
