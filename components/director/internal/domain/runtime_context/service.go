package runtime_context

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/pkg/errors"
)

//go:generate mockery -name=RuntimeContextRepository -output=automock -outpkg=automock -case=underscore
type RuntimeContextRepository interface {
	Exists(ctx context.Context, tenant, id string) (bool, error)
	GetByID(ctx context.Context, tenant, id string) (*model.RuntimeContext, error)
	GetByFiltersGlobal(ctx context.Context, filter []*labelfilter.LabelFilter) (*model.RuntimeContext, error)
	List(ctx context.Context, runtimeID string, tenant string, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.RuntimeContextPage, error)
	Create(ctx context.Context, item *model.RuntimeContext) error
	Update(ctx context.Context, item *model.RuntimeContext) error
	Delete(ctx context.Context, tenant, id string) error
}

//go:generate mockery -name=LabelRepository -output=automock -outpkg=automock -case=underscore
type LabelRepository interface {
	GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
	ListForObject(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error)
	Delete(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, key string) error
	DeleteAll(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) error
}

//go:generate mockery -name=LabelUpsertService -output=automock -outpkg=automock -case=underscore
type LabelUpsertService interface {
	UpsertMultipleLabels(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, labels map[string]interface{}) error
	UpsertLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) error
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

type service struct {
	repo      RuntimeContextRepository
	labelRepo LabelRepository

	labelUpsertService LabelUpsertService
	uidService         UIDService
}

func NewService(repo RuntimeContextRepository,
	labelRepo LabelRepository,
	labelUpsertService LabelUpsertService,
	uidService UIDService) *service {
	return &service{
		repo:               repo,
		labelRepo:          labelRepo,
		labelUpsertService: labelUpsertService,
		uidService:         uidService,
	}
}

func (s *service) List(ctx context.Context, runtimeID string, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.RuntimeContextPage, error) {
	rtmCtxTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	if pageSize < 1 || pageSize > 200 {
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 200")
	}

	return s.repo.List(ctx, runtimeID, rtmCtxTenant, filter, pageSize, cursor)
}

func (s *service) Get(ctx context.Context, id string) (*model.RuntimeContext, error) {
	rtmCtxTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	runtimeCtx, err := s.repo.GetByID(ctx, rtmCtxTenant, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Runtime Context with ID %s", id)
	}

	return runtimeCtx, nil
}

func (s *service) Exist(ctx context.Context, id string) (bool, error) {
	rtmCtxTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return false, errors.Wrapf(err, "while loading tenant from context")
	}

	exist, err := s.repo.Exists(ctx, rtmCtxTenant, id)
	if err != nil {
		return false, errors.Wrapf(err, "while getting Runtime Context with ID %s", id)
	}

	return exist, nil
}

func (s *service) Create(ctx context.Context, in model.RuntimeContextInput) (string, error) {
	rtmCtxTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", errors.Wrapf(err, "while loading tenant from context")
	}
	id := s.uidService.Generate()
	rtmCtx := in.ToRuntimeContext(id, rtmCtxTenant)

	err = s.repo.Create(ctx, rtmCtx)
	if err != nil {
		return "", errors.Wrapf(err, "while creating Runtime Context")
	}

	/*err = s.scenariosService.EnsureScenariosLabelDefinitionExists(ctx, rtmCtxTenant) TODO: Revisit when scenarios for runtime contexts are introduced
	if err != nil {
		return "", errors.Wrapf(err, "while ensuring Label Definition with key %s exists", model.ScenariosKey)
	}

	scenarios, err := s.scenarioAssignmentEngine.MergeScenariosFromInputLabelsAndAssignments(ctx, in.Labels)
	if err != nil {
		return "", errors.Wrap(err, "while merging scenarios from input and assignments")
	}

	if len(scenarios) > 0 {
		in.Labels[model.ScenariosKey] = scenarios
	} else {
		s.scenariosService.AddDefaultScenarioIfEnabled(&in.Labels)
	}*/

	err = s.labelUpsertService.UpsertMultipleLabels(ctx, rtmCtxTenant, model.RuntimeContextLabelableObject, id, in.Labels)
	if err != nil {
		return id, errors.Wrapf(err, "while creating multiple labels for Runtime Context")
	}

	return id, nil
}

func (s *service) Update(ctx context.Context, id string, in model.RuntimeContextInput) error {
	rtmCtxTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	rtmCtx, err := s.repo.GetByID(ctx, rtmCtxTenant, id)
	if err != nil {
		return errors.Wrapf(err, "while getting Runtime Context with id %s", id)
	}

	rtmCtx = in.ToRuntimeContext(id, rtmCtx.Tenant)

	err = s.repo.Update(ctx, rtmCtx)
	if err != nil {
		return errors.Wrapf(err, "while updating Runtime Context with id %s", id)
	}

	err = s.labelRepo.DeleteAll(ctx, rtmCtxTenant, model.RuntimeContextLabelableObject, id)
	if err != nil {
		return errors.Wrapf(err, "while deleting all labels for Runtime Context with id %s", id)
	}

	if in.Labels == nil {
		return nil
	}

	/*scenarios, err := s.scenarioAssignmentEngine.MergeScenariosFromInputLabelsAndAssignments(ctx, in.Labels) TODO: Revisit when scenarios for runtime contexts are introduced
	if err != nil {
		return errors.Wrap(err, "while merging scenarios from input and assignments")
	}

	if len(scenarios) > 0 {
		in.Labels[model.ScenariosKey] = scenarios
	}*/

	err = s.labelUpsertService.UpsertMultipleLabels(ctx, rtmCtxTenant, model.RuntimeContextLabelableObject, id, in.Labels)
	if err != nil {
		return errors.Wrapf(err, "while creating multiple labels for Runtime Context with id %s", id)
	}

	return nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	err = s.repo.Delete(ctx, rtmTenant, id)
	if err != nil {
		return errors.Wrapf(err, "while deleting Runtime Context with id %s", id)
	}

	// All labels are deleted (cascade delete)

	return nil
}

func (s *service) ListLabels(ctx context.Context, runtimeCtxID string) (map[string]*model.Label, error) {
	rtmCtxTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	rtmCtxExists, err := s.repo.Exists(ctx, rtmCtxTenant, runtimeCtxID)
	if err != nil {
		return nil, errors.Wrapf(err, "while checking Runtime Context existence with id %s", runtimeCtxID)
	}

	if !rtmCtxExists {
		return nil, fmt.Errorf("runtime Context with ID %s doesn't exist", runtimeCtxID)
	}

	labels, err := s.labelRepo.ListForObject(ctx, rtmCtxTenant, model.RuntimeContextLabelableObject, runtimeCtxID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting label for Runtime Context with id %s", runtimeCtxID)
	}

	return labels, nil
}
