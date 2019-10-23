package eventapi

import (
	"context"
	"fmt"

	"github.com/lib/pq"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"
)

const eventAPIDefTable string = `"public"."event_api_definitions"`

var (
	tenantColumn  string = `tenant_id`
	apiDefColumns        = []string{"id", "tenant_id", "app_id", "name", "description", "group_name", "spec_data",
		"spec_format", "spec_type", "version_value", "version_deprecated", "version_deprecated_since",
		"version_for_removal"}
	idColumns        = []string{"id"}
	updatableColumns = []string{"name", "description", "group_name", "spec_data", "spec_format", "spec_type",
		"version_value", "version_deprecated", "version_deprecated_since", "version_for_removal"}
)

//go:generate mockery -name=EventAPIDefinitionConverter -output=automock -outpkg=automock -case=underscore
type EventAPIDefinitionConverter interface {
	FromEntity(entity Entity) (model.EventAPIDefinition, error)
	ToEntity(apiModel model.EventAPIDefinition) (Entity, error)
}

type pgRepository struct {
	singleGetter    repo.SingleGetter
	pageableQuerier repo.PageableQuerier
	creator         repo.Creator
	updater         repo.Updater
	deleter         repo.Deleter
	existQuerier    repo.ExistQuerier
	conv            EventAPIDefinitionConverter
}

func NewRepository(conv EventAPIDefinitionConverter) *pgRepository {
	return &pgRepository{
		singleGetter:    repo.NewSingleGetter(eventAPIDefTable, tenantColumn, apiDefColumns),
		pageableQuerier: repo.NewPageableQuerier(eventAPIDefTable, tenantColumn, apiDefColumns),
		creator:         repo.NewCreator(eventAPIDefTable, apiDefColumns),
		updater:         repo.NewUpdater(eventAPIDefTable, updatableColumns, tenantColumn, idColumns),
		deleter:         repo.NewDeleter(eventAPIDefTable, tenantColumn),
		existQuerier:    repo.NewExistQuerier(eventAPIDefTable, tenantColumn),
		conv:            conv,
	}
}

type EventAPIDefCollection []Entity

func (r EventAPIDefCollection) Len() int {
	return len(r)
}

func (r *pgRepository) GetByID(ctx context.Context, tenantID string, id string) (*model.EventAPIDefinition, error) {
	var eventAPIDefEntity Entity
	err := r.singleGetter.Get(ctx, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)}, &eventAPIDefEntity)
	if err != nil {
		return nil, errors.Wrap(err, "while getting EventAPIDefinition")
	}

	eventAPIDefModel, err := r.conv.FromEntity(eventAPIDefEntity)
	if err != nil {
		return nil, errors.Wrap(err, "while creating EventAPIDefinition entity to model")
	}

	return &eventAPIDefModel, nil
}

func (r *pgRepository) ListByApplicationID(ctx context.Context, tenantID string, applicationID string, pageSize int, cursor string) (*model.EventAPIDefinitionPage, error) {
	appCond := fmt.Sprintf("app_id = %s ", pq.QuoteLiteral(applicationID))
	var eventAPIDefCollection EventAPIDefCollection
	page, totalCount, err := r.pageableQuerier.List(ctx, tenantID, pageSize, cursor, "id", &eventAPIDefCollection, appCond)
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

func (r *pgRepository) Create(ctx context.Context, item *model.EventAPIDefinition) error {
	if item == nil {
		return errors.New("item cannot be nil")
	}

	entity, err := r.conv.ToEntity(*item)
	if err != nil {
		return errors.Wrap(err, "while creating EventAPIDefinition model to entity")
	}

	err = r.creator.Create(ctx, entity)
	if err != nil {
		return errors.Wrap(err, "while saving entity to db")
	}

	return nil
}

func (r *pgRepository) CreateMany(ctx context.Context, items []*model.EventAPIDefinition) error {
	for index, item := range items {
		entity, err := r.conv.ToEntity(*item)
		if err != nil {
			return errors.Wrapf(err, "while creating %d item", index)
		}
		err = r.creator.Create(ctx, entity)
		if err != nil {
			return errors.Wrapf(err, "while persisting %d item", index)
		}
	}

	return nil
}

func (r *pgRepository) Update(ctx context.Context, item *model.EventAPIDefinition) error {
	if item == nil {
		return errors.New("item cannot be nil")
	}

	entity, err := r.conv.ToEntity(*item)
	if err != nil {
		return errors.Wrap(err, "while converting model to entity")
	}

	return r.updater.UpdateSingle(ctx, entity)
}

func (r *pgRepository) Exists(ctx context.Context, tenantID, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *pgRepository) Delete(ctx context.Context, tenantID string, id string) error {
	return r.deleter.DeleteOne(ctx, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *pgRepository) DeleteAllByApplicationID(ctx context.Context, tenantID string, appID string) error {
	return r.deleter.DeleteMany(ctx, tenantID, repo.Conditions{repo.NewEqualCondition("app_id", appID)})
}
