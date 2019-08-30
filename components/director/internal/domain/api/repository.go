package api

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"
)

const apiDefTable string = `"public"."api_definitions"`
const tenantColumn string = `tenant_id`

var apiDefColumns = []string{"id", "tenant_id", "app_id", "name", "description", "group_name", "target_url", "spec_data",
	"spec_format", "spec_type", "default_auth",
	"version_value", "version_deprecated", "version_deprecated_since", "version_for_removal"}

var idColumns = []string{"id"}

var updatableColumns = []string{"name", "description", "group_name", "target_url", "spec_data",
	"spec_format", "spec_type", "default_auth", "version_value", "version_deprecated",
	"version_deprecated_since", "version_for_removal"}

//go:generate mockery -name=APIDefinitionConverter -output=automock -outpkg=automock -case=underscore
type APIDefinitionConverter interface {
	FromEntity(entity Entity) (model.APIDefinition, error)
	ToEntity(apiModel model.APIDefinition) (Entity, error)
}

type pgRepository struct {
	*repo.SingleGetter
	*repo.PageableQuerier
	*repo.Creator
	*repo.Updater
	*repo.Deleter
	*repo.ExistQuerier
	conv APIDefinitionConverter
}

func NewRepository(conv APIDefinitionConverter) *pgRepository {
	return &pgRepository{
		SingleGetter:    repo.NewSingleGetter(apiDefTable, tenantColumn, apiDefColumns),
		PageableQuerier: repo.NewPageableQuerier(apiDefTable, tenantColumn, apiDefColumns),
		Creator:         repo.NewCreator(apiDefTable, apiDefColumns),
		Updater:         repo.NewUpdater(apiDefTable, updatableColumns, tenantColumn, idColumns),
		Deleter:         repo.NewDeleter(apiDefTable, tenantColumn),
		ExistQuerier:    repo.NewExistQuerier(apiDefTable, tenantColumn),
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
	page, totalCount, err := r.PageableQuerier.List(ctx, tenantID, pageSize, cursor, "id", &apiDefCollection, appCond)
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
	err := r.SingleGetter.Get(ctx, tenantID, repo.Conditions{{Field: "id", Val: id}}, &apiDefEntity)
	if err != nil {
		return nil, errors.Wrap(err, "while getting APIDefinition")
	}

	apiDefModel, err := r.conv.FromEntity(apiDefEntity)
	if err != nil {
		return nil, errors.Wrap(err, "while creating APIDefinition entity to model")
	}

	return &apiDefModel, nil
}

func (r *pgRepository) Create(ctx context.Context, tenantID string, item *model.APIDefinition) error {
	if item == nil {
		return errors.New("item cannot be nil")
	}

	entity, err := r.conv.ToEntity(*item)
	if err != nil {
		return errors.Wrap(err, "while creating APIDefinition model to entity")
	}

	err = r.Creator.Create(ctx, entity)
	if err != nil {
		return errors.Wrap(err, "while saving entity to db")
	}

	return nil
}

func (r *pgRepository) CreateMany(ctx context.Context, tenantID string, items []*model.APIDefinition) error {
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

func (r *pgRepository) Update(ctx context.Context, tenantID string, item *model.APIDefinition) error {
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
