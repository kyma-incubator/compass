package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

// WebhookRepository missing godoc
//
//go:generate mockery --name=WebhookRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type WebhookRepository interface {
	GetByID(ctx context.Context, tenant, id string, objectType model.WebhookReferenceObjectType) (*model.Webhook, error)
	GetByIDGlobal(ctx context.Context, id string) (*model.Webhook, error)
	ListByReferenceObjectID(ctx context.Context, tenant, objID string, objType model.WebhookReferenceObjectType) ([]*model.Webhook, error)
	ListByReferenceObjectIDGlobal(ctx context.Context, objID string, objType model.WebhookReferenceObjectType) ([]*model.Webhook, error)
	ListByWebhookType(ctx context.Context, webhookType model.WebhookType) ([]*model.Webhook, error)
	ListByApplicationTemplateID(ctx context.Context, applicationTemplateID string) ([]*model.Webhook, error)
	Create(ctx context.Context, tenant string, item *model.Webhook) error
	Update(ctx context.Context, tenant string, item *model.Webhook) error
	Delete(ctx context.Context, id string) error
}

// ApplicationRepository missing godoc
//
//go:generate mockery --name=ApplicationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationRepository interface {
	GetGlobalByID(ctx context.Context, id string) (*model.Application, error)
}

// UIDService missing godoc
//
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

// TenantService is responsible for service-layer tenant operations
//
//go:generate mockery --name=TenantService --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantService interface {
	ExtractTenantIDForTenantScopedFormationTemplates(ctx context.Context) (string, error)
}

// OwningResource missing godoc
type OwningResource string

type service struct {
	webhookRepo              WebhookRepository
	appRepo                  ApplicationRepository
	uidSvc                   UIDService
	tenantSvc                TenantService
	tenantMappingConfig      map[string]interface{}
	tenantMappingCallbackURL string
}

// NewService missing godoc
func NewService(repo WebhookRepository, appRepo ApplicationRepository, uidSvc UIDService, tenantSvc TenantService, tenantMappingConfig map[string]interface{}, tenantMappingCallbackURL string) *service {
	return &service{
		webhookRepo:              repo,
		uidSvc:                   uidSvc,
		appRepo:                  appRepo,
		tenantSvc:                tenantSvc,
		tenantMappingConfig:      tenantMappingConfig,
		tenantMappingCallbackURL: tenantMappingCallbackURL,
	}
}

// Get missing godoc
func (s *service) Get(ctx context.Context, id string, objectType model.WebhookReferenceObjectType) (webhook *model.Webhook, err error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil || tnt == "" {
		log.C(ctx).Infof("tenant was not loaded while getting Webhook id %s", id)
		webhook, err = s.webhookRepo.GetByIDGlobal(ctx, id)
	} else {
		webhook, err = s.webhookRepo.GetByID(ctx, tnt, id, objectType)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Webhook with ID %s", id)
	}
	return
}

// ListForApplication missing godoc
func (s *service) ListForApplication(ctx context.Context, applicationID string) ([]*model.Webhook, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.webhookRepo.ListByReferenceObjectID(ctx, tnt, applicationID, model.ApplicationWebhookReference)
}

// ListForApplicationGlobal missing godoc
func (s *service) ListForApplicationGlobal(ctx context.Context, applicationID string) ([]*model.Webhook, error) {
	return s.webhookRepo.ListByReferenceObjectIDGlobal(ctx, applicationID, model.ApplicationWebhookReference)
}

// ListByWebhookType lists all webhooks with given webhook type
func (s *service) ListByWebhookType(ctx context.Context, webhookType model.WebhookType) ([]*model.Webhook, error) {
	return s.webhookRepo.ListByWebhookType(ctx, webhookType)
}

// ListForApplicationTemplate missing godoc
func (s *service) ListForApplicationTemplate(ctx context.Context, applicationTemplateID string) ([]*model.Webhook, error) {
	return s.webhookRepo.ListByApplicationTemplateID(ctx, applicationTemplateID)
}

