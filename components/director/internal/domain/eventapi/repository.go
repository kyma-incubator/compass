package eventapi

import (
	"context"
	"fmt"

	"github.com/lib/pq"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"
)

//type inMemoryRepository struct {
//	store map[string]*model.EventAPIDefinition
//}
//
//func NewRepository() *inMemoryRepository {
//	return &inMemoryRepository{store: make(map[string]*model.EventAPIDefinition)}
//}
//
//func (r *inMemoryRepository) GetByID(id string) (*model.EventAPIDefinition, error) {
//	eventAPIDefinition := r.store[id]
//
//	if eventAPIDefinition == nil {
//		return nil, errors.Errorf("EventAPIDefinition with %s ID does not exist", id)
//	}
//
//	return eventAPIDefinition, nil
//}
//
//func (r *inMemoryRepository) Exists(ctx context.Context, tenant, id string) (bool, error) {
//	item := r.store[id]
//
//	if item == nil { // TODO: Temporary because tenant is not populated
//		//if item == nil || item.Tenant != tenant {
//		return false, nil
//	}
//
//	return true, nil
//}
//
//// TODO: Make filtering and paging
//func (r *inMemoryRepository) List(filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.EventAPIDefinitionPage, error) {
//	var items []*model.EventAPIDefinition
//	for _, r := range r.store {
//		items = append(items, r)
//	}
//
//	return &model.EventAPIDefinitionPage{
//		Data:       items,
//		TotalCount: len(items),
//		PageInfo: &pagination.Page{
//			StartCursor: "",
//			EndCursor:   "",
//			HasNextPage: false,
//		},
//	}, nil
//}
//
//func (r *inMemoryRepository) ListByApplicationID(applicationID string, pageSize *int, cursor *string) (*model.EventAPIDefinitionPage, error) {
//	var items []*model.EventAPIDefinition
//	for _, a := range r.store {
//		if a.ApplicationID == applicationID {
//			items = append(items, a)
//		}
//	}
//
//	return &model.EventAPIDefinitionPage{
//		Data:       items,
//		TotalCount: len(items),
//		PageInfo: &pagination.Page{
//			StartCursor: "",
//			EndCursor:   "",
//			HasNextPage: false,
//		},
//	}, nil
//}
//
//func (r *inMemoryRepository) Create(item *model.EventAPIDefinition) error {
//	if item == nil {
//		return errors.New("item can not be nil")
//	}
//
//	r.store[item.ID] = item
//
//	return nil
//}
//
//func (r *inMemoryRepository) CreateMany(items []*model.EventAPIDefinition) error {
//	for _, item := range items {
//		err := r.Create(item)
//		if err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func (r *inMemoryRepository) Update(item *model.EventAPIDefinition) error {
//	if item == nil {
//		return errors.New("item can not be nil")
//	}
//
//	r.store[item.ID] = item
//
//	return nil
//}
//
//func (r *inMemoryRepository) Delete(item *model.EventAPIDefinition) error {
//	if item == nil {
//		return errors.New("item can not be nil")
//	}
//
//	delete(r.store, item.ID)
//
//	return nil
//}
//
//func (r *inMemoryRepository) DeleteAllByApplicationID(id string) error {
//	for _, item := range r.store {
//		if item.ApplicationID == id {
//			err := r.Delete(item)
//			if err != nil {
//				return err
//			}
//		}
//	}
//
//	return nil
//}

const eventAPIDefTable string = `"public"."event_api_definitions"`
const tenantColumn string = `tenant_id`

var apiDefColumns = []string{"id", "tenant_id", "app_id", "name", "description", "group_name", "spec_data",
	"spec_format", "spec_type", "version_value", "version_deprecated", "version_deprecated_since",
	"version_for_removal"}

var idColumns = []string{"id"}

var updatableColumns = []string{"name", "description", "group_name", "spec_data", "spec_format", "spec_type",
	"version_value", "version_deprecated", "version_deprecated_since", "version_for_removal"}

//go:generate mockery -name=EventAPIDefinitionConverter -output=automock -outpkg=automock -case=underscore
type EventAPIDefinitionConverter interface {
	FromEntity(entity Entity) (model.EventAPIDefinition, error)
	ToEntity(apiModel model.EventAPIDefinition) (Entity, error)
}

type pgRepository struct {
	*repo.SingleGetter
	*repo.PageableQuerier
	*repo.Creator
	*repo.Updater
	*repo.Deleter
	*repo.ExistQuerier
	conv EventAPIDefinitionConverter
}

