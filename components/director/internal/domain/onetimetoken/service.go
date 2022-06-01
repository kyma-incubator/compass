package onetimetoken

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	pkgadapters "github.com/kyma-incubator/compass/components/director/pkg/adapters"

	pkgmodel "github.com/kyma-incubator/compass/components/director/pkg/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	tenantpkg "github.com/kyma-incubator/compass/components/director/pkg/tenant"

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

// SystemAuthService missing godoc
//go:generate mockery --name=SystemAuthService --output=automock --outpkg=automock --case=underscore --disable-version-string
type SystemAuthService interface {
	Create(ctx context.Context, objectType pkgmodel.SystemAuthReferenceObjectType, objectID string, authInput *model.AuthInput) (string, error)
	GetByToken(ctx context.Context, token string) (*pkgmodel.SystemAuth, error)
	GetGlobal(ctx context.Context, authID string) (*pkgmodel.SystemAuth, error)
	Update(ctx context.Context, item *pkgmodel.SystemAuth) error
}

// ApplicationConverter missing godoc
//go:generate mockery --name=ApplicationConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationConverter interface {
	ToGraphQL(in *model.Application) *graphql.Application
}

// ApplicationService missing godoc
//go:generate mockery --name=ApplicationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationService interface {
	Get(ctx context.Context, id string) (*model.Application, error)
	ListLabels(ctx context.Context, applicationID string) (map[string]*model.Label, error)
}

// ExternalTenantsService missing godoc
//go:generate mockery --name=ExternalTenantsService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ExternalTenantsService interface {
	GetExternalTenant(ctx context.Context, id string) (string, error)
	GetTenantByExternalID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
	GetTenantByID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
}

// HTTPDoer missing godoc
//go:generate mockery --name=HTTPDoer --output=automock --outpkg=automock --case=underscore --disable-version-string
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type service struct {
	connectorURL           string
	legacyConnectorURL     string
	suggestTokenHeaderKey  string
	csrTokenExpiration     time.Duration
	appTokenExpiration     time.Duration
	runtimeTokenExpiration time.Duration
	sysAuthSvc             SystemAuthService
	pairingAdapters        *pkgadapters.Adapters
	appSvc                 ApplicationService
	appConverter           ApplicationConverter
	extTenantsSvc          ExternalTenantsService
	doer                   HTTPDoer
	tokenGenerator         TokenGenerator
	timeService            directorTime.Service
}

// NewTokenService missing godoc
func NewTokenService(sysAuthSvc SystemAuthService, appSvc ApplicationService, appConverter ApplicationConverter, extTenantsSvc ExternalTenantsService, doer HTTPDoer, tokenGenerator TokenGenerator, config Config, pairingAdapters *pkgadapters.Adapters, timeService directorTime.Service) *service {
	return &service{
		connectorURL:           config.ConnectorURL,
		legacyConnectorURL:     config.LegacyConnectorURL,
		suggestTokenHeaderKey:  config.SuggestTokenHeaderKey,
		csrTokenExpiration:     config.CSRExpiration,
		appTokenExpiration:     config.ApplicationExpiration,
		runtimeTokenExpiration: config.RuntimeExpiration,
		sysAuthSvc:             sysAuthSvc,
		pairingAdapters:        pairingAdapters,
		appSvc:                 appSvc,
		appConverter:           appConverter,
		extTenantsSvc:          extTenantsSvc,
		doer:                   doer,
		tokenGenerator:         tokenGenerator,
		timeService:            timeService,
	}
}

// GenerateOneTimeToken missing godoc
func (s *service) GenerateOneTimeToken(ctx context.Context, objectID string, tokenType pkgmodel.SystemAuthReferenceObjectType) (*model.OneTimeToken, error) {
	token, suggestedToken, err := s.getToken(ctx, objectID, tokenType)
	if err != nil {
		return nil, err
	}
	if err := s.saveToken(ctx, objectID, tokenType, token); err != nil {
		return nil, err
	}
	if suggestedToken != "" {
		token.Token = suggestedToken
	}

	return token, nil
}

