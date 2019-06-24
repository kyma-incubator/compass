package document

import (
	"errors"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type inMemoryRepository struct {
	store map[string]*model.Document
}

func NewRepository() *inMemoryRepository {
	return &inMemoryRepository{store: make(map[string]*model.Document)}
}

func (r *inMemoryRepository) GetByID(id string) (*model.Document, error) {
	document := r.store[id]

	if document == nil {
		return nil, errors.New("document not found")
	}

	return document, nil
}

func (r *inMemoryRepository) ListAllByApplicationID(applicationID string) ([]*model.Document, error) {
	var items []*model.Document
	for _, r := range r.store {
		if r.ApplicationID == applicationID {
			items = append(items, r)
		}
	}

	return items, nil
}

// TODO: Add paging
func (r *inMemoryRepository) ListByApplicationID(applicationID string, pageSize *int, cursor *string) (*model.DocumentPage, error) {
	var items []*model.Document
	for _, r := range r.store {
		if r.ApplicationID == applicationID {
			items = append(items, r)
		}
	}

	return &model.DocumentPage{
		Data:       items,
		TotalCount: len(items),
		PageInfo: &pagination.Page{
			StartCursor: "",
			EndCursor:   "",
			HasNextPage: false,
		},
	}, nil
}

func (r *inMemoryRepository) Create(item *model.Document) error {
	if item == nil {
		return errors.New("item can not be empty")
	}

	if r.store[item.ApplicationID] == nil {
		return errors.New(fmt.Sprintf("application with ID %s not found", item.ApplicationID))
	}

	r.store[item.ID] = item

	return nil
}

func (r *inMemoryRepository) CreateMany(items []*model.Document) error {
	var err error
	for _, item := range items {
		if e := r.Create(item); e != nil {
			err = e
		}
	}

	return err
}

func (r *inMemoryRepository) Delete(item *model.Document) error {
	if item == nil {
		return nil
	}

	delete(r.store, item.ID)

	return nil
}

func (r *inMemoryRepository) DeleteAllByApplicationID(applicationID string) error {
	var err error
	for _, item := range r.store {
		if item.ApplicationID != applicationID {
			continue
		}

		if e := r.Delete(item); e != nil {
			err = e
		}
	}

	return err
}
