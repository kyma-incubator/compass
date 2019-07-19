package runtime

import (
	"context"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/pkg/errors"
)

//go:generate mockery -name=RuntimeRepository -output=automock -outpkg=automock -case=underscore
type RuntimeRepository interface {
	GetByID(tenant, id string) (*model.Runtime, error)
	List(tenant string, filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.RuntimePage, error)
	Create(item *model.Runtime) error
	Update(item *model.Runtime) error
	Delete(item *model.Runtime) error
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

type service struct {
	repo       RuntimeRepository
	uidService UIDService
}

func NewService(repo RuntimeRepository, uidService UIDService) *service {
	return &service{repo: repo, uidService: uidService}
}

func (s *service) List(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.RuntimePage, error) {
	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	return s.repo.List(rtmTenant, filter, pageSize, cursor)
}

func (s *service) Get(ctx context.Context, id string) (*model.Runtime, error) {
	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	runtime, err := s.repo.GetByID(rtmTenant, id)
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

	runtimeTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", errors.Wrapf(err, "while loading tenant from context")
	}
	id := s.uidService.Generate()
	rtm := in.ToRuntime(id, runtimeTenant)

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

	err = s.repo.Create(rtm)
	if err != nil {
		return "", err
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

	rtm = in.ToRuntime(id, rtm.Tenant)

	err = s.repo.Update(rtm)
	if err != nil {
		return errors.Wrap(err, "while updating Runtime")
	}

	return nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	rtm, err := s.Get(ctx, id)
	if err != nil {
		return errors.Wrap(err, "while getting Runtime")
	}

	return s.repo.Delete(rtm)
}

func (s *service) AddLabel(ctx context.Context, runtimeID string, key string, values []string) error {
	rtm, err := s.Get(ctx, runtimeID)
	if err != nil {
		return errors.Wrap(err, "while getting Runtime")
	}

	rtm.AddLabel(key, values)

	err = s.repo.Update(rtm)
	if err != nil {
		return errors.Wrapf(err, "while updating Runtime")
	}

	return nil
}

func (s *service) DeleteLabel(ctx context.Context, runtimeID string, key string, values []string) error {
	rtm, err := s.Get(ctx, runtimeID)
	if err != nil {
		return errors.Wrap(err, "while getting Runtime")
	}

	err = rtm.DeleteLabel(key, values)
	if err != nil {
		return errors.Wrapf(err, "while deleting label with key %s", key)
	}

	err = s.repo.Update(rtm)
	if err != nil {
		return errors.Wrapf(err, "while updating Runtime")
	}

	return nil
}
