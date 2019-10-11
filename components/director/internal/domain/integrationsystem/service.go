package integrationsystem

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

//go:generate mockery -name=IntegrationSystemRepository -output=automock -outpkg=automock -case=underscore
type IntegrationSystemRepository interface {
	Create(ctx context.Context, item model.IntegrationSystem) error
	Get(ctx context.Context, id string) (*model.IntegrationSystem, error)
	Exists(ctx context.Context, id string) (bool, error)
	List(ctx context.Context, pageSize int, cursor string) (model.IntegrationSystemPage, error)
	Update(ctx context.Context, model model.IntegrationSystem) error
	Delete(ctx context.Context, id string) error
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

type service struct {
	intSysRepo IntegrationSystemRepository

	uidService UIDService
}

func NewService(intSysRepo IntegrationSystemRepository, uidService UIDService) *service {
	return &service{
		intSysRepo: intSysRepo,
		uidService: uidService,
	}
}

func (s *service) Create(ctx context.Context, in model.IntegrationSystemInput) (string, error) {
	id := s.uidService.Generate()
	intSys := in.ToIntegrationSystem(id)

	err := s.intSysRepo.Create(ctx, intSys)
	if err != nil {
		return "", errors.Wrap(err, "while creating Integration System")
	}

	return id, nil
}

func (s *service) Get(ctx context.Context, id string) (*model.IntegrationSystem, error) {
	intSys, err := s.intSysRepo.Get(ctx, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Integration System with ID %s", id)
	}

	return intSys, nil
}

func (s *service) Exists(ctx context.Context, id string) (bool, error) {
	exist, err := s.intSysRepo.Exists(ctx, id)
	if err != nil {
		return false, errors.Wrapf(err, "while getting Integration System with ID %s", id)
	}

	return exist, nil
}

func (s *service) List(ctx context.Context, pageSize int, cursor string) (model.IntegrationSystemPage, error) {
	if pageSize < 1 || pageSize > 100 {
		return model.IntegrationSystemPage{}, errors.New("page size must be between 1 and 100")
	}

	return s.intSysRepo.List(ctx, pageSize, cursor)
}

func (s *service) Update(ctx context.Context, id string, in model.IntegrationSystemInput) error {
	intSys := in.ToIntegrationSystem(id)

	err := s.intSysRepo.Update(ctx, intSys)
	if err != nil {
		return errors.Wrapf(err, "while updating Integration System with ID %s", id)
	}

	return nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	err := s.intSysRepo.Delete(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "while deleting Integration System with ID %s", id)
	}

	return nil
}
