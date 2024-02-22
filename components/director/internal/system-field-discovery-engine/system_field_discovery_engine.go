package systemfielddiscoveryengine

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
)

const (
	// RegistryLabelKey is the label key for registry label
	RegistryLabelKey = "registry"
	// SaaSRegistryLabelValue is the label value for saas registry label
	SaaSRegistryLabelValue = "saas-registry"
)

// LabelService is responsible updating already existing labels, and their label definitions.
//
//go:generate mockery --name=LabelService --output=automock --outpkg=automock --case=underscore --disable-version-string
type LabelService interface {
	CreateLabel(ctx context.Context, tenant, id string, labelInput *model.LabelInput) error
}

// WebhookService is responsible for the service-layer Webhook operations.
//
//go:generate mockery --name=WebhookService --output=automock --outpkg=automock --case=underscore --disable-version-string
type WebhookService interface {
	GetByIDAndWebhookTypeGlobal(ctx context.Context, objectID string, objectType model.WebhookReferenceObjectType, webhookType model.WebhookType) (*model.Webhook, error)
}

//go:generate mockery --exported --name=uidService --output=automock --outpkg=automock --case=underscore --disable-version-string
type uidService interface {
	Generate() string
}

type systemFieldDiscoveryEngine struct {
	cfg        config.SystemFieldDiscoveryEngineConfig
	labelSvc   LabelService
	webhookSvc WebhookService
	uidSvc     uidService
}

// NewSystemFieldDiscoveryEngine creates a new SystemFieldDiscoveryEngine which is responsible for doing system field discovery engine operations configured with values from cfg.
func NewSystemFieldDiscoveryEngine(cfg config.SystemFieldDiscoveryEngineConfig, labelSvc LabelService, webhookSvc WebhookService, uidSvc uidService) (*systemFieldDiscoveryEngine, error) {
	if err := cfg.PrepareConfiguration(); err != nil {
		return nil, errors.Wrap(err, "while preparing system field discovery engine configuration")
	}
	return &systemFieldDiscoveryEngine{
		cfg:        cfg,
		labelSvc:   labelSvc,
		webhookSvc: webhookSvc,
		uidSvc:     uidSvc,
	}, nil
}

// EnrichApplicationWebhookIfNeeded enriches application webhook input with webhook of type 'SYSTEM_FIELD_DISCOVERY' if needed
func (s *systemFieldDiscoveryEngine) EnrichApplicationWebhookIfNeeded(ctx context.Context, appCreateInputModel model.ApplicationRegisterInput, systemFieldDiscovery bool, region, subacountID, appTemplateName, appName string) ([]*model.WebhookInput, bool) {
	if systemFieldDiscovery {
		log.C(ctx).Infof("Application Template with name %q has label systemFieldDiscovery with value %t. Enriching the application with name %q with webhook of type %q", appTemplateName, systemFieldDiscovery, appName, model.WebhookTypeSystemFieldDiscovery)
		appCreateInputModel.Webhooks = s.enrichWithWebhook(appCreateInputModel.Webhooks, region, subacountID)
		log.C(ctx).Infof("Successfully enriched Application with name %q with webhook of type %q", appName, model.WebhookTypeSystemFieldDiscovery)
	}

	return appCreateInputModel.Webhooks, systemFieldDiscovery
}

func (s *systemFieldDiscoveryEngine) enrichWithWebhook(modelInputWebhooks []*model.WebhookInput, region, subaccountID string) []*model.WebhookInput {
	modelInputWebhooks = append(modelInputWebhooks, &model.WebhookInput{
		Type: model.WebhookTypeSystemFieldDiscovery,
		URL:  str.Ptr(fmt.Sprintf("%s/saas-manager/v1/service/subscriptions?includeIndirectSubscriptions=true&tenantId=%s", s.cfg.RegionToSaasRegConfig[region].SaasRegistryURL, subaccountID)),
		Auth: &model.AuthInput{
			Credential: &model.CredentialDataInput{
				Oauth: &model.OAuthCredentialDataInput{
					ClientID:     s.cfg.RegionToSaasRegConfig[region].ClientID,
					ClientSecret: s.cfg.RegionToSaasRegConfig[region].ClientSecret,
					URL:          s.cfg.RegionToSaasRegConfig[region].TokenURL + s.cfg.OauthTokenPath,
				},
			},
		},
	})
	return modelInputWebhooks
}

// CreateLabelForApplicationWebhook creates label for webhook for application with id
func (s *systemFieldDiscoveryEngine) CreateLabelForApplicationWebhook(ctx context.Context, appID string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	wh, err := s.webhookSvc.GetByIDAndWebhookTypeGlobal(ctx, appID, model.ApplicationWebhookReference, model.WebhookTypeSystemFieldDiscovery)
	if err != nil {
		return err
	}
	log.C(ctx).Infof("Creating label with key: %q and value: %q for %q with id: %q", RegistryLabelKey, SaaSRegistryLabelValue, model.WebhookLabelableObject, wh.ID)
	if err := s.labelSvc.CreateLabel(ctx, tnt, s.uidSvc.Generate(), &model.LabelInput{
		Key:        RegistryLabelKey,
		Value:      SaaSRegistryLabelValue,
		ObjectID:   wh.ID,
		ObjectType: model.WebhookLabelableObject,
	}); err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while creating label with key: %q and value: %q for object type: %q and ID: %q", RegistryLabelKey, SaaSRegistryLabelValue, model.WebhookLabelableObject, wh.ID)
		return errors.Wrapf(err, "while creating label with key: %q and value: %q for object type: %q and ID: %q", RegistryLabelKey, SaaSRegistryLabelValue, model.WebhookLabelableObject, wh.ID)
	}
	log.C(ctx).Infof("Successfully created label with key: %q and value: %q for %q with id: %q", RegistryLabelKey, SaaSRegistryLabelValue, model.WebhookLabelableObject, wh.ID)

	return nil
}
