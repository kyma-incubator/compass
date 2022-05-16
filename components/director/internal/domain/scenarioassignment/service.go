package scenarioassignment

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
)

// Repository missing godoc
//go:generate mockery --name=Repository --output=automock --outpkg=automock --case=underscore --disable-version-string
type Repository interface {
	Create(ctx context.Context, model model.AutomaticScenarioAssignment) error
	ListForTargetTenant(ctx context.Context, tenantID string, targetTenantID string) ([]*model.AutomaticScenarioAssignment, error)
	GetForScenarioName(ctx context.Context, tenantID, scenarioName string) (model.AutomaticScenarioAssignment, error)
	ListAll(ctx context.Context, tenantID string) ([]*model.AutomaticScenarioAssignment, error)
	List(ctx context.Context, tenant string, pageSize int, cursor string) (*model.AutomaticScenarioAssignmentPage, error)
	DeleteForTargetTenant(ctx context.Context, tenantID string, targetTenantID string) error
	DeleteForScenarioName(ctx context.Context, tenantID string, scenarioName string) error
}

// ScenariosDefService missing godoc
//go:generate mockery --name=ScenariosDefService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ScenariosDefService interface {
	EnsureScenariosLabelDefinitionExists(ctx context.Context, tenantID string) error
	GetAvailableScenarios(ctx context.Context, tenantID string) ([]string, error)
}

// AssignmentEngine missing godoc
//go:generate mockery --name=AssignmentEngine --output=automock --outpkg=automock --case=underscore --disable-version-string
type AssignmentEngine interface {
	EnsureScenarioAssigned(ctx context.Context, in model.AutomaticScenarioAssignment) error
	RemoveAssignedScenario(ctx context.Context, in model.AutomaticScenarioAssignment) error
	RemoveAssignedScenarios(ctx context.Context, in []*model.AutomaticScenarioAssignment) error
}

// NewService missing godoc
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

// Create missing godoc
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

	return apperrors.NewNotFoundError(resource.AutomaticScenarioAssigment, in.ScenarioName)
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

// ListForSelector missing godoc
func (s *service) ListForTargetTenant(ctx context.Context, targetTenantInternalID string) ([]*model.AutomaticScenarioAssignment, error) {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	assignments, err := s.repo.ListForTargetTenant(ctx, tenantID, targetTenantInternalID)
	if err != nil {
		return nil, errors.Wrap(err, "while getting the assignments")
	}
	return assignments, nil
}

// GetForScenarioName missing godoc
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

// List missing godoc
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

// DeleteManyForSameSelector missing godoc
func (s *service) DeleteManyForSameTargetTenant(ctx context.Context, in []*model.AutomaticScenarioAssignment) error {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	targetTenant, err := s.ensureSameTargetTenant(in)
	if err != nil {
		return errors.Wrap(err, "while ensuring input is valid")
	}

	if err = s.engineSvc.RemoveAssignedScenarios(ctx, in); err != nil {
		return errors.Wrap(err, "while unassigning scenario from runtimes")
	}

	if err = s.repo.DeleteForTargetTenant(ctx, tenantID, targetTenant); err != nil {
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

	if err = s.engineSvc.RemoveAssignedScenario(ctx, in); err != nil {
		return errors.Wrap(err, "while unassigning scenario from runtimes")
	}

	if err = s.repo.DeleteForScenarioName(ctx, tenantID, in.ScenarioName); err != nil {
		return errors.Wrap(err, "while deleting the Assignment")
	}

	return nil
}

func (s *service) ensureSameTargetTenant(in []*model.AutomaticScenarioAssignment) (string, error) {
	if len(in) == 0 || in[0] == nil {
		return "", apperrors.NewInternalError("expected at least one item in Assignments slice")
	}

	targetTenant := in[0].TargetTenantID

	for _, item := range in {
		if item != nil && item.TargetTenantID != targetTenant {
			return "", apperrors.NewInternalError("all input items have to have the same target tenant")
		}
	}

	return targetTenant, nil
}
