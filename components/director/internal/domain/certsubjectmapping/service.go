package certsubjectmapping

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

// CertMappingRepository represents the certificate subject mapping repository layer
//go:generate mockery --name=CertMappingRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type CertMappingRepository interface {
	Create(ctx context.Context, item *model.CertSubjectMapping) error
	Get(ctx context.Context, id string) (*model.CertSubjectMapping, error)
	Update(ctx context.Context, model *model.CertSubjectMapping) error
	Delete(ctx context.Context, id string) error
	Exists(ctx context.Context, id string) (bool, error)
	List(ctx context.Context, pageSize int, cursor string) (*model.CertSubjectMappingPage, error)
}

type service struct {
	repo CertMappingRepository
}

// NewService returns a new object responsible for service-layer certificate subject mapping operations.
func NewService(repo CertMappingRepository) *service {
	return &service{
		repo: repo,
	}
}

// Create creates a certificate subject mapping using `item`
func (s *service) Create(ctx context.Context, item *model.CertSubjectMapping) (string, error) {
	log.C(ctx).Infof("Creating certificate subject mapping with ID: %s, subject: %s and consumer type: %s", item.ID, item.Subject, item.ConsumerType)
	if err := s.repo.Create(ctx, item); err != nil {
		return "", errors.Wrapf(err, "while creating certificate subject mapping with subject: %s and consumer type: %s", item.Subject, item.ConsumerType)
	}

	return item.ID, nil
}

// Get queries certificate subject mapping matching ID `id`
func (s *service) Get(ctx context.Context, id string) (*model.CertSubjectMapping, error) {
	log.C(ctx).Infof("Getting certificate subject mapping with ID: %s", id)
	csm, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting certificate subject mapping with ID: %s", id)
	}

	return csm, nil
}

// Update updates a certificate subject mapping using `in`
func (s *service) Update(ctx context.Context, in *model.CertSubjectMapping) error {
	log.C(ctx).Infof("Updating certificate subject mapping with ID: %s, subject: %s and consumer type: %s", in.ID, in.Subject, in.ConsumerType)

	if exists, err := s.repo.Exists(ctx, in.ID); err != nil {
		return errors.Wrapf(err, "while ensuring certificate subject mapping with ID: %s exists", in.ID)
	} else if !exists {
		return apperrors.NewNotFoundError(resource.CertSubjectMapping, in.ID)
	}

	if err := s.repo.Update(ctx, in); err != nil {
		return errors.Wrapf(err, "while updating certificate subject mapping with ID: %s", in.ID)
	}

	return nil
}

// Delete deletes a certificate subject mapping matching ID `id`
func (s *service) Delete(ctx context.Context, id string) error {
	log.C(ctx).Infof("Deleting certificate subject mapping with ID: %s", id)
	if err := s.repo.Delete(ctx, id); err != nil {
		return errors.Wrapf(err, "while deleting certificate subject mapping with ID: %s", id)
	}
	return nil
}

// Exists check if a certificate subject mapping with ID `id` exists
func (s *service) Exists(ctx context.Context, id string) (bool, error) {
	log.C(ctx).Infof("Checking certificate subject mapping existence for ID: %s", id)
	exists, err := s.repo.Exists(ctx, id)
	if err != nil {
		return false, errors.Wrapf(err, "while checking certificate subject mapping existence for ID: %s", id)
	}
	return exists, nil
}

// List retrieves certificate subject mappings with pagination based on `pageSize` and `cursor`
func (s *service) List(ctx context.Context, pageSize int, cursor string) (*model.CertSubjectMappingPage, error) {
	log.C(ctx).Info("Listing certificate subject mappings")
	if pageSize < 1 || pageSize > 300 {
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 300")
	}

	return s.repo.List(ctx, pageSize, cursor)
}
