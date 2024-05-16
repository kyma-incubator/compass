package formationtemplate

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

// FormationTemplateRepository represents the FormationTemplate repository layer
//
//go:generate mockery --name=FormationTemplateRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationTemplateRepository interface {
	Create(ctx context.Context, item *model.FormationTemplate) error
	Get(ctx context.Context, id string) (*model.FormationTemplate, error)
	List(ctx context.Context, filters []*labelfilter.LabelFilter, name *string, tenantID string, pageSize int, cursor string) (*model.FormationTemplatePage, error)
	Update(ctx context.Context, model *model.FormationTemplate) error
	Delete(ctx context.Context, id, tenantID string) error
	ExistsGlobal(ctx context.Context, id string) (bool, error)
}

// UIDService generates UUIDs for new entities
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

// WebhookRepository is responsible for repo-layer Webhook operations
//
//go:generate mockery --name=WebhookRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type WebhookRepository interface {
	CreateMany(ctx context.Context, tenant string, items []*model.Webhook) error
}

// WebhookService represents the Webhook service layer
//
//go:generate mockery --name=WebhookService --output=automock --outpkg=automock --case=underscore --disable-version-string
type WebhookService interface {
	ListForFormationTemplate(ctx context.Context, tenant, formationTemplateID string) ([]*model.Webhook, error)
}

//go:generate mockery --exported --name=labelService --output=automock --outpkg=automock --case=underscore --disable-version-string
type labelService interface {
	UpsertMultipleLabels(ctx context.Context, tenantID string, objectType model.LabelableObject, objectID string, labels map[string]interface{}) error
	UpsertLabel(ctx context.Context, tenantID string, labelInput *model.LabelInput) error
	GetByKey(ctx context.Context, tenantID string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
	Delete(ctx context.Context, tenantID string, objectType model.LabelableObject, objectID, key string) error
	ListForObject(ctx context.Context, tenantID string, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error)
}

type service struct {
	repo           FormationTemplateRepository
	uidSvc         UIDService
	converter      FormationTemplateConverter
	tenantSvc      TenantService
	webhookRepo    WebhookRepository
	webhookService WebhookService
	labelService   labelService
}

// NewService creates a FormationTemplate service
func NewService(repo FormationTemplateRepository, uidSvc UIDService, converter FormationTemplateConverter, tenantSvc TenantService, webhookRepo WebhookRepository, webhookService WebhookService, labelService labelService) *service {
	return &service{
		repo:           repo,
		uidSvc:         uidSvc,
		converter:      converter,
		tenantSvc:      tenantSvc,
		webhookRepo:    webhookRepo,
		webhookService: webhookService,
		labelService:   labelService,
	}
}

// Create creates a FormationTemplate using `in`
func (s *service) Create(ctx context.Context, in *model.FormationTemplateRegisterInput) (string, error) {
	formationTemplateID := s.uidSvc.Generate()

	if in != nil {
		log.C(ctx).Debugf("ID %s generated for Formation Template with name %s", formationTemplateID, in.Name)
	}

	tenantID, err := s.tenantSvc.ExtractTenantIDForTenantScopedFormationTemplates(ctx)
	if err != nil {
		return "", err
	}

	formationTemplateModel := s.converter.FromModelRegisterInputToModel(in, formationTemplateID, tenantID)

	err = s.repo.Create(ctx, formationTemplateModel)
	if err != nil {
		return "", errors.Wrapf(err, "while creating Formation Template with name %s", in.Name)
	}

	if in != nil && in.Labels != nil {
		if err = s.labelService.UpsertMultipleLabels(ctx, tenantID, model.FormationTemplateLabelableObject, formationTemplateID, in.Labels); err != nil {
			return "", errors.Wrapf(err, "while upserting labels for formation template with ID: %s", formationTemplateID)
		}
	}

	if err = s.webhookRepo.CreateMany(ctx, tenantID, formationTemplateModel.Webhooks); err != nil {
		return "", errors.Wrapf(err, "while creating webhooks for formation template with ID: %s", formationTemplateID)
	}

	return formationTemplateID, nil
}

func (s *service) Exist(ctx context.Context, id string) (bool, error) {
	return s.repo.ExistsGlobal(ctx, id)
}

// Get queries FormationTemplate matching ID `id`
func (s *service) Get(ctx context.Context, id string) (*model.FormationTemplate, error) {
	formationTemplate, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting formation template with ID: %s", id)
	}

	return formationTemplate, nil
}

