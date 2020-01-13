package director

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director/oauth"
	"github.com/pkg/errors"
)

type directorClient struct {
	token       oauth.Token
	oauthClient oauth.Client
}

func (cc *directorClient) getToken() error {
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