// RegenerateOneTimeToken missing godoc
func (s *service) RegenerateOneTimeToken(ctx context.Context, sysAuthID string) (*model.OneTimeToken, error) {
	sysAuth, err := s.sysAuthSvc.GetGlobal(ctx, sysAuthID)
	if err != nil {
		return nil, err
	}
	objectID, err := sysAuth.GetReferenceObjectID()
	if err != nil {
		return nil, err
	}
	if sysAuth.Value == nil {
		sysAuth.Value = &model.Auth{}
	}
	tokenType, err := sysAuth.GetReferenceObjectType()
	if err != nil {
		return nil, err
	}
	oneTimeToken, _, err := s.getToken(ctx, objectID, tokenType)
	if err != nil {
		return nil, err
	}

	sysAuth.Value.OneTimeToken = oneTimeToken
	if err := s.sysAuthSvc.Update(ctx, sysAuth); err != nil {
		return nil, err
	}

	return oneTimeToken, nil
}

func (s *service) getToken(ctx context.Context, objectID string, tokenType pkgmodel.SystemAuthReferenceObjectType) (*model.OneTimeToken, string, error) {
	if tokenType == pkgmodel.ApplicationReference {
		log.C(ctx).Infof("Getting one time token for %s with ID: %s...", tokenType, objectID)
		return s.getAppToken(ctx, objectID)
	} else {
		token, err := s.createToken(ctx, tokenType, objectID, nil)
		return token, "", err
	}
}

func (s *service) createToken(ctx context.Context, tokenType pkgmodel.SystemAuthReferenceObjectType, objectID string, oneTimeToken *model.OneTimeToken) (*model.OneTimeToken, error) {
	var err error
	if oneTimeToken == nil {
		log.C(ctx).Infof("Creating one time token for %s with ID: %s...", tokenType, objectID)
		oneTimeToken, err = s.getNewToken()
		if err != nil {
			return nil, errors.Wrapf(err, "while generating onetime token for %s", tokenType)
		}
	}

	log.C(ctx).Infof("Updating one time token with metadata for %s with ID: %s...", tokenType, objectID)

	switch tokenType {
	case pkgmodel.ApplicationReference:
		oneTimeToken.Type = tokens.ApplicationToken
	case pkgmodel.RuntimeReference:
		oneTimeToken.Type = tokens.RuntimeToken
	}
	oneTimeToken.CreatedAt = s.timeService.Now()
	oneTimeToken.Used = false
	oneTimeToken.UsedAt = time.Time{}
	expiresAfter, err := s.getExpirationDurationForToken(oneTimeToken.Type)
	if err != nil {
		return nil, err
	}
	oneTimeToken.ExpiresAt = oneTimeToken.CreatedAt.Add(expiresAfter)

	return oneTimeToken, nil
}

func (s *service) saveToken(ctx context.Context, objectID string, tokenType pkgmodel.SystemAuthReferenceObjectType, oneTimeToken *model.OneTimeToken) error {
	if _, err := s.sysAuthSvc.Create(ctx, tokenType, objectID, &model.AuthInput{OneTimeToken: oneTimeToken}); err != nil {
		return errors.Wrap(err, "while creating System Auth")
	}
	return nil
}

func (s *service) getAppToken(ctx context.Context, id string) (*model.OneTimeToken, string, error) {
	var (
		oneTimeToken *model.OneTimeToken
		err          error
	)

	app, err := s.appSvc.Get(ctx, id)
	if err != nil {
		return nil, "", errors.Wrapf(err, "while getting application [id: %s]", id)
	}

	if app.IntegrationSystemID != nil {
		intSystemToAdapterMapping := map[string]string{}
		if s.pairingAdapters != nil {
			intSystemToAdapterMapping = s.pairingAdapters.Get()
			if intSystemToAdapterMapping == nil {
				log.C(ctx).Error("pairing adapter configuration mapping cannot be nil")
				return nil, "", errors.Errorf("pairing adapter configuration mapping cannot be nil")
			}
		}

		if adapterURL, ok := intSystemToAdapterMapping[*app.IntegrationSystemID]; ok {
			log.C(ctx).Infof("Getting one time token for application with name: %s and ID: %s from pairing adapter...", app.Name, app.ID)
			oneTimeToken, err = s.getTokenFromAdapter(ctx, adapterURL, *app)
			if err != nil {
				return nil, "", errors.Wrapf(err, "while getting one time token for application from adapter with URL %s", adapterURL)
			}
		}
		log.C(ctx).Warnf("Could not find any adapter for the given integration system ID: %s", *app.IntegrationSystemID)
	}

	oneTimeToken, err = s.createToken(ctx, pkgmodel.ApplicationReference, id, oneTimeToken)
	if err != nil {
		return nil, "", err
	}

	suggestedAppTokenString := s.getSuggestedTokenForApp(ctx, app, oneTimeToken)
	return oneTimeToken, suggestedAppTokenString, nil
}

