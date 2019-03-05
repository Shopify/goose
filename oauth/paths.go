package oauth

import (
	"fmt"
	"net/url"
	"strings"
)

type Paths struct {
	// RootURL is the absolute path to this server.
	// Ex: https://example.com
	RootURL string

	// LoginPath is the path, relative to RootURL, to send the user to initialize the login.
	// Ex: /oauth/login
	LoginPath string

	// CallbackPath is the path, relative to RootURL, which the upstream will redirect after login.
	// Ex: /oauth/callback
	CallbackPath string

	// RedirectPath is the path, relative to RootURL, which the user will be redirected after login.
	// Ex: /home
	// If LoginPath is called with a ?redirect= parameter, it will override this
	RedirectPath string
}

func (p *Paths) LoginURL(redirectPath string) *url.URL {
	u := joinURL(p.RootURL, p.LoginPath)
	if redirectPath != "" {
		q := u.Query()
		q.Set("redirect", redirectPath)
		u.RawQuery = q.Encode()
	}
	return u
}

func (p *Paths) CallbackURL() *url.URL {
	return joinURL(p.RootURL, p.CallbackPath)
}

func (p *Paths) RedirectURL(path string) *url.URL {
	if path == "" {
		path = p.RedirectPath
	}
	return joinURL(p.RootURL, path)
}

func joinURL(a, b string) *url.URL {
	rawURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(a, "/"), strings.TrimPrefix(b, "/"))
	u, err := url.Parse(rawURL)
	if err != nil {
		log(nil, err).WithField("rawURL", rawURL).Fatal("unable to construct url")
	}
	return u
}
