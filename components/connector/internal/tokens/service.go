package tokens

import (
	"context"
	"fmt"

	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/connector/internal/apperrors"
)

//go:generate mockery --name=GraphQLClient --output=automock --outpkg=automock --case=underscore
type GraphQLClient interface {
	Run(ctx context.Context, req *gcli.Request, resp interface{}) error
}

//go:generate mockery --name=Service --output=automock --outpkg=automock --case=underscore
type Service interface {
	GetToken(ctx context.Context, clientId string) (string, apperrors.AppError)
}

const requestForCSRToken = `
		mutation { generateCSRToken (authID:"%s")
		  {
			token
		  }
		}`

type tokenService struct {
	cli GraphQLClient
}

func NewTokenService(cli GraphQLClient) *tokenService {
	return &tokenService{cli: cli}
}

func (svc *tokenService) GetToken(ctx context.Context, clientId string) (string, apperrors.AppError) {
	token, err := svc.getOneTimeToken(ctx, clientId)
	if err != nil {
		return "", apperrors.Internal("could not get one time token: %s", err)
	}

	return token, nil
}

func (svc *tokenService) getOneTimeToken(ctx context.Context, id string) (string, error) {
	req := gcli.NewRequest(fmt.Sprintf(requestForCSRToken, id))

	resp := CSRTokenResponse{}
	if err := svc.cli.Run(ctx, req, &resp); err != nil {
		return "", errors.Wrapf(err, "while calling director for CSR one time token")
	}

	return resp.GetTokenValue(), nil
}