func (s *service) getTokenFromAdapter(ctx context.Context, adapterURL string, app model.Application) (*model.OneTimeToken, error) {
	tntCtx, err := tenant.LoadTenantPairFromContext(ctx)
	if err != nil {
		return nil, err
	}

	extTenant := tntCtx.ExternalID
	var tnt *model.BusinessTenantMapping
	if len(extTenant) == 0 {
		if tnt, err = s.extTenantsSvc.GetTenantByID(ctx, tntCtx.InternalID); err != nil {
			return nil, errors.Wrapf(err, "while getting tenant with internal ID %q", tntCtx.InternalID)
		}
	} else {
		if tnt, err = s.extTenantsSvc.GetTenantByExternalID(ctx, tntCtx.ExternalID); err != nil {
			return nil, errors.Wrapf(err, "while getting tenant with external ID %q", tntCtx.ExternalID)
		}
	}

	extTenant = tnt.ExternalTenant
	if tnt.Type == tenantpkg.Subaccount {
		if extTenant, err = s.extTenantsSvc.GetExternalTenant(ctx, tnt.Parent); err != nil {
			return nil, errors.Wrapf(err, "while getting parent external tenant for internal tenant %q", tnt.Parent)
		}
	}

	clientUser, err := client.LoadFromContext(ctx)
	if err != nil {
		log.C(ctx).Infof("unable to provide client_user for internal tenant [%s] with corresponding external tenant [%s]", tntCtx.InternalID, extTenant)
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

	log.C(ctx).Infof("Getting one time token from pairing adapter with URL: %s", adapterURL)
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
			suggestedToken, err := legacyConnectorURLWithToken(s.legacyConnectorURL, oneTimeToken.Token)
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
		Used:         oneTimeToken.Used,
		Type:         graphql.OneTimeTokenTypeApplication,
		ExpiresAt:    (*graphql.Timestamp)(&oneTimeToken.ExpiresAt),
		CreatedAt:    (*graphql.Timestamp)(&oneTimeToken.CreatedAt),
		UsedAt:       (*graphql.Timestamp)(&oneTimeToken.UsedAt),
	})
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to generade raw encoded one time token for application, will continue with actual token: %v", err)
		return oneTimeToken.Token
	}

	return *rawEnc
}

func (s *service) getExpirationDurationForToken(tokenType tokens.TokenType) (time.Duration, error) {
	switch tokenType {
	case tokens.ApplicationToken:
		return s.appTokenExpiration, nil
	case tokens.RuntimeToken:
		return s.runtimeTokenExpiration, nil
	default:
		return time.Duration(0), errors.Errorf("%s is no valid token type", tokenType)
	}
}

func (s *service) IsTokenValid(systemAuth *pkgmodel.SystemAuth) (bool, error) {
	if systemAuth.Value == nil {
		return false, errors.Errorf("System Auth value for auth id %s is missing", systemAuth.ID)
	}

	if systemAuth.Value.OneTimeToken == nil {
		return false, errors.Errorf("One Time Token for system auth id %s is missing", systemAuth.ID)
	}

	if systemAuth.Value.OneTimeToken.Used {
		return false, errors.Errorf("One Time Token for system auth id %s has been used", systemAuth.ID)
	}

	expirationTime, err := s.getExpirationDurationForToken(systemAuth.Value.OneTimeToken.Type)
	if err != nil {
		return false, errors.Wrapf(err, "one-time token for system auth id %s has no valid expiration type", systemAuth.ID)
	}

	isExpired := systemAuth.Value.OneTimeToken.CreatedAt.Add(expirationTime).Before(s.timeService.Now())
	if isExpired {
		return false, errors.Errorf("One Time Token with validity %s for system auth with ID %s has expired", expirationTime.String(), systemAuth.ID)
	}

	return true, nil
}