// ListAllApplicationWebhooks missing godoc
func (s *service) ListAllApplicationWebhooks(ctx context.Context, applicationID string) ([]*model.Webhook, error) {
	application, err := s.appRepo.GetGlobalByID(ctx, applicationID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Application with ID %s", applicationID)
	}

	return s.retrieveWebhooks(ctx, application)
}

// ListForRuntime missing godoc
func (s *service) ListForRuntime(ctx context.Context, runtimeID string) ([]*model.Webhook, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.webhookRepo.ListByReferenceObjectID(ctx, tnt, runtimeID, model.RuntimeWebhookReference)
}

// ListForFormationTemplate lists all webhooks for a given formationTemplateID
func (s *service) ListForFormationTemplate(ctx context.Context, tenant, formationTemplateID string) ([]*model.Webhook, error) {
	if tenant == "" {
		log.C(ctx).Infof("tenant was not loaded while getting webhooks for formation template with id %s", formationTemplateID)
		return s.webhookRepo.ListByReferenceObjectIDGlobal(ctx, formationTemplateID, model.FormationTemplateWebhookReference)
	}
	return s.webhookRepo.ListByReferenceObjectID(ctx, tenant, formationTemplateID, model.FormationTemplateWebhookReference)
}

// Create creates a model.Webhook with generated ID and CreatedAt properties. Returns the ID of the webhook.
func (s *service) Create(ctx context.Context, owningResourceID string, in model.WebhookInput, objectType model.WebhookReferenceObjectType) (string, error) {
	tenantID, err := s.getTenantForWebhook(ctx, objectType.GetResourceType())
	if apperrors.IsTenantRequired(err) {
		log.C(ctx).Debugf("Creating Webhook with type: %q without tenant", in.Type)
	} else if err != nil {
		return "", err
	}

	id := s.uidSvc.Generate()

	webhook := in.ToWebhook(id, owningResourceID, objectType)

	if err = s.webhookRepo.Create(ctx, tenantID, webhook); err != nil {
		return "", errors.Wrapf(err, "while creating %s with type: %q and ID: %q for: %q", objectType, webhook.Type, id, owningResourceID)
	}
	log.C(ctx).Infof("Successfully created %s with type: %q and ID: %q for: %q", objectType, webhook.Type, id, owningResourceID)

	return webhook.ID, nil
}

// Update missing godoc
func (s *service) Update(ctx context.Context, id string, in model.WebhookInput, objectType model.WebhookReferenceObjectType) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil && objectType.GetResourceType() != resource.Webhook { // If the webhook is not global
		return err
	}
	webhook, err := s.Get(ctx, id, objectType)
	if err != nil {
		return errors.Wrap(err, "while getting Webhook")
	}

	if len(webhook.ObjectID) == 0 || (webhook.ObjectType != model.ApplicationWebhookReference && webhook.ObjectType != model.ApplicationTemplateWebhookReference && webhook.ObjectType != model.RuntimeWebhookReference && webhook.ObjectType != model.FormationTemplateWebhookReference) {
		return errors.New("while updating Webhook: webhook doesn't have neither of application_id, application_template_id, runtime_id and formation_template_id")
	}

	webhook = in.ToWebhook(id, webhook.ObjectID, webhook.ObjectType)

	if err = s.webhookRepo.Update(ctx, tnt, webhook); err != nil {
		return errors.Wrapf(err, "while updating Webhook")
	}

	return nil
}

// Delete missing godoc
func (s *service) Delete(ctx context.Context, id string, objectType model.WebhookReferenceObjectType) error {
	webhook, err := s.Get(ctx, id, objectType)
	if err != nil {
		return errors.Wrap(err, "while getting Webhook")
	}

	return s.webhookRepo.Delete(ctx, webhook.ID)
}

