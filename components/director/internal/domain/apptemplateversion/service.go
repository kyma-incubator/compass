package apptemplateversion

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

// ApplicationTemplateVersionRepository missing godoc
//
//go:generate mockery --name=ApplicationTemplateVersionRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationTemplateVersionRepository interface {
	Create(ctx context.Context, tenant string, item model.ApplicationTemplateVersion) error
	Update(ctx context.Context, tenant string, item *model.ApplicationTemplateVersion) error
	Delete(ctx context.Context, tenant, id string) error
	Exists(ctx context.Context, tenant, id string) (bool, error)
	GetByID(ctx context.Context, tenant, id string) (*model.ApplicationTemplateVersion, error)
}

// UIDService missing godoc
//
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

type service struct {
	appTemplateVersionRepo ApplicationTemplateVersionRepository
	uidService             UIDService
}

// NewService missing godoc
func NewService(appTemplateVersionRepo ApplicationTemplateVersionRepository, uidService UIDService) *service {
	return &service{
		appTemplateVersionRepo: appTemplateVersionRepo,
		uidService:             uidService,
	}
}

// Create missing godoc
func (s *service) Create(ctx context.Context, applicationTemplateID string, in model.ApplicationTemplateVersionInput) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	id := s.uidService.Generate()
	applicationTemplateVersion := in.ToApplicationTemplateVersion(id, applicationTemplateID)

	if err = s.appTemplateVersionRepo.Create(ctx, tnt, applicationTemplateVersion); err != nil {
		return "", errors.Wrapf(err, "error occurred while creating a Tombstone with id %s for Application with id %s", id, applicationTemplateID)
	}
	log.C(ctx).Debugf("Successfully created a Tombstone with id %s for Application with id %s", id, applicationTemplateID)

	return applicationTemplateVersion.ID, nil
}

// Update missing godoc
func (s *service) Update(ctx context.Context, id string, in model.TombstoneInput) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	applicationTemplateVersion, err := s.appTemplateVersionRepo.GetByID(ctx, tnt, id)
	if err != nil {
		return errors.Wrapf(err, "while getting Tombstone with id %s", id)
	}

	if err = s.appTemplateVersionRepo.Update(ctx, tnt, applicationTemplateVersion); err != nil {
		return errors.Wrapf(err, "while updating Tombstone with id %s", id)
	}
	return nil
}

// Delete missing godoc
func (s *service) Delete(ctx context.Context, id string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "while loading tenant from context")
	}

	err = s.appTemplateVersionRepo.Delete(ctx, tnt, id)
	if err != nil {
		return errors.Wrapf(err, "while deleting Tombstone with id %s", id)
	}

	return nil
}

// Exist missing godoc
func (s *service) Exist(ctx context.Context, id string) (bool, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return false, errors.Wrap(err, "while loading tenant from context")
	}

	exist, err := s.appTemplateVersionRepo.Exists(ctx, tnt, id)
	if err != nil {
		return false, errors.Wrapf(err, "while getting Tombstone with ID: %q", id)
	}

	return exist, nil
}

// Get missing godoc
func (s *service) Get(ctx context.Context, id string) (*model.ApplicationTemplateVersion, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while loading tenant from context")
	}

	applicationTemplateVersion, err := s.appTemplateVersionRepo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Tombstone with ID: %q", id)
	}

	return applicationTemplateVersion, nil
}
