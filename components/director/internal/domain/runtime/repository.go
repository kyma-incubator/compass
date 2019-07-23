package runtime

import (
	"errors"

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

func (r *inMemoryRepository) GetByID(tenant, id string) (*model.Runtime, error) {
	rtm := r.store[id]
	if rtm == nil || rtm.Tenant != tenant {
		return nil, errors.New("runtime not found")
	}

	return rtm, nil
}

// TODO: Make filtering and paging
func (r *inMemoryRepository) List(tenant string, filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.RuntimePage, error) {
	var items []*model.Runtime
	for _, item := range r.store {
		if item.Tenant == tenant {
			items = append(items, item)
		}
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
	if item == nil {
		return errors.New("item can not be empty")
	}

	found := r.findRuntimeNameWithinTenant(item.Tenant, item.Name)
	if found {
		return errors.New("Runtime name is not unique within tenant")
	}

	r.store[item.ID] = item

	return nil
}

func (r *inMemoryRepository) Update(item *model.Runtime) error {
	if item == nil {
		return errors.New("item can not be empty")
	}

	rtm := r.store[item.ID]
	if rtm == nil {
		return errors.New("Runtime not found")
	}

	if rtm.Name != item.Name {
		found := r.findRuntimeNameWithinTenant(item.Tenant, item.Name)
		if found {
			return errors.New("Runtime name is not unique within tenant")
		}
	}

	rtm.Name = item.Name
	rtm.Description = item.Description
	rtm.Labels = item.Labels

	return nil
}

func (r *inMemoryRepository) Delete(item *model.Runtime) error {
	if item == nil {
		return nil
	}

	delete(r.store, item.ID)

	return nil
}

func (r *inMemoryRepository) findRuntimeNameWithinTenant(tenant, name string) bool {
	for _, runtime := range r.store {
		if runtime.Name == name && runtime.Tenant == tenant {
			return true
		}
	}
	return false
}
