package oauth

import (
	"context"
	"errors"

	"golang.org/x/oauth2"
)

// Authorizer validates that a User is allowed to be logged in, throwing an error if not.
type Authorizer func(ctx context.Context, m Manager, token *oauth2.Token, user *User) error

var ErrEmailNotVerified = errors.New("user email is not verified")

func EmailVerifiedAuthorizer(_ context.Context, _ Manager, _ *oauth2.Token, user *User) error {
	if !user.EmailVerified {
		return ErrEmailNotVerified
	}
	return nil
}

var ErrInvalidDomain = errors.New("user email is not from whitelisted domain")

func NewDomainAuthorizer(domain string) Authorizer {
	return func(ctx context.Context, m Manager, _ *oauth2.Token, user *User) error {
		if user.Domain != domain {
			log(ctx, nil).
				WithField("expected", domain).
				WithField("actual", domain).
				Warn(ErrInvalidDomain.Error())
			return ErrInvalidDomain
		}
		return nil
	}
}

func NewCompositeAuthorizer(authorizers ...Authorizer) Authorizer {
	return func(ctx context.Context, m Manager, token *oauth2.Token, user *User) error {
		for _, a := range authorizers {
			if err := a(ctx, m, token, user); err != nil {
				return err
			}
		}

		return nil
	}
}
