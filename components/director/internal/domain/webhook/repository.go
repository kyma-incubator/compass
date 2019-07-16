package webhook

import (
	"errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type inMemoryRepository struct {
	store map[string]*model.Webhook
}

func NewRepository() *inMemoryRepository {
	return &inMemoryRepository{store: make(map[string]*model.Webhook)}
}

func (r *inMemoryRepository) GetByID(id string) (*model.Webhook, error) {
	webhook := r.store[id]

	if webhook == nil {
		return nil, errors.New("webhook not found")
	}

	return webhook, nil
}

func (r *inMemoryRepository) ListByApplicationID(applicationID string) ([]*model.Webhook, error) {
	var items []*model.Webhook
	for _, r := range r.store {
		if r.ApplicationID == applicationID {
			items = append(items, r)
		}
	}

	return items, nil
}

func (r *inMemoryRepository) Create(item *model.Webhook) error {
	if item == nil {
		return errors.New("item can not be empty")
	}

	r.store[item.ID] = item

	return nil
}

func (r *inMemoryRepository) CreateMany(items []*model.Webhook) error {
	var err error
	for _, item := range items {
		if e := r.Create(item); e != nil {
			err = e
		}
	}

	return err
}

func (r *inMemoryRepository) Update(item *model.Webhook) error {
	if item == nil {
		return errors.New("item can not be empty")
	}

	if r.store[item.ID] == nil {
		return errors.New("webhook not found")
	}

	r.store[item.ID] = item

	return nil
}

func (r *inMemoryRepository) Delete(item *model.Webhook) error {
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
