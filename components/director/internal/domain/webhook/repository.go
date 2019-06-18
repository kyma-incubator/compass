package webhook

import (
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type inMemoryRepository struct {
	store map[string]*model.ApplicationWebhook
}

func NewRepository() *inMemoryRepository {
	return &inMemoryRepository{store: make(map[string]*model.ApplicationWebhook)}
}

func (r *inMemoryRepository) GetByID(id string) (*model.ApplicationWebhook, error) {
	panic("not implemented")
}

// TODO: Make filtering and paging
func (r *inMemoryRepository) List(filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) ([]*model.ApplicationWebhook, error) {
	panic("not implemented")
}

func (r *inMemoryRepository) Create(item *model.ApplicationWebhookInput) error {
	panic("not implemented")
}

func (r *inMemoryRepository) Update(item *model.ApplicationWebhook) error {
	panic("not implemented")
}

func (r *inMemoryRepository) Delete(item *model.ApplicationWebhook) error {
	panic("not implemented")
}

