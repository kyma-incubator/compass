package scenarioassignment

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
)

//go:generate mockery -name=Repository -output=automock -outpkg=automock -case=underscore
type Repository interface {
	Create(ctx context.Context, model model.AutomaticScenarioAssignment) error
	ListForSelector(ctx context.Context, in model.LabelSelector, tenantID string) ([]*model.AutomaticScenarioAssignment, error)
	GetForScenarioName(ctx context.Context, tenantID, scenarioName string) (model.AutomaticScenarioAssignment, error)
	List(ctx context.Context, tenant string, pageSize int, cursor string) (*model.AutomaticScenarioAssignmentPage, error)
	DeleteForSelector(ctx context.Context, tenantID string, selector model.LabelSelector) error
	DeleteForScenarioName(ctx context.Context, tenantID string, scenarioName string) error
}

//go:generate mockery -name=ScenariosDefService -output=automock -outpkg=automock -case=underscore
type ScenariosDefService interface {
	EnsureScenariosLabelDefinitionExists(ctx context.Context, tenantID string) error
	GetAvailableScenarios(ctx context.Context, tenantID string) ([]string, error)
}

//go:generate mockery -name=AssignmentEngine -output=automock -outpkg=automock -case=underscore
type AssignmentEngine interface {
	EnsureScenarioAssigned(ctx context.Context, in model.AutomaticScenarioAssignment) error
	RemoveAssignedScenario(ctx context.Context, in model.AutomaticScenarioAssignment) error
	RemoveAssignedScenarios(ctx context.Context, in []*model.AutomaticScenarioAssignment) error
}

func NewService(repo Repository, scenarioDefSvc ScenariosDefService, engineSvc AssignmentEngine) *service {
	return &service{
		repo:            repo,
		scenariosDefSvc: scenarioDefSvc,
		engineSvc:       engineSvc,
	}
}

type service struct {
	repo            Repository
	scenariosDefSvc ScenariosDefService
	engineSvc       AssignmentEngine
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
	if err != nil {
		if apperrors.IsNotUniqueError(err) {
			return model.AutomaticScenarioAssignment{}, apperrors.NewInvalidOperationError("a given scenario already has an assignment")
		}

		return model.AutomaticScenarioAssignment{}, errors.Wrap(err, "while persisting Assignment")
	}

	err = s.engineSvc.EnsureScenarioAssigned(ctx, in)
	if err != nil {
		return model.AutomaticScenarioAssignment{}, errors.Wrap(err, "while assigning scenario to runtimes matching selector")
	}

	return in, nil
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

func (s *service) getAvailableScenarios(ctx context.Context, tenantID string) ([]string, error) {
	if err := s.scenariosDefSvc.EnsureScenariosLabelDefinitionExists(ctx, tenantID); err != nil {
		return nil, errors.Wrap(err, "while ensuring that `scenarios` label definition exist")
	}

	out, err := s.scenariosDefSvc.GetAvailableScenarios(ctx, tenantID)
	if err != nil {
		return nil, errors.Wrap(err, "while getting available scenarios")
	}
	return out, nil
}

func (s *service) ListForSelector(ctx context.Context, in model.LabelSelector) ([]*model.AutomaticScenarioAssignment, error) {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	assignments, err := s.repo.ListForSelector(ctx, in, tenantID)
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

func (s *service) List(ctx context.Context, pageSize int, cursor string) (*model.AutomaticScenarioAssignmentPage, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if pageSize < 1 || pageSize > 200 {
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 200")
	}

	return s.repo.List(ctx, tnt, pageSize, cursor)
}

func (s *service) DeleteManyForSameSelector(ctx context.Context, in []*model.AutomaticScenarioAssignment) error {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	selector, err := s.ensureSameSelector(in)
	if err != nil {
		return errors.Wrap(err, "while ensuring input is valid")
	}

	err = s.engineSvc.RemoveAssignedScenarios(ctx, in)
	if err != nil {
		return errors.Wrap(err, "while unassigning scenario from runtimes")
	}

	err = s.repo.DeleteForSelector(ctx, tenantID, selector)
	if err != nil {
		return errors.Wrap(err, "while deleting the Assignments")
	}

	return nil
}

// Delete deletes the assignment for a given scenario in a scope of a tenant
func (s *service) Delete(ctx context.Context, in model.AutomaticScenarioAssignment) error {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "while loading tenant from context")
	}

	err = s.engineSvc.RemoveAssignedScenario(ctx, in)
	if err != nil {
		return errors.Wrap(err, "while unassigning scenario from runtimes")
	}

	err = s.repo.DeleteForScenarioName(ctx, tenantID, in.ScenarioName)
	if err != nil {
		return errors.Wrap(err, "while deleting the Assignment")
	}

	return nil
}

func (s *service) ensureSameSelector(in []*model.AutomaticScenarioAssignment) (model.LabelSelector, error) {
	if in == nil || len(in) == 0 || in[0] == nil {
		return model.LabelSelector{}, apperrors.NewInternalError("expected at least one item in Assignments slice")
	}

	selector := in[0].Selector

	for _, item := range in {
		if item != nil && item.Selector != selector {
			return model.LabelSelector{}, apperrors.NewInternalError("all input items have to have the same selector")
		}
	}

	return selector, nil
}
