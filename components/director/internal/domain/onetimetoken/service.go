package onetimetoken

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/avast/retry-go"
	"github.com/kyma-incubator/compass/components/director/internal/domain/client"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tokens"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/pairing"
	directorTime "github.com/kyma-incubator/compass/components/director/pkg/time"
	"github.com/pkg/errors"
)

//go:generate mockery -name=SystemAuthService -output=automock -outpkg=automock -case=underscore
type SystemAuthService interface {
	Create(ctx context.Context, objectType model.SystemAuthReferenceObjectType, objectID string, authInput *model.AuthInput) (string, error)
	GetByToken(ctx context.Context, token string) (*model.SystemAuth, error)
	GetGlobal(ctx context.Context, authID string) (*model.SystemAuth, error)
	Update(ctx context.Context, item *model.SystemAuth) error
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
	connectorURL              string
	sysAuthSvc                SystemAuthService
	intSystemToAdapterMapping map[string]string
	appSvc                    ApplicationService
	appConverter              ApplicationConverter
	extTenantsSvc             ExternalTenantsService
	doer                      HTTPDoer
	tokenGenerator            TokenGenerator
	timeService               directorTime.Service
}

func NewTokenService(sysAuthSvc SystemAuthService, appSvc ApplicationService, appConverter ApplicationConverter, extTenantsSvc ExternalTenantsService, doer HTTPDoer, tokenGenerator TokenGenerator, connectorURL string, intSystemToAdapterMapping map[string]string, timeService directorTime.Service) *service {
	return &service{
		connectorURL:              connectorURL,
		sysAuthSvc:                sysAuthSvc,
		intSystemToAdapterMapping: intSystemToAdapterMapping,
		appSvc:                    appSvc,
		appConverter:              appConverter,
		extTenantsSvc:             extTenantsSvc,
		doer:                      doer,
		tokenGenerator:            tokenGenerator,
		timeService:               timeService,
	}
}

func (s service) GenerateOneTimeToken(ctx context.Context, id string, tokenType model.SystemAuthReferenceObjectType) (*model.OneTimeToken, error) {
	var oneTimeToken *model.OneTimeToken
	var err error
	if tokenType == model.ApplicationReference {
		app, err := s.appSvc.Get(ctx, id)
		if err != nil {
			return &model.OneTimeToken{}, errors.Wrapf(err, "while getting application [id: %s]", id)
		}

		if app.IntegrationSystemID != nil {
			if adapterURL, ok := s.intSystemToAdapterMapping[*app.IntegrationSystemID]; ok {
				oneTimeToken, err = s.getTokenFromAdapter(ctx, adapterURL, *app)
				if err != nil {
					return &model.OneTimeToken{}, errors.Wrapf(err, "while getting one time token for %s from adapter with URL %s", tokenType, adapterURL)
				}
			}
		}
	}

	if oneTimeToken == nil {
		if oneTimeToken, err = s.getNewToken(); err != nil {
			return nil, errors.Wrapf(err, "while generating onetime token for %s", tokenType)
		}
	}

	switch tokenType {
	case model.ApplicationReference:
		oneTimeToken.Type = tokens.ApplicationToken
	case model.RuntimeReference:
		oneTimeToken.Type = tokens.RuntimeToken
	}
	oneTimeToken.CreatedAt = s.timeService.Now()
	oneTimeToken.Used = false
	oneTimeToken.UsedAt = time.Time{}

	if _, err = s.sysAuthSvc.Create(ctx, tokenType, id, &model.AuthInput{OneTimeToken: oneTimeToken}); err != nil {
		return &model.OneTimeToken{}, errors.Wrap(err, "while creating System Auth")
	}

	return oneTimeToken, nil
}

func (s *service) getNewToken() (*model.OneTimeToken, error) {
	tokenString, err := s.tokenGenerator.NewToken()
	if err != nil {
		return nil, err
	}
	return &model.OneTimeToken{
		Token:        tokenString,
		ConnectorURL: s.connectorURL,
	}, nil
}

func (s *service) RegenerateOneTimeToken(ctx context.Context, sysAuthID string, tokenType tokens.TokenType) (model.OneTimeToken, error) {
	sysAuth, err := s.sysAuthSvc.GetGlobal(ctx, sysAuthID)
	if err != nil {
		return model.OneTimeToken{}, err
	}
	if sysAuth.Value == nil {
		sysAuth.Value = &model.Auth{}
	}

	tokenString, err := s.tokenGenerator.NewToken()
	if err != nil {
		return model.OneTimeToken{}, errors.Wrapf(err, "while generating onetime token")
	}
	oneTimeToken := &model.OneTimeToken{
		Token:        tokenString,
		ConnectorURL: s.connectorURL,
		Type:         tokenType,
		CreatedAt:    s.timeService.Now(),
		Used:         false,
		UsedAt:       time.Time{},
	}

	sysAuth.Value.OneTimeToken = oneTimeToken
	if err := s.sysAuthSvc.Update(ctx, sysAuth); err != nil {
		return model.OneTimeToken{}, err
	}

	return *oneTimeToken, nil
}

func (s *service) getTokenFromAdapter(ctx context.Context, adapterURL string, app model.Application) (*model.OneTimeToken, error) {
	extTenant, err := s.extTenantsSvc.GetExternalTenant(ctx, app.Tenant)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting external tenant for internal tenant [%s]", app.Tenant)
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
		return nil, errors.Wrap(err, "while marshaling data for adapter")
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
		return nil, errors.Wrapf(err, "while calling adapter [%s] for application [%s] with integration system [%s]", adapterURL, app.ID, *app.IntegrationSystemID)
	}
	return &model.OneTimeToken{
		Token: externalToken,
	}, nil
}
