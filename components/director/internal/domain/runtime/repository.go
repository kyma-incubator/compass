package runtime

import (
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type Repository struct {
	store map[string]*model.Runtime
}

type RuntimePage struct {
	Data       []*model.Runtime
	PageInfo   *pagination.Page
	TotalCount int
}

func NewRuntimeRepository() *Repository {
	return &Repository{store: make(map[string]*model.Runtime)}
}

func (r *Repository) GetByID(id string) (*model.Runtime, error) {
	runtime := r.store[id]
	return runtime, nil
}

// TODO: Make filtering and paging
func (r *Repository) List(filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*RuntimePage, error) {
	var items []*model.Runtime
	for _, r := range r.store {
		items = append(items, r)
	}

	return &RuntimePage{
		Data:       items,
		TotalCount: len(items),
		PageInfo: &pagination.Page{
			StartCursor: "",
			EndCursor:   "",
			HasNextPage: false,
		},
	}, nil
}

func (r *Repository) Create(item *model.Runtime) error {
	r.store[item.ID] = item

	return nil
}

func (r *Repository) Update(item *model.Runtime) error {
	r.store[item.ID] = item

	return nil
}

func (r *Repository) Delete(item *model.Runtime) error {
	delete(r.store, item.ID)

	return nil
}
