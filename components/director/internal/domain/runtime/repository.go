package runtime

import (
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type inMemoryRepository struct {
	store map[string]*model.Runtime
}

func NewRepository() *inMemoryRepository {
	return &inMemoryRepository{store: make(map[string]*model.Runtime)}
}

func (r *inMemoryRepository) GetByID(id string) (*model.Runtime, error) {
	runtime := r.store[id]
	return runtime, nil
}

// TODO: Make filtering and paging
func (r *inMemoryRepository) List(filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.RuntimePage, error) {
	var items []*model.Runtime
	for _, r := range r.store {
		items = append(items, r)
	}

	return &model.RuntimePage{
		Data:       items,
		TotalCount: len(items),
		PageInfo: &pagination.Page{
			StartCursor: "",
			EndCursor:   "",
			HasNextPage: false,
		},
	}, nil
}

func (r *inMemoryRepository) Create(item *model.Runtime) error {
	r.store[item.ID] = item

	return nil
}

func (r *inMemoryRepository) Update(item *model.Runtime) error {
	r.store[item.ID] = item

	return nil
}

func (r *inMemoryRepository) Delete(item *model.Runtime) error {
	delete(r.store, item.ID)

	return nil
}
