package onetimetoken

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

const requestForRuntime = `
		mutation { generateRuntimeToken (runtimeID:"%s")
		  {
			token
		  }
		}`

const requestForApplication = `
		mutation { generateApplicationToken (appID:"%s")
		  {
			token
		  }
		}`

//go:generate mockery -name=GraphQLClient -output=automock -outpkg=automock -case=underscore
type GraphQLClient interface {
	Run(ctx context.Context, req *gcli.Request, resp interface{}) error
}

//go:generate mockery -name=SystemAuthService -output=automock -outpkg=automock -case=underscore
type SystemAuthService interface {
	Create(ctx context.Context, objectType model.SystemAuthReferenceObjectType, objectID string, authInput *model.AuthInput) (string, error)
}
type service struct {
	cli          GraphQLClient
	connectorURL string
	sysAuthSvc   SystemAuthService
}

func NewTokenService(gcli GraphQLClient, sysAuthSvc SystemAuthService, connectorURL string) *service {
	return &service{cli: gcli, connectorURL: connectorURL, sysAuthSvc: sysAuthSvc}
}

func (s service) GenerateOneTimeToken(ctx context.Context, id string, tokenType model.SystemAuthReferenceObjectType) (model.OneTimeToken, error) {
	sysAuthID, err := s.sysAuthSvc.Create(ctx, tokenType, id, nil)
	if err != nil {
		return model.OneTimeToken{}, errors.Wrap(err, "while creating System Auth")
	}

	token, err := s.getOneTimeToken(ctx, sysAuthID, tokenType)
	if err != nil {
		return model.OneTimeToken{}, errors.Wrapf(err, "while generating onetime token for %s", tokenType)
	}

	return model.OneTimeToken{Token: token, ConnectorURL: s.connectorURL}, nil
}

func (s service) getOneTimeToken(ctx context.Context, id string, tokenType model.SystemAuthReferenceObjectType) (string, error) {
	var request *gcli.Request

	switch tokenType {
	case model.RuntimeReference:
		request = gcli.NewRequest(fmt.Sprintf(requestForRuntime, id))
	case model.ApplicationReference:
		request = gcli.NewRequest(fmt.Sprintf(requestForApplication, id))
	default:
		return "", errors.Errorf("cannot generate token for %T", tokenType)
	}

	output := ConnectorTokenModel{}
	err := s.cli.Run(ctx, request, &output)
	if err != nil {
		return "", errors.Wrapf(err, "while calling connector for %s one time token", tokenType)
	}

	return output.Token(tokenType), err
}