func NewPostgresRepository(conv EventAPIDefinitionConverter) *pgRepository {
	return &pgRepository{
		SingleGetter:    repo.NewSingleGetter(eventAPIDefTable, tenantColumn, apiDefColumns),
		PageableQuerier: repo.NewPageableQuerier(eventAPIDefTable, tenantColumn, apiDefColumns),
		Creator:         repo.NewCreator(eventAPIDefTable, apiDefColumns),
		Updater:         repo.NewUpdater(eventAPIDefTable, updatableColumns, tenantColumn, idColumns),
		Deleter:         repo.NewDeleter(eventAPIDefTable, tenantColumn),
		ExistQuerier:    repo.NewExistQuerier(eventAPIDefTable, tenantColumn),
		conv:            conv,
	}
}

type EventAPIDefCollection []Entity

func (r EventAPIDefCollection) Len() int {
	return len(r)
}

func (r *pgRepository) ListByApplicationID(ctx context.Context, tenantID string, applicationID string, pageSize int, cursor string) (*model.EventAPIDefinitionPage, error) {
	appCond := fmt.Sprintf("app_id = %s ", pq.QuoteLiteral(applicationID))
	var eventAPIDefCollection EventAPIDefCollection
	page, totalCount, err := r.PageableQuerier.List(ctx, tenantID, pageSize, cursor, "id", &eventAPIDefCollection, appCond)
	if err != nil {
		return nil, err
	}

	var items []*model.EventAPIDefinition

	for _, apiDefEnt := range eventAPIDefCollection {
		m, err := r.conv.FromEntity(apiDefEnt)
		if err != nil {
			return nil, errors.Wrap(err, "while creating EventAPIDefinition model from entity")
		}
		items = append(items, &m)
	}

	return &model.EventAPIDefinitionPage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page,
	}, nil
}

func (r *pgRepository) GetByID(ctx context.Context, tenantID string, id string) (*model.EventAPIDefinition, error) {
	var eventAPIDefEntity Entity
	err := r.SingleGetter.Get(ctx, tenantID, repo.Conditions{{Field: "id", Val: id}}, &eventAPIDefEntity)
	if err != nil {
		return nil, errors.Wrap(err, "while getting EventAPIDefinition")
	}

	eventAPIDefModel, err := r.conv.FromEntity(eventAPIDefEntity)
	if err != nil {
		return nil, errors.Wrap(err, "while creating EventAPIDefinition entity to model")
	}

	return &eventAPIDefModel, nil
}

func (r *pgRepository) Create(ctx context.Context, tenantID string, item *model.EventAPIDefinition) error {
	if item == nil {
		return errors.New("item cannot be nil")
	}

	entity, err := r.conv.ToEntity(*item)
	if err != nil {
		return errors.Wrap(err, "while creating EventAPIDefinition model to entity")
	}

	err = r.Creator.Create(ctx, entity)
	if err != nil {
		return errors.Wrap(err, "while saving entity to db")
	}

	return nil
}

func (r *pgRepository) CreateMany(ctx context.Context, tenantID string, items []*model.EventAPIDefinition) error {
	for index, item := range items {
		entity, err := r.conv.ToEntity(*item)
		if err != nil {
			return errors.Wrapf(err, "while creating %d item", index)
		}
		err = r.Creator.Create(ctx, entity)
		if err != nil {
			return errors.Wrapf(err, "while persisting %d item", index)
		}
	}

	return nil
}

func (r *pgRepository) Update(ctx context.Context, tenantID string, item *model.EventAPIDefinition) error {
	if item == nil {
		return errors.New("item cannot be nil")
	}

	entity, err := r.conv.ToEntity(*item)
	if err != nil {
		return errors.Wrap(err, "while converting model to entity")
	}

	return r.Updater.UpdateSingle(ctx, entity)
}

func (r *pgRepository) Exists(ctx context.Context, tenantID, id string) (bool, error) {
	return r.ExistQuerier.Exists(ctx, tenantID, repo.Conditions{repo.Condition{Field: "id", Val: id}})
}

func (r *pgRepository) Delete(ctx context.Context, tenantID string, id string) error {
	return r.DeleteOne(ctx, tenantID, repo.Conditions{repo.Condition{Field: "id", Val: id}})
}

func (r *pgRepository) DeleteAllByApplicationID(ctx context.Context, tenantID string, appID string) error {
	return r.DeleteMany(ctx, tenantID, repo.Conditions{repo.Condition{Field: "app_id", Val: appID}})
}
