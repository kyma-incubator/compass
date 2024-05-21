package systemfielddiscoveryengine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/system-field-discovery-engine/config"
	pkgAuth "github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/pkg/errors"
)

// SystemFieldDiscoveryRegistry defines the system field discovery registry
type SystemFieldDiscoveryRegistry string

// ToString converts the system field discovery registry type to string
func (sfd SystemFieldDiscoveryRegistry) ToString() string {
	return string(sfd)
}

const (
	regionLabelKey = "region"

	// SystemFieldDiscoverySaaSRegistry system field discovery registry of type saas registry
	SystemFieldDiscoverySaaSRegistry SystemFieldDiscoveryRegistry = "saas-registry"
)

// subscription represents subscription object in a saas-manager response payload.
type subscription struct {
	AppURL string `json:"url"`
}

// subscriptionsResponse represents collection of all subscription objects in a saas-manager response payload.
type subscriptionsResponse struct {
	Subscriptions []subscription `json:"subscriptions"`
}

// ApplicationService is responsible for the service-layer Application operations.
//
//go:generate mockery --name=ApplicationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationService interface {
	UpdateBaseURLAndReadyState(ctx context.Context, appID, baseURL string, ready bool) error
	Get(ctx context.Context, id string) (*model.Application, error)
}

// ApplicationTemplateService is responsible for the service-layer ApplicationTemplate operations.
//
//go:generate mockery --name=ApplicationTemplateService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationTemplateService interface {
	GetLabel(ctx context.Context, appTemplateID string, key string) (*model.Label, error)
}

// TenantService is responsible for tenant operations
//
//go:generate mockery --name=TenantService --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantService interface {
	GetLowestOwnerForResource(ctx context.Context, resourceType resource.Type, objectID string) (string, error)
	GetTenantByID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
}

// SystemFieldDiscoveryEngineConfig is responsible for system field discovery configuration
//
//go:generate mockery --name=SystemFieldDiscoveryEngineConfig --output=automock --outpkg=automock --case=underscore --disable-version-string
type SystemFieldDiscoveryEngineConfig interface {
	PrepareConfiguration() (*config.SystemFieldDiscoveryEngineConfig, error)
}

// Client is responsible for making HTTP requests.
//
//go:generate mockery --name=Client --output=automock --outpkg=automock --case=underscore --disable-version-string
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

// Service consists of various resource services responsible for service-layer system field discovery engine operations.
type Service struct {
	cfg      config.SystemFieldDiscoveryEngineConfig
	client   Client
	transact persistence.Transactioner

	appSvc         ApplicationService
	appTemplateSvc ApplicationTemplateService
	tenantSvc      TenantService
}

// NewSystemFieldDiscoverEngineService returns a new object responsible for service-layer system field discovery engine operations.
func NewSystemFieldDiscoverEngineService(cfg SystemFieldDiscoveryEngineConfig, client Client, transact persistence.Transactioner, appSvc ApplicationService, appTemplateSvc ApplicationTemplateService, tenantSvc TenantService) (*Service, error) {
	conf, err := cfg.PrepareConfiguration()
	if err != nil {
		return nil, errors.Wrap(err, "while preparing system field discovery engine configuration")
	}
	return &Service{
		cfg:            *conf,
		transact:       transact,
		client:         client,
		appSvc:         appSvc,
		appTemplateSvc: appTemplateSvc,
		tenantSvc:      tenantSvc,
	}, nil
}

