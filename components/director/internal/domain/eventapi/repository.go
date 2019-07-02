package eventapi

import (
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/pkg/errors"
)

type inMemoryRepository struct {
	store map[string]*model.EventAPIDefinition
}

func NewRepository() *inMemoryRepository {
	return &inMemoryRepository{store: make(map[string]*model.EventAPIDefinition)}
}

func (r *inMemoryRepository) GetByID(id string) (*model.EventAPIDefinition, error) {
	eventAPIDefinition := r.store[id]

	if eventAPIDefinition == nil {
		return nil, errors.Errorf("EventAPIDefinition with %s ID does not exist", id)
	}

	return eventAPIDefinition, nil
}

// TODO: Make filtering and paging
func (r *inMemoryRepository) List(filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.EventAPIDefinitionPage, error) {
	var items []*model.EventAPIDefinition
	for _, r := range r.store {
		items = append(items, r)
	}

	return &model.EventAPIDefinitionPage{
		Data:       items,
		TotalCount: len(items),
		PageInfo: &pagination.Page{
			StartCursor: "",
			EndCursor:   "",
			HasNextPage: false,
		},
	}, nil
}

func (r *inMemoryRepository) ListByApplicationID(applicationID string, pageSize *int, cursor *string) (*model.EventAPIDefinitionPage, error) {
	var items []*model.EventAPIDefinition
	for _, a := range r.store {
		if a.ApplicationID == applicationID {
			items = append(items, a)
		}
	}

	return &model.EventAPIDefinitionPage{
		Data:       items,
		TotalCount: len(items),
		PageInfo: &pagination.Page{
			StartCursor: "",
			EndCursor:   "",
			HasNextPage: false,
		},
	}, nil
}

func (r *inMemoryRepository) Create(item *model.EventAPIDefinition) error {
	if item == nil {
		return errors.New("item can not be nil")
	}

	r.store[item.ID] = item

	return nil
}

func (r *inMemoryRepository) CreateMany(items []*model.EventAPIDefinition) error {
	for _, item := range items {
		err := r.Create(item)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *inMemoryRepository) Update(item *model.EventAPIDefinition) error {
	if item == nil {
		return errors.New("item can not be nil")
	}

	r.store[item.ID] = item

	return nil
}

func (r *inMemoryRepository) Delete(item *model.EventAPIDefinition) error {
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
