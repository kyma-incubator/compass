package runtime

import (
	"context"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/uid"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"

	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/pkg/errors"
)

//go:generate mockery -name=RuntimeRepository -output=automock -outpkg=automock -case=underscore
type RuntimeRepository interface {
	GetByID(id string) (*model.Runtime, error)
	List(filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.RuntimePage, error)
	Create(item *model.Runtime) error
	Update(item *model.Runtime) error
	Delete(item *model.Runtime) error
}

type service struct {
	repo RuntimeRepository
}

func NewService(repo RuntimeRepository) *service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.RuntimePage, error) {
	return s.repo.List(filter, pageSize, cursor)
}

func (s *service) Get(ctx context.Context, id string) (*model.Runtime, error) {
	runtime, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Runtime with ID %s", id)
	}

	return runtime, nil
}

func (s *service) Create(ctx context.Context, in model.RuntimeInput) (string, error) {
	id := uid.Generate()
	runtimeTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", errors.Wrapf(err, "while loading tenant from context")
	}

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
	rtm, err := s.Get(ctx, id)
	if err != nil {
		return errors.Wrap(err, "while getting Runtime")
	}

	rtm.Name = in.Name
	rtm.Description = in.Description
	rtm.Labels = in.Labels
	rtm.Annotations = in.Annotations

	err = s.repo.Update(rtm)
	if err != nil {
		return errors.Wrapf(err, "while updating Runtime")
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

func (s *service) AddAnnotation(ctx context.Context, runtimeID string, key string, value interface{}) error {
	rtm, err := s.Get(ctx, runtimeID)
	if err != nil {
		return errors.Wrap(err, "while getting Runtime")
	}

	err = rtm.AddAnnotation(key, value)
	if err != nil {
		return errors.Wrapf(err, "while adding new annotation %s", key)
	}

	err = s.repo.Update(rtm)
	if err != nil {
		return errors.Wrapf(err, "while updating Runtime")
	}

	return nil
}

func (s *service) DeleteAnnotation(ctx context.Context, runtimeID string, key string) error {
	rtm, err := s.Get(ctx, runtimeID)
	if err != nil {
		return errors.Wrap(err, "while getting Runtime")
	}

	err = rtm.DeleteAnnotation(key)
	if err != nil {
		return errors.Wrapf(err, "while deleting annotation with key %s", key)
	}

	err = s.repo.Update(rtm)
	if err != nil {
		return errors.Wrapf(err, "while updating Runtime with ID %s", runtimeID)
	}

	return nil
}
