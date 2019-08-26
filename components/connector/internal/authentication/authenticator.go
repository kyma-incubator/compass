package authentication

import (
	"context"

	"github.com/kyma-incubator/compass/components/connector/internal/tokens"
	"github.com/pkg/errors"
)

//go:generate mockery -name=Authenticator
type Authenticator interface {
	AuthenticateToken(context context.Context) (tokens.TokenData, error)
}

func NewAuthenticator(tokenService tokens.Service) Authenticator {
	return &authenticator{
		tokenService: tokenService,
	}
}

type authenticator struct {
	tokenService tokens.Service
}

func (a *authenticator) AuthenticateToken(context context.Context) (tokens.TokenData, error) {
	token, err := GetStringFromContext(context, ConnectorTokenKey)
	if err != nil {
		return tokens.TokenData{}, errors.Wrap(err, "Failed to authenticate request, token not provided")
	}

	tokenData, err := a.tokenService.Resolve(token)
	if err != nil {
		return tokens.TokenData{}, errors.Wrap(err, "Failed to authenticate request, token is invalid")
	}

	return tokenData, nil
}
