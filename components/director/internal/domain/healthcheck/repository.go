package healthcheck

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type Repository struct {
	store map[string]*model.Runtime
}

func NewRepository() *Repository {
	return &Repository{store: make(map[string]*model.Runtime)}
}
