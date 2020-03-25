package scenarioassignment

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

//go:generate mockery -name=Repository -output=automock -outpkg=automock -case=underscore
type Repository interface {
	Create(ctx context.Context, model model.AutomaticScenarioAssignment) error
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
	err := s.repo.Create(ctx, in)
	if err != nil {
		return model.AutomaticScenarioAssignment{}, errors.Wrap(err, "while persisting Assignment")
	}
	return in, nil
}
