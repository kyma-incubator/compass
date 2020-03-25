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
	GetForSelector(ctx context.Context, in model.LabelSelector, tenant string) ([]*model.AutomaticScenarioAssignment, error)
	GetForScenarioName(ctx context.Context, tenantID, scenarioName string) (model.AutomaticScenarioAssignment, error)
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
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return model.AutomaticScenarioAssignment{}, err
	}

	in.Tenant = tenantID

	err = s.repo.Create(ctx, in)
	if err != nil {
		return model.AutomaticScenarioAssignment{}, errors.Wrap(err, "while persisting Assignment")
	}
	return in, nil
}

func (s *service) GetForSelector(ctx context.Context, in model.LabelSelector) ([]*model.AutomaticScenarioAssignment, error) {
	tenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	assignments, err := s.repo.GetForSelector(ctx, in, tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while getting the assignments")
	}
	return assignments, nil
}
func (s *service) GetForScenarioName(ctx context.Context, scenarioName string) (model.AutomaticScenarioAssignment, error) {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return model.AutomaticScenarioAssignment{}, err
	}

	sa, err := s.repo.GetForScenarioName(ctx, tenantID, scenarioName)
	if err != nil {
		return model.AutomaticScenarioAssignment{}, errors.Wrap(err, "while getting Assignment")
	}
	return sa, nil
}
