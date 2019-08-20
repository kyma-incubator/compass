package document

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

const documentTable string = `"public"."documents"`

var documentColumns = []string{"id", "tenant_id", "app_id", "title", "display_name", "description", "format", "kind", "data", "fetch_request"}

type pgRepository struct {
	*repo.ExistQuerier
	*repo.SingleGetter
	*repo.Deleter
	*repo.PageableQuerier
	*repo.Creator
	*repo.Updater
	*converter
}

func NewPostgresRepository() *pgRepository {
	return &pgRepository{
		ExistQuerier:    repo.NewExistQuerier(documentTable, "tenant_id"),
		SingleGetter:    repo.NewSingleGetter(documentTable, "tenant_id", documentColumns),
		Deleter:         repo.NewDeleter(documentTable, "tenant_id"),
		PageableQuerier: repo.NewPageableQuerier(documentTable, "tenant_id", documentColumns),
		Creator:         repo.NewCreator(documentTable, documentColumns),
		Updater:         repo.NewUpdater(documentTable, []string{"name", "description", "status_condition", "status_timestamp"}, "tenant_id", []string{"id"}),
		converter:       NewConverter(fetchrequest.NewConverter(auth.NewConverter())),
	}
}

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

func (r *pgRepository) Create(ctx context.Context, item *model.Document) error {
	if item == nil {
		return errors.New("item can not be empty")
	}

	documentEnt, err := r.converter.ToEntity(*item)
	if err != nil {
		return errors.Wrap(err, "while creating document entity from model")
	}

	return r.Creator.Create(ctx, documentEnt)
}
