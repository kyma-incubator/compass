package scenarioassignment

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

//go:generate mockery -name=Repository -output=automock -outpkg=automock -case=underscore
type Repository interface {
	Create(ctx context.Context, model model.AutomaticScenarioAssignment) error
	GetByScenarioName(ctx context.Context, tnt, scenarioName string) (model.AutomaticScenarioAssignment, error)
}

func NewService(repo Repository) *service {
	return &service{
		repo: repo,
	}
}

type service struct {
	repo Repository
}

func (s *service) Create(ctx context.Context, in model.AutomaticScenarioAssignment) (model.AutomaticScenarioAssignment, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return model.AutomaticScenarioAssignment{}, err
	}

	in.Tenant = tnt

	err = s.repo.Create(ctx, in)
	if err != nil {
		return model.AutomaticScenarioAssignment{}, errors.Wrap(err, "while persisting Assignment")
	}
	return in, nil
}

func (s *service) GetByScenarioName(ctx context.Context, scenarioName string) (model.AutomaticScenarioAssignment, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return model.AutomaticScenarioAssignment{}, err
	}

	sa, err := s.repo.GetByScenarioName(ctx, tnt, scenarioName)
	if err != nil {
		return model.AutomaticScenarioAssignment{}, errors.Wrap(err, "while getting Assignment")
	}
	return sa, nil
}
