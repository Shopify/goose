package githuboauth

import (
	"context"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/Shopify/goose/logger"
	"github.com/Shopify/goose/oauth"
)

var log = logger.New("oauth")

func Authenticator(ctx context.Context, m oauth.Manager, token *oauth2.Token) (*oauth.User, error) {
	client := github.NewClient(m.GetClient(ctx, token))

	ghUser, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, errors.Wrap(err, "unable to verify user")
	}

	user := convertUser(ghUser)

	return user, nil
}

func convertUser(user *github.User) *oauth.User {
	return &oauth.User{
		Profile:       user.GetLogin(),
		Name:          user.GetName(),
		Email:         user.GetEmail(),
		EmailVerified: true,
		Picture:       user.GetAvatarURL(),
	}
}
