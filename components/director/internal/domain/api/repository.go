package api

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"
)

const apiDefTable string = `"public"."api_definitions"`

var (
	tenantColumn  = "tenant_id"
	apiDefColumns = []string{"id", "tenant_id", "app_id", "name", "description", "group_name", "target_url", "spec_data",
		"spec_format", "spec_type", "default_auth",
		"version_value", "version_deprecated", "version_deprecated_since", "version_for_removal"}
	idColumns        = []string{"id"}
	updatableColumns = []string{"name", "description", "group_name", "target_url", "spec_data", "spec_format", "spec_type",
		"default_auth", "version_value", "version_deprecated", "version_deprecated_since", "version_for_removal"}
)

//go:generate mockery -name=APIDefinitionConverter -output=automock -outpkg=automock -case=underscore
type APIDefinitionConverter interface {
	FromEntity(entity Entity) (model.APIDefinition, error)
	ToEntity(apiModel model.APIDefinition) (Entity, error)
}

type pgRepository struct {
	creator         repo.Creator
	singleGetter    repo.SingleGetter
	pageableQuerier repo.PageableQuerier
	updater         repo.Updater
	deleter         repo.Deleter
	existQuerier    repo.ExistQuerier
	conv            APIDefinitionConverter
}

func NewRepository(conv APIDefinitionConverter) *pgRepository {
	return &pgRepository{
		singleGetter:    repo.NewSingleGetter(apiDefTable, tenantColumn, apiDefColumns),
		pageableQuerier: repo.NewPageableQuerier(apiDefTable, tenantColumn, apiDefColumns),
		creator:         repo.NewCreator(apiDefTable, apiDefColumns),
		updater:         repo.NewUpdater(apiDefTable, updatableColumns, tenantColumn, idColumns),
		deleter:         repo.NewDeleter(apiDefTable, tenantColumn),
		existQuerier:    repo.NewExistQuerier(apiDefTable, tenantColumn),
		conv:            conv,
	}
}

type APIDefCollection []Entity

func (r APIDefCollection) Len() int {
	return len(r)
}

func (r *pgRepository) ListByApplicationID(ctx context.Context, tenantID string, applicationID string, pageSize int, cursor string) (*model.APIDefinitionPage, error) {
	appCond := fmt.Sprintf("%s = '%s'", "app_id", applicationID)
	var apiDefCollection APIDefCollection
	page, totalCount, err := r.pageableQuerier.List(ctx, tenantID, pageSize, cursor, "id", &apiDefCollection, appCond)
	if err != nil {
		return nil, err
	}

	var items []*model.APIDefinition

	for _, apiDefEnt := range apiDefCollection {
		m, err := r.conv.FromEntity(apiDefEnt)
		if err != nil {
			return nil, errors.Wrap(err, "while creating APIDefinition model from entity")
		}
		items = append(items, &m)
	}

	return &model.APIDefinitionPage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page,
	}, nil
}

func (r *pgRepository) GetByID(ctx context.Context, tenantID string, id string) (*model.APIDefinition, error) {
	var apiDefEntity Entity
	err := r.singleGetter.Get(ctx, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &apiDefEntity)
	if err != nil {
		return nil, errors.Wrap(err, "while getting APIDefinition")
	}

	apiDefModel, err := r.conv.FromEntity(apiDefEntity)
	if err != nil {
		return nil, errors.Wrap(err, "while creating APIDefinition entity to model")
	}

	return &apiDefModel, nil
}

func (r *pgRepository) GetForApplication(ctx context.Context, tenant string, id string, applicationID string) (*model.APIDefinition, error) {
	var ent Entity

	conditions := repo.Conditions{
		repo.NewEqualCondition("id", id),
		repo.NewEqualCondition("app_id", applicationID),
	}
	if err := r.singleGetter.Get(ctx, tenant, conditions, repo.NoOrderBy, &ent); err != nil {
		return nil, err
	}

	apiDefModel, err := r.conv.FromEntity(ent)
	if err != nil {
		return nil, errors.Wrap(err, "while creating api definition model from entity")
	}

	return &apiDefModel, nil
}

func (r *pgRepository) Create(ctx context.Context, item *model.APIDefinition) error {
	if item == nil {
		return errors.New("item cannot be nil")
	}

	entity, err := r.conv.ToEntity(*item)
	if err != nil {
		return errors.Wrap(err, "while creating APIDefinition model to entity")
	}

	err = r.creator.Create(ctx, entity)
	if err != nil {
		return errors.Wrap(err, "while saving entity to db")
	}

	return nil
}

func (r *pgRepository) CreateMany(ctx context.Context, items []*model.APIDefinition) error {
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

func (r *pgRepository) Update(ctx context.Context, item *model.APIDefinition) error {
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
