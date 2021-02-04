package spec

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"
)

const specificationsTable string = `public.specifications`

const apiDefIDColumn = "api_def_id"
const eventAPIDefIDColumn = "event_def_id"

var (
	specificationsColumns = []string{"id", "tenant_id", apiDefIDColumn, eventAPIDefIDColumn, "spec_data", "api_spec_format", "api_spec_type", "event_spec_format", "event_spec_type", "custom_type"}
	tenantColumn          = "tenant_id"
)

//go:generate mockery -name=Converter -output=automock -outpkg=automock -case=underscore
type Converter interface {
	ToEntity(in model.Spec) Entity
	FromEntity(in Entity) (model.Spec, error)
}

type repository struct {
	creator      repo.Creator
	lister       repo.Lister
	getter       repo.SingleGetter
	deleter      repo.Deleter
	updater      repo.Updater
	existQuerier repo.ExistQuerier
	conv         Converter
}

func NewRepository(conv Converter) *repository {
	return &repository{
		creator: repo.NewCreator(resource.Specification, specificationsTable, specificationsColumns),
		getter:  repo.NewSingleGetter(resource.Specification, specificationsTable, tenantColumn, specificationsColumns),
		lister: repo.NewListerWithOrderBy(resource.Specification, specificationsTable, tenantColumn, specificationsColumns, repo.OrderByParams{
			{
				Field: "created_at",
				Dir:   repo.AscOrderBy,
			},
		}),
		deleter:      repo.NewDeleter(resource.Specification, specificationsTable, tenantColumn),
		updater:      repo.NewUpdater(resource.Specification, specificationsTable, []string{"spec_data", "api_spec_format", "api_spec_type", "event_spec_format", "event_spec_type"}, tenantColumn, []string{"id"}),
		existQuerier: repo.NewExistQuerier(resource.Specification, specificationsTable, tenantColumn),
		conv:         conv,
	}
}

func (r *repository) GetByID(ctx context.Context, tenantID string, id string) (*model.Spec, error) {
	var specEntity Entity
	err := r.getter.Get(ctx, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &specEntity)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Specification with id %q", id)
	}

	specModel, err := r.conv.FromEntity(specEntity)
	if err != nil {
		return nil, err
	}

	return &specModel, nil
}

func (r *repository) Create(ctx context.Context, item *model.Spec) error {
	if item == nil {
		return apperrors.NewInternalError("item can not be empty")
	}

	entity := r.conv.ToEntity(*item)

	return r.creator.Create(ctx, entity)
}

func (r *repository) ListByReferenceObjectID(ctx context.Context, tenant string, objectType model.SpecReferenceObjectType, objectID string) ([]*model.Spec, error) {
	fieldName, err := r.referenceObjectFieldName(objectType)
	if err != nil {
		return nil, err
	}
	conditions := repo.Conditions{
		repo.NewEqualCondition(fieldName, objectID),
	}

	var specCollection specCollection
	err = r.lister.List(ctx, tenant, &specCollection, conditions...)
	if err != nil {
		return nil, err
	}

	var items []*model.Spec

	for _, specEnt := range specCollection {
		m, err := r.conv.FromEntity(specEnt)
		if err != nil {
			return nil, err
		}

		items = append(items, &m)
	}

	return items, nil
}

func (r *repository) Delete(ctx context.Context, tenant, id string) error {
	return r.deleter.DeleteOne(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *repository) DeleteByReferenceObjectID(ctx context.Context, tenant string, objectType model.SpecReferenceObjectType, objectID string) error {
	fieldName, err := r.referenceObjectFieldName(objectType)
	if err != nil {
		return err
	}

	return r.deleter.DeleteMany(ctx, tenant, repo.Conditions{repo.NewEqualCondition(fieldName, objectID)})
}

func (r *repository) Update(ctx context.Context, item *model.Spec) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(*item)

	return r.updater.UpdateSingle(ctx, entity)
}

func (r *repository) Exists(ctx context.Context, tenantID, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *repository) referenceObjectFieldName(objectType model.SpecReferenceObjectType) (string, error) {
	switch objectType {
	case model.APISpecReference:
		return apiDefIDColumn, nil
	case model.EventSpecReference:
		return eventAPIDefIDColumn, nil
	}

	return "", apperrors.NewInternalError("Invalid type of the Specification reference object")
}

type specCollection []Entity

func (r specCollection) Len() int {
	return len(r)
}
