package token

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

var requestForRuntime = `
		mutation { generateRuntimeToken (runtimeID:"%s")
		  {
			token
		  }
		}`

var requestForApplication = `
		mutation { generateApplicationToken (appID:"%s")
		  {
			token
		  }
		}`

//go:generate mockery -name=GraphQLClient -output=automock -outpkg=automock -case=underscore
type GraphQLClient interface {
	Run(ctx context.Context, req *gcli.Request, resp interface{}) error
}

type service struct {
	cli          GraphQLClient
	connectorURL string
}

func NewTokenService(gcli GraphQLClient, connectorURL string) *service {
	return &service{cli: gcli, connectorURL: connectorURL}
}

func (s service) GenerateOneTimeToken(ctx context.Context, id string, tokenType TokenType) (model.OneTimeToken, error) {

	token, err := s.getOneTimeToken(ctx, id, tokenType)
	if err != nil {
		return model.OneTimeToken{}, errors.Wrap(err, "while generating onetime token for runtime")
	}

	//TODO: create empty entry in DB
	return model.OneTimeToken{Token: token, ConnectorURL: s.connectorURL}, nil
}

func (s service) getOneTimeToken(ctx context.Context, id string, tokenType TokenType) (string, error) {
	var request *gcli.Request
	switch tokenType {
	case RuntimeToken:
		request = gcli.NewRequest(fmt.Sprintf(requestForRuntime, id))
	case ApplicationToken:
		request = gcli.NewRequest(fmt.Sprintf(requestForApplication, id))
	}

	output := ExternalTokenModel{}
	err := s.cli.Run(ctx, request, &output)
	if err != nil {
		return "", errors.Wrap(err, "while calling connector for runtime one time token")
	}

	return output.Token(tokenType), err
}
