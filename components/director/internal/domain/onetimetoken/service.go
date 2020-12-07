package onetimetoken

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/domain/client"

	"github.com/avast/retry-go"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/pairing"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

const requestForRuntime = `
		mutation { generateRuntimeToken (authID:"%s")
		  {
			token
		  }
		}`

const requestForApplication = `
		mutation { generateApplicationToken (authID:"%s")
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

//go:generate mockery -name=ApplicationConverter -output=automock -outpkg=automock -case=underscore
type ApplicationConverter interface {
	ToGraphQL(in *model.Application) *graphql.Application
}

//go:generate mockery -name=ApplicationService -output=automock -outpkg=automock -case=underscore
type ApplicationService interface {
	Get(ctx context.Context, id string) (*model.Application, error)
}

//go:generate mockery -name=ExternalTenantsService -output=automock -outpkg=automock -case=underscore
type ExternalTenantsService interface {
	GetExternalTenant(ctx context.Context, id string) (string, error)
}

//go:generate mockery -name=HTTPDoer -output=automock -outpkg=automock -case=underscore
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type service struct {
	cli                       GraphQLClient
	connectorURL              string
	sysAuthSvc                SystemAuthService
	intSystemToAdapterMapping map[string]string
	appSvc                    ApplicationService
	appConverter              ApplicationConverter
	extTenantsSvc             ExternalTenantsService
	doer                      HTTPDoer
}

func NewTokenService(gcli GraphQLClient, sysAuthSvc SystemAuthService, appSvc ApplicationService, appConverter ApplicationConverter, extTenantsSvc ExternalTenantsService, doer HTTPDoer, connectorURL string, intSystemToAdapterMapping map[string]string) *service {
	return &service{cli: gcli, connectorURL: connectorURL, sysAuthSvc: sysAuthSvc, intSystemToAdapterMapping: intSystemToAdapterMapping, appSvc: appSvc, appConverter: appConverter, extTenantsSvc: extTenantsSvc, doer: doer}
}

func (s service) GenerateOneTimeToken(ctx context.Context, id string, tokenType model.SystemAuthReferenceObjectType) (model.OneTimeToken, error) {
	sysAuthID, err := s.sysAuthSvc.Create(ctx, tokenType, id, nil)
	if err != nil {
		return model.OneTimeToken{}, errors.Wrap(err, "while creating System Auth")
	}

	if tokenType == model.ApplicationReference {
		app, err := s.appSvc.Get(ctx, id)

		if err != nil {
			return model.OneTimeToken{}, errors.Wrapf(err, "while getting application [id: %s]", id)
		}

		if app.IntegrationSystemID != nil {
			if adapterURL, ok := s.intSystemToAdapterMapping[*app.IntegrationSystemID]; ok {
				return s.getTokenFromAdapter(ctx, adapterURL, *app)
			}
		}
	}

	token, err := s.getOneTimeToken(ctx, sysAuthID, tokenType)
	if err != nil {
		return model.OneTimeToken{}, errors.Wrapf(err, "while generating onetime token for %s", tokenType)
	}

	return model.OneTimeToken{Token: token, ConnectorURL: s.connectorURL}, nil
}

func (s *service) getTokenFromAdapter(ctx context.Context, adapterURL string, app model.Application) (model.OneTimeToken, error) {
	extTenant, err := s.extTenantsSvc.GetExternalTenant(ctx, app.Tenant)
	if err != nil {
		return model.OneTimeToken{}, errors.Wrapf(err, "while getting external tenant for internal tenant [%s]", app.Tenant)
	}

	clientUser, err := client.LoadFromContext(ctx)
	if err != nil {
		log.C(ctx).Infof("unable to provide client_user for internal tenant [%s] with corresponding external tenant [%s]", app.Tenant, extTenant)
	}

	graphqlApp := s.appConverter.ToGraphQL(&app)
	data := pairing.RequestData{
		Application: *graphqlApp,
		Tenant:      extTenant,
		ClientUser:  clientUser,
	}

	asJSON, err := json.Marshal(data)
	if err != nil {
		return model.OneTimeToken{}, errors.Wrap(err, "while marshaling data for adapter")
	}

	var externalToken string
	err = retry.Do(func() error {
		buf := bytes.NewBuffer(asJSON)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, adapterURL, buf)
		if err != nil {
			return errors.Wrap(err, "while creating request")
		}

		resp, err := s.doer.Do(req)
		if err != nil {
			return errors.Wrap(err, "while executing request")
		}

		defer func() {
			err := resp.Body.Close()
			if err != nil {
				log.C(ctx).Warnf("Got error on closing response body: [%v]", err)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("wrong status code, got [%d], expected [%d]", resp.StatusCode, http.StatusOK)
		}

		responseBody := pairing.ResponseData{}
		if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
			return errors.Wrap(err, "while decoding response from Adapter")
		}

		externalToken = responseBody.Token
		return nil
	}, retry.Attempts(3))
	if err != nil {
		return model.OneTimeToken{}, errors.Wrapf(err, "while calling adapter [%s] for application [%s] with integration system [%s]", adapterURL, app.ID, *app.IntegrationSystemID)
	}
	return model.OneTimeToken{
		Token: externalToken,
	}, nil
}

func (s service) getOneTimeToken(ctx context.Context, id string, tokenType model.SystemAuthReferenceObjectType) (string, error) {
	var req *gcli.Request

	switch tokenType {
	case model.RuntimeReference:
		req = gcli.NewRequest(fmt.Sprintf(requestForRuntime, id))
	case model.ApplicationReference:
		req = gcli.NewRequest(fmt.Sprintf(requestForApplication, id))
	default:
		return "", errors.Errorf("cannot generate token for %T", tokenType)
	}

	output := ConnectorTokenModel{}
	err := s.cli.Run(ctx, req, &output)
	if err != nil {
		return "", errors.Wrapf(err, "while calling connector for %s one time token", tokenType)
	}

	return output.Token(tokenType), err
}
