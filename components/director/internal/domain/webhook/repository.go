package webhook

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type inMemoryRepository struct {
	store map[string]*model.ApplicationWebhook
}

func NewRepository() *inMemoryRepository {
	return &inMemoryRepository{store: make(map[string]*model.ApplicationWebhook)}
}

func (inMemoryRepository) ListByApplicationID(applicationID string) ([]*model.ApplicationWebhook, error) {
	panic("implement me")
}

func (inMemoryRepository) CreateMany(items []*model.ApplicationWebhook) error {
	panic("implement me")
}

func (inMemoryRepository) DeleteAllByApplicationID(id string) error {
	panic("implement me")
}