// List pagination lists FormationTemplate based on `pageSize` and `cursor`
func (s *service) List(ctx context.Context, filters []*labelfilter.LabelFilter, name *string, pageSize int, cursor string) (*model.FormationTemplatePage, error) {
	if pageSize < 1 || pageSize > 200 {
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 200")
	}

	tenantID, err := s.tenantSvc.ExtractTenantIDForTenantScopedFormationTemplates(ctx)
	if err != nil {
		return nil, err
	}

	return s.repo.List(ctx, filters, name, tenantID, pageSize, cursor)
}

// Update updates a FormationTemplate matching ID `id` using `in`
func (s *service) Update(ctx context.Context, id string, in *model.FormationTemplateUpdateInput) error {
	if err := s.ensureFormationTemplateExists(ctx, id); err != nil {
		return err
	}

	tenantID, err := s.tenantSvc.ExtractTenantIDForTenantScopedFormationTemplates(ctx)
	if err != nil {
		return err
	}

	err = s.repo.Update(ctx, s.converter.FromModelUpdateInputToModel(in, id, tenantID))
	if err != nil {
		return errors.Wrapf(err, "while updating formation template with ID: %s", id)
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
		return errors.Wrapf(err, "while deleting formation template with ID: %s", id)
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

// SetLabel takes the provided label input and either creates (in case the label does not exist)
// or updates (in case the label already exists) a label for formation template
func (s *service) SetLabel(ctx context.Context, labelInput *model.LabelInput) error {
	tenantID, err := s.tenantSvc.ExtractTenantIDForTenantScopedFormationTemplates(ctx)
	if err != nil {
		return err
	}

	if err := s.ensureFormationTemplateExists(ctx, labelInput.ObjectID); err != nil {
		return err
	}

	if err = s.labelService.UpsertLabel(ctx, tenantID, labelInput); err != nil {
		return errors.Wrap(err, "while upserting label for formation template")
	}

	return nil
}

// DeleteLabel deletes a label with the provided key for formation template that has the provided 'formationTemplateID'
func (s *service) DeleteLabel(ctx context.Context, formationTemplateID string, key string) error {
	tenantID, err := s.tenantSvc.ExtractTenantIDForTenantScopedFormationTemplates(ctx)
	if err != nil {
		return err
	}

	if err := s.ensureFormationTemplateExists(ctx, formationTemplateID); err != nil {
		return err
	}

	if err := s.labelService.Delete(ctx, tenantID, model.FormationTemplateLabelableObject, formationTemplateID, key); err != nil {
		return errors.Wrap(err, "while deleting formation template label")
	}

	return nil
}

// GetLabel retrieves a label with the provided key for formation template that has the provided 'formationTemplateID'
func (s *service) GetLabel(ctx context.Context, formationTemplateID string, key string) (*model.Label, error) {
	tenantID, err := s.tenantSvc.ExtractTenantIDForTenantScopedFormationTemplates(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.ensureFormationTemplateExists(ctx, formationTemplateID); err != nil {
		return nil, err
	}

	lbl, err := s.labelService.GetByKey(ctx, tenantID, model.FormationTemplateLabelableObject, formationTemplateID, key)
	if err != nil {
		return nil, errors.Wrapf(err, "while gettting label for formation template with ID: %s and key: %s", formationTemplateID, key)
	}

	return lbl, nil
}

// ListLabels retrieves all labels for application template
func (s *service) ListLabels(ctx context.Context, formationTemplateID string) (map[string]*model.Label, error) {
	tenantID, err := s.tenantSvc.ExtractTenantIDForTenantScopedFormationTemplates(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.ensureFormationTemplateExists(ctx, formationTemplateID); err != nil {
		return nil, err
	}

	labels, err := s.labelService.ListForObject(ctx, tenantID, model.FormationTemplateLabelableObject, formationTemplateID)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing label for formation template with ID: %s", formationTemplateID)
	}

	return labels, nil
}

func (s *service) ensureFormationTemplateExists(ctx context.Context, formationTemplateID string) error {
	if ftExists, err := s.repo.ExistsGlobal(ctx, formationTemplateID); err != nil {
		return errors.Wrapf(err, "while checking existence of formation template with ID: %s", formationTemplateID)
	} else if !ftExists {
		return apperrors.NewNotFoundError(resource.FormationTemplate, formationTemplateID)
	}

	return nil
}
