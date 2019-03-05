package googleoauth

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/Shopify/goose/logger"
	"github.com/Shopify/goose/oauth"
)

var log = logger.New("oauth")

func Authenticator(ctx context.Context, m oauth.Manager, token *oauth2.Token) (*oauth.User, error) {
	client := m.GetClient(ctx, token)

	user, err := getUserInfo(ctx, client)
	if err != nil {
		return nil, errors.Wrap(err, "unable to verify user")
	}

	return user, nil
}

func getUserInfo(ctx context.Context, client *http.Client) (user *oauth.User, err error) {
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		log(ctx, err).Warn("error querying google api for user info")
		return nil, err
	}

	defer func() { err = resp.Body.Close() }()
	jsonData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log(ctx, err).Warn("error reading body of user info")
		return nil, err
	}

	user = &oauth.User{}
	if err := json.Unmarshal(jsonData, user); err != nil {
		log(ctx, err).Warn("error decoding body of user info")
		return nil, err
	}

	return user, nil
}
