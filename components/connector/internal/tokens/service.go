package tokens

import (
	"context"
	"fmt"

	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/connector/internal/apperrors"
	"github.com/kyma-incubator/compass/components/connector/internal/authentication"
)

//go:generate mockery --name=GraphQLClient --output=automock --outpkg=automock --case=underscore --disable-version-string
type GraphQLClient interface {
	Run(ctx context.Context, req *gcli.Request, resp interface{}) error
}

//go:generate mockery --name=Service --output=automock --outpkg=automock --case=underscore --disable-version-string
type Service interface {
	GetToken(ctx context.Context, clientId, consumerType string) (string, apperrors.AppError)
}

var consumerTypeToQuery = map[string]string{
	"Application": `mutation {
		result: requestOneTimeTokenForApplication(id:"%s", systemAuthID:"%s") {
	      		token
		}
	}`,
	"Runtime": `mutation {
		result: requestOneTimeTokenForRuntime(id:"%s", systemAuthID:"%s") {
			token
		}
	}`,
}

type tokenService struct {
	cli GraphQLClient
}

func NewTokenService(cli GraphQLClient) *tokenService {
	return &tokenService{cli: cli}
}

func (svc *tokenService) GetToken(ctx context.Context, clientId, consumerType string) (string, apperrors.AppError) {
	token, err := svc.getOneTimeToken(ctx, clientId, consumerType)
	if err != nil {
		return "", apperrors.Internal("could not get one time token: %s", err)
	}

	return token, nil
}

func (svc *tokenService) getOneTimeToken(ctx context.Context, id, consumerType string) (string, error) {
	query, found := consumerTypeToQuery[consumerType]
	if !found {
		return "", errors.Errorf("No token generation for consumer type %s", consumerType)
	}
	req := gcli.NewRequest(fmt.Sprintf(query, "", id))

	tenant, err := authentication.GetStringFromContext(ctx, authentication.TenantKey)
	if err != nil {
		return "", errors.Wrap(err, "Failed to authenticate request, tenant not found")
	}
	req.Header.Set("Tenant", tenant)

	resp := TokenResponse{}
	if err := svc.cli.Run(ctx, req, &resp); err != nil {
		return "", errors.Wrapf(err, "while calling director for CSR one time token")
	}

	return resp.GetTokenValue(), nil
}
