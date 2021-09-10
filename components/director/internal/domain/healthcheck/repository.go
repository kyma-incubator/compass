package healthcheck

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type inMemoryRepository struct {
	store map[string]*model.Runtime
}

// NewRepository missing godoc
func NewRepository() *inMemoryRepository {
	return &inMemoryRepository{store: make(map[string]*model.Runtime)}
}
