package eventdef

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"
)

const eventAPIDefTable string = `"public"."event_api_definitions"`

var (
	tenantColumn  string = `tenant_id`
	apiDefColumns        = []string{"id", "tenant_id", "app_id", "package_id", "name", "description", "group_name", "spec_data",
		"spec_format", "spec_type", "version_value", "version_deprecated", "version_deprecated_since",
		"version_for_removal"}
	idColumns        = []string{"id"}
	updatableColumns = []string{"name", "description", "group_name", "spec_data", "spec_format", "spec_type",
		"version_value", "version_deprecated", "version_deprecated_since", "version_for_removal"}
)

//go:generate mockery -name=EventAPIDefinitionConverter -output=automock -outpkg=automock -case=underscore
type EventAPIDefinitionConverter interface {
	FromEntity(entity Entity) (model.EventDefinition, error)
	ToEntity(apiModel model.EventDefinition) (Entity, error)
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

func (r *pgRepository) GetByID(ctx context.Context, tenantID string, id string) (*model.EventDefinition, error) {
	var eventAPIDefEntity Entity
	err := r.singleGetter.Get(ctx, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &eventAPIDefEntity)
	if err != nil {
		return nil, errors.Wrap(err, "while getting EventDefinition")
	}

	eventAPIDefModel, err := r.conv.FromEntity(eventAPIDefEntity)
	if err != nil {
		return nil, errors.Wrap(err, "while creating EventDefinition entity to model")
	}

	return &eventAPIDefModel, nil
}

func (r *pgRepository) GetForApplication(ctx context.Context, tenant string, id string, applicationID string) (*model.EventDefinition, error) {
	var ent Entity

	conditions := repo.Conditions{
		repo.NewEqualCondition("id", id),
		repo.NewEqualCondition("app_id", applicationID),
	}
	if err := r.singleGetter.Get(ctx, tenant, conditions, repo.NoOrderBy, &ent); err != nil {
		return nil, err
	}

	eventAPIModel, err := r.conv.FromEntity(ent)
	if err != nil {
		return nil, errors.Wrap(err, "while creating event definition model from entity")
	}

	return &eventAPIModel, nil
}

func (r *pgRepository) GetForPackage(ctx context.Context, tenant string, id string, packageID string) (*model.EventDefinition, error) {
	var ent Entity

	conditions := repo.Conditions{
		repo.NewEqualCondition("id", id),
		repo.NewEqualCondition("package_id", packageID),
	}
	if err := r.singleGetter.Get(ctx, tenant, conditions, repo.NoOrderBy, &ent); err != nil {
		return nil, err
	}

	eventAPIModel, err := r.conv.FromEntity(ent)
	if err != nil {
		return nil, errors.Wrap(err, "while creating event definition model from entity")
	}

	return &eventAPIModel, nil
}

func (r *pgRepository) ListByApplicationID(ctx context.Context, tenantID string, applicationID string, pageSize int, cursor string) (*model.EventDefinitionPage, error) {
	appCond := fmt.Sprintf("%s = '%s'", "app_id", applicationID)
	return r.list(ctx, tenantID, pageSize, cursor, appCond)
}

func (r *pgRepository) ListByPackageID(ctx context.Context, tenantID string, packageID string, pageSize int, cursor string) (*model.EventDefinitionPage, error) {
	pkgCond := fmt.Sprintf("%s = '%s'", "package_id", packageID)
	return r.list(ctx, tenantID, pageSize, cursor, pkgCond)
}

func (r *pgRepository) list(ctx context.Context, tenant string, pageSize int, cursor string, conditions string) (*model.EventDefinitionPage, error) {
	var eventCollection EventAPIDefCollection
	page, totalCount, err := r.pageableQuerier.List(ctx, tenant, pageSize, cursor, "id", &eventCollection, conditions)
	if err != nil {
		return nil, err
	}

	var items []*model.EventDefinition

	for _, eventEnt := range eventCollection {
		m, err := r.conv.FromEntity(eventEnt)
		if err != nil {
			return nil, errors.Wrap(err, "while creating APIDefinition model from entity")
		}
		items = append(items, &m)
	}

	return &model.EventDefinitionPage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page,
	}, nil
}

func (r *pgRepository) Create(ctx context.Context, item *model.EventDefinition) error {
	if item == nil {
		return errors.New("item cannot be nil")
	}

	entity, err := r.conv.ToEntity(*item)
	if err != nil {
		return errors.Wrap(err, "while creating EventDefinition model to entity")
	}

	err = r.creator.Create(ctx, entity)
	if err != nil {
		return errors.Wrap(err, "while saving entity to db")
	}

	return nil
}

func (r *pgRepository) CreateMany(ctx context.Context, items []*model.EventDefinition) error {
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

func (r *pgRepository) Update(ctx context.Context, item *model.EventDefinition) error {
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