func (s *service) EnrichWebhooksWithTenantMappingWebhooks(in []*graphql.WebhookInput) ([]*graphql.WebhookInput, error) {
	webhooks := make([]*graphql.WebhookInput, 0)
	for _, w := range in {
		if w.Version == nil {
			webhooks = append(webhooks, w)
			continue
		}

		if w.URL == nil || w.Mode == nil {
			return nil, errors.New("url and mode are required fields when version is provided")
		}
		tenantMappingWebhooks, err := s.getTenantMappingWebhooks(w.Mode.String(), *w.Version)
		if err != nil {
			return nil, err
		}
		for _, tenantMappingWebhook := range tenantMappingWebhooks {
			urlTemplate := *tenantMappingWebhook.URLTemplate
			if strings.Contains(urlTemplate, "%s") {
				urlTemplate = fmt.Sprintf(*tenantMappingWebhook.URLTemplate, *w.URL)
			}

			headerTemplate := *tenantMappingWebhook.HeaderTemplate
			if *w.Mode == graphql.WebhookModeAsyncCallback && strings.Contains(headerTemplate, "%s") {
				headerTemplate = fmt.Sprintf(*tenantMappingWebhook.HeaderTemplate, s.tenantMappingCallbackURL)
			}
			wh := &graphql.WebhookInput{
				Type:           tenantMappingWebhook.Type,
				Auth:           w.Auth,
				Mode:           w.Mode,
				URLTemplate:    &urlTemplate,
				InputTemplate:  tenantMappingWebhook.InputTemplate,
				HeaderTemplate: &headerTemplate,
				OutputTemplate: tenantMappingWebhook.OutputTemplate,
			}
			webhooks = append(webhooks, wh)
		}
	}
	return webhooks, nil
}

func (s *service) getTenantMappingWebhooks(mode, version string) ([]graphql.WebhookInput, error) {
	modeObj, ok := s.tenantMappingConfig[mode]
	if !ok {
		return nil, errors.Errorf("missing tenant mapping configuration for mode %s", mode)
	}
	modeMap, ok := modeObj.(map[string]interface{})
	if !ok {
		return nil, errors.Errorf("unexpected mode type, should be a map, but was %T", mode)
	}
	webhooks, ok := modeMap[version]
	if !ok {
		return nil, errors.Errorf("missing tenant mapping configuration for mode %s and version %s", mode, version)
	}

	webhooksJSON, err := json.Marshal(webhooks)
	if err != nil {
		return nil, errors.Wrap(err, "while marshaling webhooks")
	}

	var tenantMappingWebhooks []graphql.WebhookInput
	if err := json.Unmarshal(webhooksJSON, &tenantMappingWebhooks); err != nil {
		return nil, errors.Wrap(err, "while unmarshaling webhooks")
	}

	return tenantMappingWebhooks, nil
}

func (s *service) retrieveWebhooks(ctx context.Context, application *model.Application) ([]*model.Webhook, error) {
	appWebhooks, err := s.ListForApplication(ctx, application.ID)
	if err != nil {
		return nil, err
	}

	var appTemplateWebhooks []*model.Webhook
	if application.ApplicationTemplateID != nil {
		appTemplateWebhooks, err = s.ListForApplicationTemplate(ctx, *application.ApplicationTemplateID)
		if err != nil {
			return nil, err
		}
	}

	webhooksMap := make(map[model.WebhookType]*model.Webhook)
	for i, webhook := range appTemplateWebhooks {
		webhooksMap[webhook.Type] = appTemplateWebhooks[i]
	}
	// Override values derived from template
	for i, webhook := range appWebhooks {
		webhooksMap[webhook.Type] = appWebhooks[i]
	}

	webhooks := make([]*model.Webhook, 0)
	for key := range webhooksMap {
		webhooks = append(webhooks, webhooksMap[key])
	}

	return webhooks, nil
}

func (s *service) getTenantForWebhook(ctx context.Context, whType resource.Type) (string, error) {
	if whType == resource.FormationTemplateWebhook {
		return s.tenantSvc.ExtractTenantIDForTenantScopedFormationTemplates(ctx)
	}
	return tenant.LoadFromContext(ctx)
}