// ProcessSaasRegistryApplication triggers system field discovery engine for an app and tenant in saas registry
func (s *Service) ProcessSaasRegistryApplication(ctx context.Context, appID, tenantID string) error {
	ctx, err := s.saveLowestOwnerForAppToContextInTx(ctx, appID)
	if err != nil {
		return err
	}

	region, err := s.getRegionLabelInTx(ctx, appID)
	if err != nil {
		return errors.Wrapf(err, "retrieving label with key %q for application with id %q failed", regionLabelKey, appID)
	}

	if _, regionExists := s.cfg.RegionToSaasRegConfig[region]; !regionExists {
		return fmt.Errorf("region %q is not present into the saas reg configuration for application with id %q", region, appID)
	}

	url := fmt.Sprintf("%s/saas-manager/v1/service/subscriptions?includeIndirectSubscriptions=true&tenantId=%s", s.cfg.RegionToSaasRegConfig[region].SaasRegistryURL, tenantID)
	ctx = s.saveCredentialsToContext(ctx, region)
	respBody, err := executeCall(ctx, s.client, url)
	if err != nil {
		return errors.Wrapf(err, "failed executing request for url %q and appID %q", url, appID)
	}
	var response subscriptionsResponse
	if err = json.Unmarshal(respBody, &response); err != nil {
		log.C(ctx).Errorf(errors.Wrap(err, "failed to unmarshal subscriptions response").Error())
		return errors.Wrapf(err, "while unmarshaling subscription response")
	}
	for _, subscription := range response.Subscriptions {
		processed, err := s.processSubscription(ctx, subscription, url, appID)
		if err != nil {
			return errors.Wrapf(err, "failed processing subscription for url %q and app with id %q", url, appID)
		}
		if processed {
			log.C(ctx).Infof("Successfully processed url %q for app with id %q", url, appID)
			break
		}
		log.C(ctx).Infof("Response for url %q and does not contain app URL", url)
	}
	return nil
}

func (s *Service) getRegionLabelInTx(ctx context.Context, appID string) (string, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return "", err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	app, err := s.appSvc.Get(ctx, appID)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("error while getting applicationw with id %q", appID)
		return "", err
	}

	if app.ApplicationTemplateID == nil {
		return "", errors.Errorf("application with id %s does not have application template id", app.ID)
	}
	label, err := s.appTemplateSvc.GetLabel(ctx, *app.ApplicationTemplateID, regionLabelKey)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("error while getting label with key %q for applicationTemplate with ID %q", regionLabelKey, *app.ApplicationTemplateID)
		return "", err
	}
	regionValue, ok := label.Value.(string)
	if !ok {
		return "", errors.Errorf("%q label for applicationTemplate with ID %q is not a string", regionLabelKey, *app.ApplicationTemplateID)
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return regionValue, nil
}

func (s *Service) saveLowestOwnerForAppToContextInTx(ctx context.Context, appID string) (context.Context, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)
	ctx, err = s.saveLowestOwnerForAppToContext(ctx, appID)
	if err != nil {
		return nil, err
	}

	return ctx, tx.Commit()
}

func (s *Service) saveLowestOwnerForAppToContext(ctx context.Context, appID string) (context.Context, error) {
	internalTntID, err := s.tenantSvc.GetLowestOwnerForResource(ctx, resource.Application, appID)
	if err != nil {
		return nil, err
	}

	tnt, err := s.tenantSvc.GetTenantByID(ctx, internalTntID)
	if err != nil {
		return nil, err
	}

	ctx = tenant.SaveToContext(ctx, internalTntID, tnt.ExternalTenant)

	return ctx, nil
}

func (s *Service) saveCredentialsToContext(ctx context.Context, region string) context.Context {
	credentials := &pkgAuth.OAuthCredentials{
		ClientID:     s.cfg.RegionToSaasRegConfig[region].ClientID,
		ClientSecret: s.cfg.RegionToSaasRegConfig[region].ClientSecret,
		TokenURL:     s.cfg.RegionToSaasRegConfig[region].TokenURL + s.cfg.OauthTokenPath,
	}
	ctx = pkgAuth.SaveToContext(ctx, credentials)
	return ctx
}

func (s *Service) processSubscription(ctx context.Context, subscription subscription, url, appID string) (bool, error) {
	if subscription.AppURL == "" {
		return false, nil
	}

	tx, err := s.transact.Begin()
	if err != nil {
		return false, errors.Wrapf(err, "failed to begin a transaction for processing subscriptions for url %q and app with id %q", url, appID)
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	if err := s.appSvc.UpdateBaseURLAndReadyState(ctx, appID, subscription.AppURL, true); err != nil {
		return false, errors.Wrapf(err, "failed to update base url and ready state for app with id %q", appID)
	}

	return true, tx.Commit()
}

func executeCall(ctx context.Context, client Client, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "error while creating request for URL %q", url)
	}

	req = req.WithContext(ctx)
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "error while executing request for URL %q", url)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.C(ctx).Error(err, "Failed to close HTTP response body")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected status code: expected: %d, but got: %d", http.StatusOK, resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse HTTP response body")
	}

	return respBody, nil
}
