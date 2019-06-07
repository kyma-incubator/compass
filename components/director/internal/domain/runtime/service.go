package runtime

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/pkg/errors"
)

//go:generate mockery -name=RuntimeRepository -output=automock -outpkg=automock -case=underscore
type RuntimeRepository interface {
	GetByID(id string) (*model.Runtime, error)
	List(filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*RuntimePage, error)
	Create(item *model.Runtime) error
	Update(item *model.Runtime) error
	Delete(item *model.Runtime) error
}

type Service struct {
	repo RuntimeRepository
}

func NewService(repo RuntimeRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*RuntimePage, error) {
	return s.repo.List(filter, pageSize, cursor)
}

func (s *Service) Get(ctx context.Context, id string) (*model.Runtime, error) {
	runtime, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Runtime with ID %s", id)
	}

	return runtime, nil
}

func (s *Service) Create(ctx context.Context, in model.RuntimeInput) (string, error) {
	id := uuid.New().String()
	runtimeTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", errors.Wrapf(err, "while loading tenant from context")
	}

	rtm := &model.Runtime{
		ID:          id,
		Name:        in.Name,
		Description: in.Description,
		Tenant:      runtimeTenant,
		Labels:      in.Labels,
		Annotations: in.Annotations,
	}

	err = s.repo.Create(rtm)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (s *Service) Update(ctx context.Context, id string, in model.RuntimeInput) error {
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

func (s *Service) Delete(ctx context.Context, id string) error {
	rtm, err := s.Get(ctx, id)
	if err != nil {
		return errors.Wrap(err, "while getting Runtime")
	}

	return s.repo.Delete(rtm)
}

func (s *Service) AddLabel(ctx context.Context, runtimeID string, key string, values []string) error {
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

func (s *Service) DeleteLabel(ctx context.Context, runtimeID string, key string, values []string) error {
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

func (s *Service) AddAnnotation(ctx context.Context, runtimeID string, key string, value string) error {
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

func (s *Service) DeleteAnnotation(ctx context.Context, runtimeID string, key string) error {
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
