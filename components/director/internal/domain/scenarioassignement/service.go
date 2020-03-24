package scenarioassignement

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

//go:generate mockery -name=AutomaticScenarioAssignmentReposity -output=automock -outpkg=automock -case=underscore
type AutomaticScenarioAssignmentReposity interface {
	GetForSelector(ctx context.Context, in model.LabelSelector, tenant string) ([]*model.AutomaticScenarioAssignment, error)
}

type service struct {
	repo AutomaticScenarioAssignmentReposity
}

func NewService(repo AutomaticScenarioAssignmentReposity) *service {
	return &service{repo: repo}
}

func (s *service) GetForSelector(ctx context.Context, in model.LabelSelector, tenant string) ([]*model.AutomaticScenarioAssignment, error) {
	assignments, err := s.repo.GetForSelector(ctx, in, tenant)
	if err != nil {
		return nil, err
	}
	return assignments, nil
}
