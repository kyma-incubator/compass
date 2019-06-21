package api

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type inMemoryRepository struct {
	store map[string]*model.APIDefinition
}

func NewRepository() *inMemoryRepository {
	return &inMemoryRepository{store: make(map[string]*model.APIDefinition)}
}

func (inMemoryRepository) ListByApplicationID(applicationID string) ([]*model.APIDefinition, error) {
	panic("implement me")
}

func (inMemoryRepository) CreateMany(items []*model.APIDefinition) error {
	panic("implement me")
}

func (inMemoryRepository) DeleteAllByApplicationID(id string) error {
	panic("implement me")
}
