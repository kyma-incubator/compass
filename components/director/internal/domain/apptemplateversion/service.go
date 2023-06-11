package apptemplateversion

import (
	"context"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	directortime "github.com/kyma-incubator/compass/components/director/pkg/time"
	"github.com/pkg/errors"
)

// ApplicationTemplateVersionRepository is responsible for repo-layer ApplicationTemplateVersion operations
//
//go:generate mockery --name=ApplicationTemplateVersionRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationTemplateVersionRepository interface {
	GetByAppTemplateIDAndVersion(ctx context.Context, appTemplateID, version string) (*model.ApplicationTemplateVersion, error)
	ListByAppTemplateID(ctx context.Context, appTemplateID string) ([]*model.ApplicationTemplateVersion, error)
	Create(ctx context.Context, item model.ApplicationTemplateVersion) error
	Exists(ctx context.Context, id string) (bool, error)
	Update(ctx context.Context, model model.ApplicationTemplateVersion) error
}

// UIDService is responsible for UID operations
//
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

// TimeService is responsible for Time operations
//
//go:generate mockery --name=TimeService --output=automock --outpkg=automock --case=underscore --disable-version-string
type TimeService interface {
	Now() time.Time
}

type service struct {
	appTemplateVersionRepo ApplicationTemplateVersionRepository
	uidService             UIDService
	timeService            directortime.Service
}

// NewService returns a new object responsible for service-layer ApplicationTemplateVersion operations.
func NewService(appTemplateVersionRepo ApplicationTemplateVersionRepository, uidService UIDService, timeService TimeService) *service {
	return &service{
		appTemplateVersionRepo: appTemplateVersionRepo,
		uidService:             uidService,
		timeService:            timeService,
	}
}

// Create creates an ApplicationTemplateVersion for a given applicationTemplateID
func (s *service) Create(ctx context.Context, applicationTemplateID string, in *model.ApplicationTemplateVersionInput) (string, error) {
	if in == nil {
		return "", nil
	}

	id := s.uidService.Generate()
	applicationTemplateVersion := in.ToApplicationTemplateVersion(id, applicationTemplateID)
	applicationTemplateVersion.CreatedAt = s.timeService.Now()

	if err := s.appTemplateVersionRepo.Create(ctx, applicationTemplateVersion); err != nil {
		return "", errors.Wrapf(err, "error occurred while creating an Application Template Version with id %s", id)
	}
	log.C(ctx).Infof("Successfully created Application Template Version with ID %s", id)

	return id, nil
}

// Update checks if a ApplicationTemplateVersion exists and updates it
func (s *service) Update(ctx context.Context, id, appTemplateID string, in model.ApplicationTemplateVersionInput) error {
	exists, err := s.appTemplateVersionRepo.Exists(ctx, id)
	if err != nil {
		return err
	}

	if !exists {
		return errors.Errorf("Application Template Version with ID %s does not exist", id)
	}

	applicationTemplateVersion := in.ToApplicationTemplateVersion(id, appTemplateID)

	if err = s.appTemplateVersionRepo.Update(ctx, applicationTemplateVersion); err != nil {
		return errors.Wrapf(err, "while updating Application Template Version with id %s", id)
	}

	return nil
}

// GetByAppTemplateIDAndVersion gets an ApplicationTemplateVersion by a given Application Template ID and a version
func (s *service) GetByAppTemplateIDAndVersion(ctx context.Context, appTemplateID, version string) (*model.ApplicationTemplateVersion, error) {
	applicationTemplateVersion, err := s.appTemplateVersionRepo.GetByAppTemplateIDAndVersion(ctx, appTemplateID, version)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Application Template Version with Version %q and Application Template ID: %q", version, appTemplateID)
	}

	return applicationTemplateVersion, nil
}

// ListByAppTemplateID lists multiple ApplicationTemplateVersion by Application Template ID
func (s *service) ListByAppTemplateID(ctx context.Context, appTemplateID string) ([]*model.ApplicationTemplateVersion, error) {
	applicationTemplateVersion, err := s.appTemplateVersionRepo.ListByAppTemplateID(ctx, appTemplateID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Application Template Version with Application Template ID: %q", appTemplateID)
	}

	return applicationTemplateVersion, nil
}
