package apptemplateversion

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

// ApplicationTemplateVersionRepository missing godoc
//
//go:generate mockery --name=ApplicationTemplateVersionRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationTemplateVersionRepository interface {
	GetByAppTemplateIDAndVersion(ctx context.Context, appTemplateID, version string) (*model.ApplicationTemplateVersion, error)
	ListByAppTemplateID(ctx context.Context, appTemplateID string) ([]*model.ApplicationTemplateVersion, error)
	Create(ctx context.Context, item model.ApplicationTemplateVersion) error
	Exists(ctx context.Context, id string) (bool, error)
	Update(ctx context.Context, model model.ApplicationTemplateVersion) error
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
func (s *service) Create(ctx context.Context, applicationTemplateID string, in *model.ApplicationTemplateVersionInput) (string, error) {
	if in == nil {
		return "", nil
	}

	id := s.uidService.Generate()
	applicationTemplateVersion := in.ToApplicationTemplateVersion(id, applicationTemplateID)

	if err := s.appTemplateVersionRepo.Create(ctx, applicationTemplateVersion); err != nil {
		return "", errors.Wrapf(err, "error occurred while creating a Application Template Version with id %s", id)
	}
	log.C(ctx).Debugf("Successfully create Application Template Version with ID %s", id)

	return id, nil
}

func (s *service) Update(ctx context.Context, id, appTemplateID string, in *model.ApplicationTemplateVersionInput) error {
	exists, err := s.appTemplateVersionRepo.Exists(ctx, id)
	if err != nil {
		return err
	}

	if !exists {
		return errors.Errorf("Application Template Version with ID %s does not exist", id)
	}

	appTemplateVersion := in.ToApplicationTemplateVersion(id, appTemplateID)

	if err = s.appTemplateVersionRepo.Update(ctx, appTemplateVersion); err != nil {
		return errors.Wrapf(err, "while updating APIDefinition with id %s", id)
	}

	return nil
}

// GetByAppTemplateIDAndVersion missing godoc
func (s *service) GetByAppTemplateIDAndVersion(ctx context.Context, appTemplateID, version string) (*model.ApplicationTemplateVersion, error) {
	applicationTemplateVersion, err := s.appTemplateVersionRepo.GetByAppTemplateIDAndVersion(ctx, appTemplateID, version)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Application Template Version with Version %q and Application Template ID: %q", version, appTemplateID)
	}

	return applicationTemplateVersion, nil
}

// GetByAppTemplateID missing godoc
func (s *service) ListByAppTemplateID(ctx context.Context, appTemplateID string) ([]*model.ApplicationTemplateVersion, error) {
	applicationTemplateVersion, err := s.appTemplateVersionRepo.ListByAppTemplateID(ctx, appTemplateID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Application Template Version with Application Template ID: %q", appTemplateID)
	}

	return applicationTemplateVersion, nil
}
