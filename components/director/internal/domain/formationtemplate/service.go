package formationtemplate

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	tnt "github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/pkg/errors"
)

// FormationTemplateRepository represents the FormationTemplate repository layer
//go:generate mockery --name=FormationTemplateRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationTemplateRepository interface {
	Create(ctx context.Context, item *model.FormationTemplate) error
	Get(ctx context.Context, id string) (*model.FormationTemplate, error)
	List(ctx context.Context, tenantID string, pageSize int, cursor string) (*model.FormationTemplatePage, error)
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
	GetTenantByID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
}

type service struct {
	repo      FormationTemplateRepository
	uidSvc    UIDService
	converter FormationTemplateConverter
	tenantSvc TenantService
}

// NewService creates a FormationTemplate service
func NewService(repo FormationTemplateRepository, uidSvc UIDService, converter FormationTemplateConverter, tenantSvc TenantService) *service {
	return &service{
		repo:      repo,
		uidSvc:    uidSvc,
		converter: converter,
		tenantSvc: tenantSvc,
	}
}

// Create creates a FormationTemplate using `in`
func (s *service) Create(ctx context.Context, in *model.FormationTemplateInput) (string, error) {
	formationTemplateID := s.uidSvc.Generate()

	log.C(ctx).Debugf("ID %s generated for Formation Template with name %s", formationTemplateID, in.Name)

	tenantID, err := s.extractTenantIDForTenantScopedFormationTemplates(ctx)
	if err != nil {
		return "", err
	}

	err = s.repo.Create(ctx, s.converter.FromModelInputToModel(in, formationTemplateID, tenantID))
	if err != nil {
		return "", errors.Wrapf(err, "while creating Formation Template with name %s", in.Name)
	}

	return formationTemplateID, nil
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
func (s *service) List(ctx context.Context, pageSize int, cursor string) (*model.FormationTemplatePage, error) {
	if pageSize < 1 || pageSize > 200 {
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 200")
	}

	tenantID, err := s.extractTenantIDForTenantScopedFormationTemplates(ctx)
	if err != nil {
		return nil, err
	}

	return s.repo.List(ctx, tenantID, pageSize, cursor)
}

// Update updates a FormationTemplate matching ID `id` using `in`
func (s *service) Update(ctx context.Context, id string, in *model.FormationTemplateInput) error {
	exists, err := s.repo.Exists(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "while ensuring Formation Template with ID %s exists", id)
	} else if !exists {
		return apperrors.NewNotFoundError(resource.FormationTemplate, id)
	}

	tenantID, err := s.extractTenantIDForTenantScopedFormationTemplates(ctx)
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
	tenantID, err := s.extractTenantIDForTenantScopedFormationTemplates(ctx)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id, tenantID); err != nil {
		return errors.Wrapf(err, "while deleting Formation Template with ID %s", id)
	}

	return nil
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
