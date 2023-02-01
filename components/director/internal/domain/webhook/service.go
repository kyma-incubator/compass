package webhook

import (
	"context"
	tnt "github.com/kyma-incubator/compass/components/director/pkg/tenant"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

// WebhookRepository missing godoc
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
//go:generate mockery --name=ApplicationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationRepository interface {
	GetGlobalByID(ctx context.Context, id string) (*model.Application, error)
}

// UIDService missing godoc
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

// TenantService is responsible for service-layer tenant operations
//go:generate mockery --name=TenantService --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantService interface {
	GetTenantByID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
}

// OwningResource missing godoc
type OwningResource string

type service struct {
	webhookRepo WebhookRepository
	appRepo     ApplicationRepository
	uidSvc      UIDService
	tenantSvc   TenantService
}

// NewService missing godoc
func NewService(repo WebhookRepository, appRepo ApplicationRepository, uidSvc UIDService, tenantSvc TenantService) *service {
	return &service{
		webhookRepo: repo,
		uidSvc:      uidSvc,
		appRepo:     appRepo,
		tenantSvc:   tenantSvc,
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
	tenantId, err := s.getTenantForWebhook(ctx, objectType.GetResourceType())
	if apperrors.IsTenantRequired(err) {
		log.C(ctx).Debugf("Creating Webhook with type: %q without tenant", in.Type)
	} else if err != nil {
		return "", err
	}

	id := s.uidSvc.Generate()

	webhook := in.ToWebhook(id, owningResourceID, objectType)

	if err = s.webhookRepo.Create(ctx, tenantId, webhook); err != nil {
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
		return s.extractTenantIDForTenantScopedFormationTemplates(ctx)
	}
	return tenant.LoadFromContext(ctx)
}

// getTenantFromContext validates and returns the tenant present in the context:
// 		1. if only one ID is present -> throw TenantNotFoundError
// 		2. if both internalID and externalID are present -> proceed with tenant scoped formation templates (return the internalID from ctx)
// 		3. if both internalID and externalID are not present -> proceed with global formation templates (return empty id)
func (s *service) getTenantFromContext(ctx context.Context) (string, error) {
	tntCtx, err := tenant.LoadTenantPairFromContextNoChecks(ctx)
	if err != nil {
		return "", err
	}

	if (tntCtx.InternalID == "" && tntCtx.ExternalID != "") || (tntCtx.InternalID != "" && tntCtx.ExternalID == "") {
		return "", apperrors.NewTenantNotFoundError(tntCtx.ExternalID)
	}

	var internalTenantID string
	if tntCtx.InternalID != "" && tntCtx.ExternalID != "" {
		internalTenantID = tntCtx.InternalID
	}

	return internalTenantID, nil
}

// extractTenantIDForTenantScopedFormationTemplates returns the tenant ID based on its type:
//		1. If it's not SA or GA -> return error
//		2. If it's GA -> return the GA id
//		3. If it's a SA -> return its parent GA id
func (s *service) extractTenantIDForTenantScopedFormationTemplates(ctx context.Context) (string, error) {
	internalTenantID, err := s.getTenantFromContext(ctx)
	if err != nil {
		return "", err
	}

	if internalTenantID == "" {
		return internalTenantID, nil
	}

	tenantObject, err := s.tenantSvc.GetTenantByID(ctx, internalTenantID)
	if err != nil {
		return "", err
	}

	if tenantObject.Type != tnt.Account && tenantObject.Type != tnt.Subaccount {
		return "", errors.New("tenant used for tenant scoped Formation Templates must be of type account or subaccount")
	}

	if tenantObject.Type == tnt.Account {
		return tenantObject.ID, nil
	}

	gaTenantObject, err := s.tenantSvc.GetTenantByID(ctx, tenantObject.Parent)
	if err != nil {
		return "", err
	}

	return gaTenantObject.ID, nil
}
