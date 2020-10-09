package oauth

import (
	"context"
	"time"

	httputils "github.com/kyma-incubator/compass/components/system-broker/pkg/http"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/log"
)

func NewTokenProviderFromValue(t string) *TokenProviderFromValue {
	return &TokenProviderFromValue{
		token: t,
	}
}

type TokenProviderFromValue struct {
	token string
}

func (c *TokenProviderFromValue) GetAuthorizationToken(ctx context.Context) (httputils.Token, error) {
	tokenResponse := httputils.Token{
		AccessToken: c.token,
		Expiration:  9999,
	}

	log.C(ctx).Info("Successfully unmarshal response oauth token for accessing Director")
	tokenResponse.Expiration += time.Now().Unix()

	return tokenResponse, nil
}
