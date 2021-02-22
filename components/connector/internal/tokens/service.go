package tokens

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	gcli "github.com/machinebox/graphql"

	"github.com/kyma-incubator/compass/components/connector/internal/apperrors"
)

//go:generate mockery -name=GraphQLClient -output=automock -outpkg=automock -case=underscore
type GraphQLClient interface {
	Run(ctx context.Context, req *gcli.Request, resp interface{}) error
}

//go:generate mockery -name=Service
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

func (s *tokenService) getOneTimeToken(ctx context.Context, id string) (string, error) {
	req := gcli.NewRequest(fmt.Sprintf(requestForCSRToken, id))

	// todo refactor this
	var resp map[string]map[string]interface{}
	err := s.cli.Run(ctx, req, &resp)
	if err != nil {
		return "", errors.Wrapf(err, "while calling director for CSR one time token")
	}

	return resp["generateCSRToken"]["token"].(string), nil
}
