package runtime

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/pkg/errors"
)

//go:generate mockery -name=RuntimeRepository -output=automock -outpkg=automock -case=underscore
type RuntimeRepository interface {
	Exists(ctx context.Context, tenant, id string) (bool, error)
	GetByID(ctx context.Context, tenant, id string) (*model.Runtime, error)
	List(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.RuntimePage, error)
	Create(ctx context.Context, item *model.Runtime) error
	Update(ctx context.Context, item *model.Runtime) error
	Delete(ctx context.Context, id string) error
}

//go:generate mockery -name=LabelRepository -output=automock -outpkg=automock -case=underscore
type LabelRepository interface {
	GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
	List(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error)
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
	repo      RuntimeRepository
	labelRepo LabelRepository

	labelUpsertService LabelUpsertService
	uidService         UIDService
}

func NewService(repo RuntimeRepository, labelRepo LabelRepository, labelUpsertService LabelUpsertService, uidService UIDService) *service {
	return &service{repo: repo, labelRepo: labelRepo, labelUpsertService: labelUpsertService, uidService: uidService}
}

func (s *service) List(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.RuntimePage, error) {
	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	return s.repo.List(ctx, rtmTenant, filter, pageSize, cursor)
}

func (s *service) Get(ctx context.Context, id string) (*model.Runtime, error) {
	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	runtime, err := s.repo.GetByID(ctx, rtmTenant, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Runtime with ID %s", id)
	}

	return runtime, nil
}

func (s *service) Create(ctx context.Context, in model.RuntimeInput) (string, error) {
	err := in.Validate()
	if err != nil {
		return "", errors.Wrap(err, "while validating Runtime input")
	}

	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", errors.Wrapf(err, "while loading tenant from context")
	}
	id := s.uidService.Generate()
	rtm := in.ToRuntime(id, rtmTenant)

	// TODO: Generate AgentAuth: https://github.com/kyma-incubator/compass/issues/91
	rtm.AgentAuth = &model.Auth{
		Credential: model.CredentialData{
			Basic: &model.BasicCredentialData{
				Username: "foo",
				Password: "bar",
			},
		},
	}
	rtm.Status = &model.RuntimeStatus{
		Condition: model.RuntimeStatusConditionInitial,
		Timestamp: time.Now(),
	}

	err = s.repo.Create(ctx, rtm)
	if err != nil {
		return "", errors.Wrapf(err, "while creating Runtime")
	}

	err = s.labelUpsertService.UpsertMultipleLabels(ctx, rtmTenant, model.RuntimeLabelableObject, id, in.Labels)
	if err != nil {
		return id, errors.Wrapf(err, "while creating multiple labels for Runtime")
	}

	return id, nil
}

func (s *service) Update(ctx context.Context, id string, in model.RuntimeInput) error {
	err := in.Validate()
	if err != nil {
		return errors.Wrap(err, "while validating Runtime input")
	}

	rtm, err := s.Get(ctx, id)
	if err != nil {
		return errors.Wrap(err, "while getting Runtime")
	}

	currentStatuts := rtm.Status

	rtm = in.ToRuntime(id, rtm.Tenant)

	if rtm.Status.Condition == "" {
		rtm.Status = currentStatuts
	}

	err = s.repo.Update(ctx, rtm)
	if err != nil {
		return errors.Wrap(err, "while updating Runtime")
	}

	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	err = s.labelRepo.DeleteAll(ctx, rtmTenant, model.RuntimeLabelableObject, id)
	if err != nil {
		return errors.Wrapf(err, "while deleting all labels for Runtime")
	}

	err = s.labelUpsertService.UpsertMultipleLabels(ctx, rtmTenant, model.RuntimeLabelableObject, id, in.Labels)
	if err != nil {
		return errors.Wrapf(err, "while creating multiple labels for Runtime")
	}

	return nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	exists, err := s.repo.Exists(ctx, rtmTenant, id)
	if err != nil {
		return errors.Wrap(err, "while getting Runtime")
	}
	if !exists {
		return fmt.Errorf("Runtime %s doesn't exist", id)
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "while deleting Runtime")
	}

	// All labels are deleted (cascade delete)

	return nil
}

func (s *service) SetLabel(ctx context.Context, labelInput *model.LabelInput) error {
	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	rtmExists, err := s.repo.Exists(ctx, rtmTenant, labelInput.ObjectID)
	if err != nil {
		return errors.Wrap(err, "while checking Runtime existence")
	}
	if !rtmExists {
		return fmt.Errorf("Runtime with ID %s doesn't exist", labelInput.ObjectID)
	}

	err = s.labelUpsertService.UpsertLabel(ctx, rtmTenant, labelInput)
	if err != nil {
		return errors.Wrapf(err, "while creating label for Runtime")
	}

	return nil
}

func (s *service) GetLabel(ctx context.Context, runtimeID string, key string) (*model.Label, error) {
	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	rtmExists, err := s.repo.Exists(ctx, rtmTenant, runtimeID)
	if err != nil {
		return nil, errors.Wrap(err, "while checking Runtime existence")
	}
	if !rtmExists {
		return nil, fmt.Errorf("Runtime with ID %s doesn't exist", runtimeID)
	}

	label, err := s.labelRepo.GetByKey(ctx, rtmTenant, model.RuntimeLabelableObject, runtimeID, key)
	if err != nil {
		return nil, errors.Wrap(err, "while getting label for Runtime")
	}

	return label, nil
}

func (s *service) ListLabels(ctx context.Context, runtimeID string) (map[string]*model.Label, error) {
	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	rtmExists, err := s.repo.Exists(ctx, rtmTenant, runtimeID)
	if err != nil {
		return nil, errors.Wrap(err, "while checking Runtime existence")
	}

	if !rtmExists {
		return nil, fmt.Errorf("Runtime with ID %s doesn't exist", runtimeID)
	}

	labels, err := s.labelRepo.List(ctx, rtmTenant, model.RuntimeLabelableObject, runtimeID)
	if err != nil {
		return nil, errors.Wrap(err, "while getting label for Runtime")
	}

	return labels, nil
}

func (s *service) DeleteLabel(ctx context.Context, runtimeID string, key string) error {
	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	rtmExists, err := s.repo.Exists(ctx, rtmTenant, runtimeID)
	if err != nil {
		return errors.Wrap(err, "while checking Runtime existence")
	}
	if !rtmExists {
		return fmt.Errorf("Runtime with ID %s doesn't exist", runtimeID)
	}

	err = s.labelRepo.Delete(ctx, rtmTenant, model.RuntimeLabelableObject, runtimeID, key)
	if err != nil {
		return errors.Wrapf(err, "while deleting Runtime label")
	}

	return nil
}
