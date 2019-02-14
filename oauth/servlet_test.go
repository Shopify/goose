package oauth_test

import (
	"golang.org/x/oauth2"

	"github.com/Shopify/goose/oauth"
	"github.com/Shopify/goose/srvutil"
)

func ExampleNewServlet() {
	var config oauth2.Config
	var authenticator oauth.Authenticator
	var authorizer oauth.Authorizer

	paths := &oauth.Paths{
		RootURL:      "https://example.com",
		LoginPath:    "/oauth/login",
		CallbackPath: "/oauth/callback",
		RedirectPath: "/homepage",
	}

	manager := oauth.NewManager(&config, authenticator, authorizer)
	oAuthServlet := oauth.NewServlet(manager, paths)

	var protected srvutil.Servlet

	_ = srvutil.CombineServlets(
		srvutil.UseServlet(protected, oAuthServlet.Middleware),
		oAuthServlet,
	)
}
