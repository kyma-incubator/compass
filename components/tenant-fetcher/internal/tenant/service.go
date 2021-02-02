package tenant

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/tenant-fetcher/internal/model"
	"github.com/pkg/errors"
)

//go:generate mockery --name=TenantService --output=automock --outpkg=automock --case=underscore
type TenantService interface {
	Create(ctx context.Context, item model.TenantModel) error
	DeleteByTenant(ctx context.Context, tenantId string) error
}

//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore
type UIDService interface {
	Generate() string
}

type service struct {
	repository TenantRepository
	transact   persistence.Transactioner
	uidService UIDService
}

func NewService(tenant TenantRepository, transact persistence.Transactioner, uidService UIDService) *service {
	return &service{
		repository: tenant,
		transact:   transact,
		uidService: uidService,
	}
}

func (s *service) Create(ctx context.Context, item model.TenantModel) error {
	tx, err := s.transact.Begin()
	if err != nil {
		return errors.Wrapf(err, "while beginning db transaction")
	}

	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	id := s.uidService.Generate()

	if err := s.repository.Create(ctx, item, id); err != nil {
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
