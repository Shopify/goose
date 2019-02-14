package githuboauth

import (
	"context"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/Shopify/goose/oauth"
)

var ErrUserNotInOrg = errors.New("user not in required organization")

func NewOrgAuthorizer(org string) oauth.Authorizer {
	return func(ctx context.Context, manager oauth.Manager, token *oauth2.Token, user *oauth.User) error {
		client := github.NewClient(manager.GetClient(ctx, token))

		ok, _, err := client.Organizations.IsMember(ctx, org, user.Profile)
		if err != nil {
			return errors.Wrap(err, "error validating user")
		}
		if !ok {
			log(nil, nil).WithField("organization", org).Info(ErrUserNotInOrg.Error())
			return ErrUserNotInOrg
		}

		return nil
	}
}
