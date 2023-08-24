package formationtemplate

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

// FormationTemplateRepository represents the FormationTemplate repository layer
//go:generate mockery --name=FormationTemplateRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationTemplateRepository interface {
	Create(ctx context.Context, item *model.FormationTemplate) error
	Get(ctx context.Context, id string) (*model.FormationTemplate, error)
	List(ctx context.Context, name *string, tenantID string, pageSize int, cursor string) (*model.FormationTemplatePage, error)
	Update(ctx context.Context, model *model.FormationTemplate) error
	Delete(ctx context.Context, id, tenantID string) error
	Exists(ctx context.Context, id string) (bool, error)
}

// UIDService generates UUIDs for new entities
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

// TenantService is responsible for service-layer tenant operations
//go:generate mockery --name=TenantService --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantService interface {
	ExtractTenantIDForTenantScopedFormationTemplates(ctx context.Context) (string, error)
}

// WebhookRepository is responsible for repo-layer Webhook operations
//go:generate mockery --name=WebhookRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type WebhookRepository interface {
	CreateMany(ctx context.Context, tenant string, items []*model.Webhook) error
}

// WebhookService represents the Webhook service layer
//go:generate mockery --name=WebhookService --output=automock --outpkg=automock --case=underscore --disable-version-string
type WebhookService interface {
	ListForFormationTemplate(ctx context.Context, tenant, formationTemplateID string) ([]*model.Webhook, error)
}

type service struct {
	repo           FormationTemplateRepository
	uidSvc         UIDService
	converter      FormationTemplateConverter
	tenantSvc      TenantService
	webhookRepo    WebhookRepository
	webhookService WebhookService
}

// NewService creates a FormationTemplate service
func NewService(repo FormationTemplateRepository, uidSvc UIDService, converter FormationTemplateConverter, tenantSvc TenantService, webhookRepo WebhookRepository, webhookService WebhookService) *service {
	return &service{
		repo:           repo,
		uidSvc:         uidSvc,
		converter:      converter,
		tenantSvc:      tenantSvc,
		webhookRepo:    webhookRepo,
		webhookService: webhookService,
	}
}

// Create creates a FormationTemplate using `in`
func (s *service) Create(ctx context.Context, in *model.FormationTemplateInput) (string, error) {
	formationTemplateID := s.uidSvc.Generate()

	if in != nil {
		log.C(ctx).Debugf("ID %s generated for Formation Template with name %s", formationTemplateID, in.Name)
	}

	tenantID, err := s.tenantSvc.ExtractTenantIDForTenantScopedFormationTemplates(ctx)
	if err != nil {
		return "", err
	}

	formationTemplateModel := s.converter.FromModelInputToModel(in, formationTemplateID, tenantID)

	err = s.repo.Create(ctx, formationTemplateModel)
	if err != nil {
		return "", errors.Wrapf(err, "while creating Formation Template with name %s", in.Name)
	}

	if err = s.webhookRepo.CreateMany(ctx, tenantID, formationTemplateModel.Webhooks); err != nil {
		return "", errors.Wrapf(err, "while creating Webhooks for Formation Template with ID: %s", formationTemplateID)
	}

	return formationTemplateID, nil
}

func (s *service) Exist(ctx context.Context, id string) (bool, error) {
	return s.repo.Exists(ctx, id)
}

// Get queries FormationTemplate matching ID `id`
func (s *service) Get(ctx context.Context, id string) (*model.FormationTemplate, error) {
	formationTemplate, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Formation Template with id %s", id)
	}

	return formationTemplate, nil
}

// List pagination lists FormationTemplate based on `pageSize` and `cursor`
func (s *service) List(ctx context.Context, name *string, pageSize int, cursor string) (*model.FormationTemplatePage, error) {
	if pageSize < 1 || pageSize > 200 {
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 200")
	}

	tenantID, err := s.tenantSvc.ExtractTenantIDForTenantScopedFormationTemplates(ctx)
	if err != nil {
		return nil, err
	}

	return s.repo.List(ctx, name, tenantID, pageSize, cursor)
}

// Update updates a FormationTemplate matching ID `id` using `in`
func (s *service) Update(ctx context.Context, id string, in *model.FormationTemplateInput) error {
	exists, err := s.repo.Exists(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "while ensuring Formation Template with ID %s exists", id)
	} else if !exists {
		return apperrors.NewNotFoundError(resource.FormationTemplate, id)
	}

	tenantID, err := s.tenantSvc.ExtractTenantIDForTenantScopedFormationTemplates(ctx)
	if err != nil {
		return err
	}

	err = s.repo.Update(ctx, s.converter.FromModelInputToModel(in, id, tenantID))
	if err != nil {
		return errors.Wrapf(err, "while updating Formation Template with ID %s", id)
	}

	return nil
}

// Delete deletes a FormationTemplate matching ID `id`
func (s *service) Delete(ctx context.Context, id string) error {
	tenantID, err := s.tenantSvc.ExtractTenantIDForTenantScopedFormationTemplates(ctx)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id, tenantID); err != nil {
		return errors.Wrapf(err, "while deleting Formation Template with ID %s", id)
	}

	return nil
}

// ListWebhooksForFormationTemplate lists webhooks for a FormationTemplate matching ID `formationTemplateID`
func (s *service) ListWebhooksForFormationTemplate(ctx context.Context, formationTemplateID string) ([]*model.Webhook, error) {
	tenantID, err := s.tenantSvc.ExtractTenantIDForTenantScopedFormationTemplates(ctx)
	if err != nil {
		return nil, err
	}

	return s.webhookService.ListForFormationTemplate(ctx, tenantID, formationTemplateID)
}
