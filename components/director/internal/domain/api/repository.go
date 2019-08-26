package api

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/pkg/errors"
)

type inMemoryRepository struct {
	store map[string]*model.APIDefinition
}

func NewAPIRepository() *inMemoryRepository {
	return &inMemoryRepository{store: make(map[string]*model.APIDefinition)}
}

func (r *inMemoryRepository) GetByID(id string) (*model.APIDefinition, error) {
	if api, ok := r.store[id]; ok {
		return api, nil
	}

	api := r.store[id]

	if api == nil {
		return nil, errors.Errorf("APIDefinition with %s ID does not exist", id)
	}

	return api, nil
}

func (r *inMemoryRepository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	item := r.store[id]

	if item == nil { // TODO: Temporary because tenant is not populated
		//if item == nil || item.TenantID != tenant {
		return false, nil
	}

	return true, nil
}

// TODO: Make filtering and paging
func (r *inMemoryRepository) List(filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.APIDefinitionPage, error) {
	var items []*model.APIDefinition
	for _, r := range r.store {
		items = append(items, r)
	}

	return &model.APIDefinitionPage{
		Data:       items,
		TotalCount: len(items),
		PageInfo: &pagination.Page{
			StartCursor: "",
			EndCursor:   "",
			HasNextPage: false,
		},
	}, nil
}

func (r *inMemoryRepository) ListByApplicationID(applicationID string, pageSize *int, cursor *string) (*model.APIDefinitionPage, error) {
	var items []*model.APIDefinition
	for _, a := range r.store {
		if a.ApplicationID == applicationID {
			items = append(items, a)
		}
	}

	return &model.APIDefinitionPage{
		Data:       items,
		TotalCount: len(items),
		PageInfo: &pagination.Page{
			StartCursor: "",
			EndCursor:   "",
			HasNextPage: false,
		},
	}, nil
}

func (r *inMemoryRepository) Create(item *model.APIDefinition) error {
	if item == nil {
		return errors.New("item can not be nil")
	}

	r.store[item.ID] = item

	return nil
}

func (r *inMemoryRepository) CreateMany(items []*model.APIDefinition) error {
	for _, item := range items {
		err := r.Create(item)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *inMemoryRepository) Update(item *model.APIDefinition) error {
	if item == nil {
		return errors.New("item can not be nil")
	}

	r.store[item.ID] = item

	return nil
}

func (r *inMemoryRepository) Delete(item *model.APIDefinition) error {
	if item == nil {
		return errors.New("item can not be nil")
	}

	delete(r.store, item.ID)

	return nil
}

func (r *inMemoryRepository) DeleteAllByApplicationID(id string) error {
	for _, item := range r.store {
		if item.ApplicationID == id {
			err := r.Delete(item)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
