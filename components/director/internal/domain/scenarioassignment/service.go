package scenarioassignment

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

//go:generate mockery -name=Repository -output=automock -outpkg=automock -case=underscore
type Repository interface {
	Create(ctx context.Context, model model.AutomaticScenarioAssignment) error
	GetForSelector(ctx context.Context, in model.LabelSelector, tenantID string) ([]*model.AutomaticScenarioAssignment, error)
	GetForScenarioName(ctx context.Context, tenantID, scenarioName string) (model.AutomaticScenarioAssignment, error)
	DeleteForSelector(ctx context.Context, tenantID string, selector model.LabelSelector) error
}

//go:generate mockery -name=ScenariosDefService -output=automock -outpkg=automock -case=underscore
type ScenariosDefService interface {
	EnsureScenariosLabelDefinitionExists(ctx context.Context, tenant string) error
	GetAvailableScenarios(ctx context.Context, tenant string) ([]string, error)
}

func NewService(repo Repository, scenarioDefSvc ScenariosDefService) *service {
	return &service{
		repo:            repo,
		scenariosDefSvc: scenarioDefSvc,
	}
}

type service struct {
	repo            Repository
	scenariosDefSvc ScenariosDefService
}

func (s *service) Create(ctx context.Context, in model.AutomaticScenarioAssignment) (model.AutomaticScenarioAssignment, error) {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return model.AutomaticScenarioAssignment{}, err
	}

	in.Tenant = tenantID
	if err := s.validateThatScenarioExists(ctx, in); err != nil {
		return model.AutomaticScenarioAssignment{}, err
	}
	err = s.repo.Create(ctx, in)
	switch {
	case err == nil:
		return in, nil
	case apperrors.IsNotUnique(err):
		return model.AutomaticScenarioAssignment{}, errors.New("a given scenario already has an assignment")
	default:
		return model.AutomaticScenarioAssignment{}, errors.Wrap(err, "while persisting Assignment")
	}
}

func (s *service) validateThatScenarioExists(ctx context.Context, in model.AutomaticScenarioAssignment) error {
	availableScenarios, err := s.getAvailableScenarios(ctx, in.Tenant)
	if err != nil {
		return err
	}

	for _, av := range availableScenarios {
		if av == in.ScenarioName {
			return nil
		}
	}

	return fmt.Errorf("scenario `%s` does not exist", in.ScenarioName)
}

func (s *service) getAvailableScenarios(ctx context.Context, tenant string) ([]string, error) {
	if err := s.scenariosDefSvc.EnsureScenariosLabelDefinitionExists(ctx, tenant); err != nil {
		return nil, errors.Wrap(err, "while ensuring that `scenarios` label definition exist")
	}

	out, err := s.scenariosDefSvc.GetAvailableScenarios(ctx, tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while getting available scenarios")
	}
	return out, nil
}

func (s *service) GetForSelector(ctx context.Context, in model.LabelSelector) ([]*model.AutomaticScenarioAssignment, error) {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	assignments, err := s.repo.GetForSelector(ctx, in, tenantID)
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

func (s *service) DeleteForSelector(ctx context.Context, selector model.LabelSelector) error {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	err = s.repo.DeleteForSelector(ctx, tenantID, selector)
	if err != nil {
		return errors.Wrap(err, "while deleting the Assignments")
	}

	return nil
}
