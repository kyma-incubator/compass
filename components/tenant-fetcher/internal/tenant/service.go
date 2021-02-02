package tenant

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/tenant-fetcher/internal/model"
	"github.com/pkg/errors"
)

//go:generate mockery -name=Converter -output=automock -outpkg=automock -case=underscore
type TenantService interface {
	Create(ctx context.Context, item model.TenantModel) error
	DeleteByTenant(ctx context.Context, tenantId string) error
}

type service struct {
	repository TenantRepository
	transact   persistence.Transactioner
}

func NewService(tenant TenantRepository, transact persistence.Transactioner) *service {
	return &service{
		repository: tenant,
		transact:   transact,
	}
}

func (s *service) Create(ctx context.Context, item model.TenantModel) error {
	tx, err := s.transact.Begin()
	if err != nil {
		return errors.Wrapf(err, "while beginning db transaction")
	}

	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err := s.repository.Create(ctx, item); err != nil {
		return errors.Wrap(err, "while creating tenant")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "while committing transaction")
	}

	return nil
}

func (s *service) DeleteByTenant(ctx context.Context, tenantId string) error {
	tx, err := s.transact.Begin()
	if err != nil {
		return errors.Wrapf(err, "while beginning db transaction")
	}

	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err := s.repository.DeleteByTenant(ctx, tenantId); err != nil {
		return errors.Wrap(err, "while creating tenant")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "while committing transaction")
	}

	return nil
}
