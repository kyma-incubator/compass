package onetimetoken

import (
	"bytes"
	"context"
	"encoding/base64"
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

//go:generate mockery --name=SystemAuthService --output=automock --outpkg=automock --case=underscore
type SystemAuthService interface {
	Create(ctx context.Context, objectType model.SystemAuthReferenceObjectType, objectID string, authInput *model.AuthInput) (string, error)
	GetByToken(ctx context.Context, token string) (*model.SystemAuth, error)
	GetGlobal(ctx context.Context, authID string) (*model.SystemAuth, error)
	Update(ctx context.Context, item *model.SystemAuth) error
}

//go:generate mockery --name=ApplicationConverter --output=automock --outpkg=automock --case=underscore
type ApplicationConverter interface {
	ToGraphQL(in *model.Application) *graphql.Application
}

//go:generate mockery --name=ApplicationService --output=automock --outpkg=automock --case=underscore
type ApplicationService interface {
	Get(ctx context.Context, id string) (*model.Application, error)
	ListLabels(ctx context.Context, applicationID string) (map[string]*model.Label, error)
}

//go:generate mockery --name=ExternalTenantsService --output=automock --outpkg=automock --case=underscore
type ExternalTenantsService interface {
	GetExternalTenant(ctx context.Context, id string) (string, error)
}

//go:generate mockery --name=HTTPDoer --output=automock --outpkg=automock --case=underscore
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type service struct {
	connectorURL              string
	legacyConnectorURL        string
	suggestTokenHeaderKey     string
	sysAuthSvc                SystemAuthService
	intSystemToAdapterMapping map[string]string
	appSvc                    ApplicationService
	appConverter              ApplicationConverter
	extTenantsSvc             ExternalTenantsService
	doer                      HTTPDoer
	tokenGenerator            TokenGenerator
	timeService               directorTime.Service
}

func NewTokenService(sysAuthSvc SystemAuthService, appSvc ApplicationService, appConverter ApplicationConverter, extTenantsSvc ExternalTenantsService, doer HTTPDoer, tokenGenerator TokenGenerator, config Config, intSystemToAdapterMapping map[string]string, timeService directorTime.Service) *service {
	return &service{
		connectorURL:              config.ConnectorURL,
		legacyConnectorURL:        config.LegacyConnectorURL,
		suggestTokenHeaderKey:     config.SuggestTokenHeaderKey,
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
	if tokenType == model.ApplicationReference {
		return s.getAppToken(ctx, id)
	}

	return s.createToken(ctx, id, tokenType, nil)
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

func (s *service) createToken(ctx context.Context, id string, tokenType model.SystemAuthReferenceObjectType, oneTimeToken *model.OneTimeToken) (*model.OneTimeToken, error) {
	if oneTimeToken == nil {
		ott, err := s.getNewToken()
		if err != nil {
			return nil, errors.Wrapf(err, "while generating onetime token for %s", tokenType)
		}

		oneTimeToken = ott
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

	sysAuthID, err := s.sysAuthSvc.Create(ctx, tokenType, id, &model.AuthInput{OneTimeToken: oneTimeToken})
	if err != nil {
		return nil, errors.Wrap(err, "while creating System Auth")
	}

	tokenPayloadJson, err := json.Marshal(model.TokenPayload{Token: oneTimeToken.Token, SystemAuthID: sysAuthID})
	if err != nil {
		return nil, errors.Wrap(err, "while creating Token Payload")
	}

	oneTimeToken.Token = base64.URLEncoding.EncodeToString(tokenPayloadJson)

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

func (s *service) getAppToken(ctx context.Context, id string) (*model.OneTimeToken, error) {
	var (
		oneTimeToken *model.OneTimeToken
		err          error
	)

	app, err := s.appSvc.Get(ctx, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting application [id: %s]", id)
	}

	if app.IntegrationSystemID != nil {
		if adapterURL, ok := s.intSystemToAdapterMapping[*app.IntegrationSystemID]; ok {
			oneTimeToken, err = s.getTokenFromAdapter(ctx, adapterURL, *app)
			if err != nil {
				return nil, errors.Wrapf(err, "while getting one time token for application from adapter with URL %s", adapterURL)
			}
		}
	}

	oneTimeToken, err = s.createToken(ctx, id, model.ApplicationReference, oneTimeToken)
	if err != nil {
		return nil, err
	}

	oneTimeToken.Token = s.getSuggestedTokenForApp(ctx, app, oneTimeToken)
	return oneTimeToken, nil
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

// getSuggestedTokenForApp returns the token that an application would use, depending on its type - if the application belongs to an integration
// system, then it would use the token as is, if the application is considered "legacy" - an application which still uses the "connectivity-adapter" -
// then it would use the legacy connector URL, then any new applications which are not managed by an integration system, would use the base64 encoded JSON
// containing the token, and the connector URL
func (s *service) getSuggestedTokenForApp(ctx context.Context, app *model.Application, oneTimeToken *model.OneTimeToken) string {
	if !tokenSuggestionEnabled(ctx, s.suggestTokenHeaderKey) {
		return oneTimeToken.Token
	}

	if app.IntegrationSystemID != nil {
		log.C(ctx).Infof("Application with ID %s belongs to an integration system, will use the actual token", app.ID)
		return oneTimeToken.Token
	}

	appLabels, err := s.appSvc.ListLabels(ctx, app.ID)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to check if application with ID %s is of legacy type, will use the actual token: %v", app.ID, err)
		return oneTimeToken.Token
	}

	if label, ok := appLabels["legacy"]; ok {
		if isLegacy, ok := (label.Value).(bool); ok && isLegacy {
			suggestedToken, err := legacyConnectorUrlWithToken(s.legacyConnectorURL, oneTimeToken.Token)
			if err != nil {
				log.C(ctx).WithError(err).Errorf("Failed to obtain legacy connector URL with token for application with ID %s, will use the actual token: %v", app.ID, err)
				return oneTimeToken.Token
			}

			log.C(ctx).Infof("Application with ID %s is of legacy type, will use legacy token URL", app.ID)
			return suggestedToken
		}
	}

	rawEnc, err := rawEncoded(&graphql.TokenWithURL{
		Token:        oneTimeToken.Token,
		ConnectorURL: oneTimeToken.ConnectorURL,
	})
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to generade raw encoded one time token for application, will continue with actual token: %v", err)
		return oneTimeToken.Token
	}

	return *rawEnc
}
