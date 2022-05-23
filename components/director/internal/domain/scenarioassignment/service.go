package scenarioassignment

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
)

// Repository missing godoc
//go:generate mockery --name=Repository --output=automock --outpkg=automock --case=underscore --disable-version-string
type Repository interface {
	ListForTargetTenant(ctx context.Context, tenantID string, targetTenantID string) ([]*model.AutomaticScenarioAssignment, error)
	GetForScenarioName(ctx context.Context, tenantID, scenarioName string) (model.AutomaticScenarioAssignment, error)
	ListAll(ctx context.Context, tenantID string) ([]*model.AutomaticScenarioAssignment, error)
	List(ctx context.Context, tenant string, pageSize int, cursor string) (*model.AutomaticScenarioAssignmentPage, error)
}

// ScenariosDefService missing godoc
//go:generate mockery --name=ScenariosDefService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ScenariosDefService interface {
	EnsureScenariosLabelDefinitionExists(ctx context.Context, tenantID string) error
	GetAvailableScenarios(ctx context.Context, tenantID string) ([]string, error)
}

// NewService missing godoc
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

// ListForTargetTenant missing godoc
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
