package operation

import (
	"context"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

// OperationRepository is responsible for repository-layer operation operations
//
//go:generate mockery --name=OperationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type OperationRepository interface {
	Create(ctx context.Context, model *model.Operation) error
	DeleteOlderThan(ctx context.Context, opType, status string, date time.Time) error
}

// UIDService is responsible for service-layer uid operations
//
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

type service struct {
	opRepo     OperationRepository
	uidService UIDService
}

// NewService creates operations service
func NewService(opRepo OperationRepository, uidService UIDService) *service {
	return &service{
		opRepo:     opRepo,
		uidService: uidService,
	}
}

// Create creates new operation entity
func (s *service) Create(ctx context.Context, in *model.OperationInput) error {
	id := s.uidService.Generate()
	op := in.ToOperation(id)

	if err := s.opRepo.Create(ctx, op); err != nil {
		return errors.Wrapf(err, "error occurred while creating an Operation with id %s and type %s", op.ID, op.OpType)
	}

	log.C(ctx).Infof("Successfully created an Operation with id %s and type %s", op.ID, op.OpType)
	return nil
}

// CreateMultiple creates multiple operations
func (s *service) CreateMultiple(ctx context.Context, in []*model.OperationInput) error {
	if in == nil {
		return nil
	}

	for _, op := range in {
		if op == nil {
			continue
		}

		if err := s.Create(ctx, op); err != nil {
			return err
		}
	}

	return nil
}

// DeleteOlderThan deletes all operations of type `opType` with status `status` older than `days`
func (s *service) DeleteOlderThan(ctx context.Context, opType, status string, days int) error {
	if err := s.opRepo.DeleteOlderThan(ctx, opType, status, time.Now().AddDate(0, 0, -1*days)); err != nil {
		return errors.Wrapf(err, "while deleting Operations of type %s and status %s older than %d", opType, status, days)
	}

	return nil
}
