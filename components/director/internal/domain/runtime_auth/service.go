package runtime_auth

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenant"

	"github.com/pkg/errors"
)

//go:generate mockery -name=Repository -output=automock -outpkg=automock -case=underscore
type Repository interface {
	Get(ctx context.Context, tenant string, apiID string, runtimeID string) (*model.RuntimeAuth, error)
	GetOrDefault(ctx context.Context, tenant string, apiID string, runtimeID string) (*model.RuntimeAuth, error)
	ListForAllRuntimes(ctx context.Context, tenant string, apiID string) ([]model.RuntimeAuth, error)
	Upsert(ctx context.Context, item model.RuntimeAuth) error
	Delete(ctx context.Context, tenant string, apiID string, runtimeID string) error
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

type service struct {
	repo       Repository
	uidService UIDService
}

func NewService(repo Repository, uidService UIDService) *service {
	return &service{
		repo:       repo,
		uidService: uidService,
	}
}

func (s *service) Get(ctx context.Context, apiID string, runtimeID string) (*model.RuntimeAuth, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	ra, err := s.repo.Get(ctx, tnt, apiID, runtimeID)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching Runtime Auth")
	}

	return ra, nil
}

func (s *service) GetOrDefault(ctx context.Context, apiID string, runtimeID string) (*model.RuntimeAuth, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	ra, err := s.repo.GetOrDefault(ctx, tnt, apiID, runtimeID)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching Runtime Auth")
	}

	return ra, nil
}

func (s *service) ListForAllRuntimes(ctx context.Context, apiID string) ([]model.RuntimeAuth, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	ras, err := s.repo.ListForAllRuntimes(ctx, tnt, apiID)
	if err != nil {
		return nil, errors.Wrap(err, "while listing Runtime Auths")
	}

	return ras, nil
}

func (s *service) Set(ctx context.Context, apiID string, runtimeID string, in model.AuthInput) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	id := s.uidService.Generate()

	newAuth := &model.RuntimeAuth{
		ID:        &id,
		TenantID:  tnt,
		RuntimeID: runtimeID,
		APIDefID:  apiID,
		Value:     in.ToAuth(),
	}

	err = s.repo.Upsert(ctx, *newAuth)

	return errors.Wrap(err, "while setting Runtime Auth")
}

func (s *service) Delete(ctx context.Context, apiID string, runtimeID string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	err = s.repo.Delete(ctx, tnt, apiID, runtimeID)

	return errors.Wrap(err, "while deleting Runtime Auth")
}
