package director

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director/oauth"
	"github.com/pkg/errors"
)

func NewDirectorClient(oauthClient oauth.Client) *directorClient {
	return &directorClient{
		oauthClient: oauthClient,
		token:       oauth.Token{},
	}
}

// TODO: wrap interface
type directorClient struct {
	oauthClient oauth.Client
	token       oauth.Token
}

func (cc *directorClient) setToken() error {
	token, err := cc.oauthClient.GetAuthorizationToken()
	if err != nil {
		return errors.Wrap(err, "Error while obtaining token")
	}

	if token.EmptyOrExpired() {
		return errors.New("Obtained empty or expired token")
	}

	cc.token = token
	return nil
}
