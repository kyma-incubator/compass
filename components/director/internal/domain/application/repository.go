package application

import (
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/pkg/errors"
)

type inMemoryRepository struct {
	store map[string]*model.Application
}

func NewRepository() *inMemoryRepository {
	return &inMemoryRepository{store: make(map[string]*model.Application)}
}

func (r *inMemoryRepository) GetByID(id string) (*model.Application, error) {
	application := r.store[id]

	if application == nil {
		return nil, errors.New("application not found")
	}

	return application, nil
}

// TODO: Make filtering and paging
func (r *inMemoryRepository) List(filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.ApplicationPage, error) {
	var items []*model.Application
	for _, r := range r.store {
		items = append(items, r)
	}

	return &model.ApplicationPage{
		Data:       items,
		TotalCount: len(items),
		PageInfo: &pagination.Page{
			StartCursor: "",
			EndCursor:   "",
			HasNextPage: false,
		},
	}, nil
}

func (r *inMemoryRepository) Create(item *model.Application) error {
	if item == nil {
		return errors.New("item can not be empty")
	}

	r.store[item.ID] = item

	return nil
}

func (r *inMemoryRepository) Update(item *model.Application) error {
	if item == nil {
		return errors.New("item can not be empty")
	}

	r.store[item.ID] = item

	return nil
}

func (r *inMemoryRepository) Delete(item *model.Application) error {
	if item == nil {
		return nil
	}

	delete(r.store, item.ID)

	return nil
}
