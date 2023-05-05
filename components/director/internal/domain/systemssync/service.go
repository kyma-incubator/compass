package systemssync

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

// SystemsSyncRepository represents the Systems Sync timestamps repository layer
//
//go:generate mockery --name=SystemsSyncRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type SystemsSyncRepository interface {
	List(ctx context.Context) ([]*model.SystemSynchronizationTimestamp, error)
	Upsert(ctx context.Context, in *model.SystemSynchronizationTimestamp) error
}

type service struct {
	syncSystemsRepo SystemsSyncRepository
}

// NewService returns new SystemsSyncService
func NewService(syncSystemsRepo SystemsSyncRepository) *service {
	return &service{
		syncSystemsRepo: syncSystemsRepo,
	}
}

// List returns all synchronization timestamps of the systems
func (s *service) List(ctx context.Context) ([]*model.SystemSynchronizationTimestamp, error) {
	log.C(ctx).Infof("Listing systems sync timestamps for all tenants")

	syncTimestamps, err := s.syncSystemsRepo.List(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error while listing the sync timestamps")
	}

	return syncTimestamps, nil
}

// Upsert updates sync timestamp or creates new one if it doesn't exist
func (s *service) Upsert(ctx context.Context, in *model.SystemSynchronizationTimestamp) error {
	log.C(ctx).Infof("Upserting systems sync timestamps")

	err := s.syncSystemsRepo.Upsert(ctx, in)
	if err != nil {
		return errors.Wrap(err, "error while upserting the sync timestamp")
	}

	return nil
}
