package document

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type inMemoryRepository struct {
	store map[string]*model.Document
}

func NewRepository() *inMemoryRepository {
	return &inMemoryRepository{store: make(map[string]*model.Document)}
}

func (inMemoryRepository) ListByApplicationID(applicationID string) ([]*model.Document, error) {
	panic("implement me")
}

func (inMemoryRepository) CreateMany(items []*model.Document) error {
	panic("implement me")
}

func (inMemoryRepository) DeleteAllByApplicationID(id string) error {
	panic("implement me")
}

