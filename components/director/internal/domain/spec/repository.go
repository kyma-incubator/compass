package spec

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"
)

const (
	specificationsTable = `public.specifications`
	apiDefIDColumn      = "api_def_id"
	eventAPIDefIDColumn = "event_def_id"
	pageSize            = 1
	cursor              = ""
)

var (
	specificationsColumns = []string{"id", apiDefIDColumn, eventAPIDefIDColumn, "spec_data", "spec_data_hash", "api_spec_format", "api_spec_type", "event_spec_format", "event_spec_type", "custom_type"}
	orderByColumns        = repo.OrderByParams{repo.NewAscOrderBy("created_at"), repo.NewAscOrderBy("id")}
)

// Converter missing godoc
//go:generate mockery --name=Converter --output=automock --outpkg=automock --case=underscore --disable-version-string
type Converter interface {
	ToEntity(in *model.Spec) *Entity
	FromEntity(in *Entity) (*model.Spec, error)
}

type repository struct {
	creator        repo.Creator
	lister         repo.Lister
	idLister         repo.Lister
	unionLister    repo.UnionLister
	getter         repo.SingleGetter
	dataHashGetter repo.SingleGetter
	deleter        repo.Deleter
	updater        repo.Updater
	existQuerier   repo.ExistQuerier
	conv           Converter
}

// NewRepository missing godoc
func NewRepository(conv Converter) *repository {
	return &repository{
		creator:        repo.NewCreator(specificationsTable, specificationsColumns),
		getter:         repo.NewSingleGetter(specificationsTable, specificationsColumns),
		dataHashGetter: repo.NewSingleGetter(specificationsTable, []string{"spec_data_hash"}),
		lister: repo.NewListerWithOrderBy(specificationsTable, specificationsColumns, repo.OrderByParams{
			{
				Field: "created_at",
				Dir:   repo.AscOrderBy,
			},
		}),
		idLister: repo.NewListerWithOrderBy(specificationsTable, []string{"id"}, repo.OrderByParams{
			{
				Field: "created_at",
				Dir:   repo.AscOrderBy,
			},
		}),
		unionLister:  repo.NewUnionLister(specificationsTable, specificationsColumns),
		deleter:      repo.NewDeleter(specificationsTable),
		updater:      repo.NewUpdater(specificationsTable, []string{"spec_data", "spec_data_hash", "api_spec_format", "api_spec_type", "event_spec_format", "event_spec_type"}, []string{"id"}),
		existQuerier: repo.NewExistQuerier(specificationsTable),
		conv:         conv,
	}
}

// GetByID missing godoc
func (r *repository) GetByID(ctx context.Context, tenantID string, id string, objectType model.SpecReferenceObjectType) (*model.Spec, error) {
	var specEntity Entity
	err := r.getter.Get(ctx, objectType.GetResourceType(), tenantID, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &specEntity)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Specification with id %q", id)
	}

	specModel, err := r.conv.FromEntity(&specEntity)
	if err != nil {
		return nil, err
	}

	return specModel, nil
}

// GetDataHashByID missing godoc
func (r *repository) GetDataHashByID(ctx context.Context, tenantID string, id string, objectType model.SpecReferenceObjectType) (*string, error) {
	var specEntity Entity
	err := r.dataHashGetter.Get(ctx, objectType.GetResourceType(), tenantID, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &specEntity)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting data hash for Specification with id %q", id)
	}

	return repo.StringPtrFromNullableString(specEntity.SpecDataHash), nil
}

// Create missing godoc
func (r *repository) Create(ctx context.Context, tenant string, item *model.Spec) error {
	if item == nil {
		return apperrors.NewInternalError("item can not be empty")
	}

	entity := r.conv.ToEntity(item)

	return r.creator.Create(ctx, item.ObjectType.GetResourceType(), tenant, entity)
}

// ListIDByReferenceObjectID missing godoc
func (r *repository) ListIDByReferenceObjectID(ctx context.Context, tenant string, objectType model.SpecReferenceObjectType, objectID string) ([]string, error) {
	fieldName, err := r.referenceObjectFieldName(objectType)
	if err != nil {
		return nil, err
	}
	conditions := repo.Conditions{
		repo.NewEqualCondition(fieldName, objectID),
	}

	var specCollection SpecCollection
	err = r.idLister.List(ctx, objectType.GetResourceType(), tenant, &specCollection, conditions...)
	if err != nil {
		return nil, err
	}

	items := make([]string, 0, len(specCollection))

	for _, specEnt := range specCollection {
		items = append(items, specEnt.ID)
	}

	return items, nil
}

// ListByReferenceObjectID missing godoc
func (r *repository) ListByReferenceObjectID(ctx context.Context, tenant string, objectType model.SpecReferenceObjectType, objectID string) ([]*model.Spec, error) {
	fieldName, err := r.referenceObjectFieldName(objectType)
	if err != nil {
		return nil, err
	}
	conditions := repo.Conditions{
		repo.NewEqualCondition(fieldName, objectID),
	}

	var specCollection SpecCollection
	err = r.lister.List(ctx, objectType.GetResourceType(), tenant, &specCollection, conditions...)
	if err != nil {
		return nil, err
	}

	items := make([]*model.Spec, 0, len(specCollection))

	for _, specEnt := range specCollection {
		m, err := r.conv.FromEntity(&specEnt)
		if err != nil {
			return nil, err
		}

		items = append(items, m)
	}

	return items, nil
}

func (r *repository) ListByReferenceObjectIDs(ctx context.Context, tenant string, objectType model.SpecReferenceObjectType, objectIDs []string) ([]*model.Spec, error) {
	objectFieldName, err := r.referenceObjectFieldName(objectType)
	if err != nil {
		return nil, err
	}

	conditions := repo.Conditions{
		repo.NewNotNullCondition(objectFieldName),
	}

	var specs SpecCollection
	_, err = r.unionLister.List(ctx, objectType.GetResourceType(), tenant, objectIDs, objectFieldName, pageSize, cursor, orderByColumns, &specs, conditions...)
	if err != nil {
		return nil, err
	}

	specifications := make([]*model.Spec, 0, len(specs))
	for _, s := range specs {
		entity, err := r.conv.FromEntity(&s)
		if err != nil {
			return nil, err
		}
		specifications = append(specifications, entity)
	}

	return specifications, nil
}

// Delete missing godoc
func (r *repository) Delete(ctx context.Context, tenant, id string, objectType model.SpecReferenceObjectType) error {
	return r.deleter.DeleteOne(ctx, objectType.GetResourceType(), tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// DeleteByReferenceObjectID missing godoc
func (r *repository) DeleteByReferenceObjectID(ctx context.Context, tenant string, objectType model.SpecReferenceObjectType, objectID string) error {
	fieldName, err := r.referenceObjectFieldName(objectType)
	if err != nil {
		return err
	}

	return r.deleter.DeleteMany(ctx, objectType.GetResourceType(), tenant, repo.Conditions{repo.NewEqualCondition(fieldName, objectID)})
}

// Update missing godoc
func (r *repository) Update(ctx context.Context, tenant string, item *model.Spec) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(item)

	return r.updater.UpdateSingle(ctx, item.ObjectType.GetResourceType(), tenant, entity)
}

// Exists missing godoc
func (r *repository) Exists(ctx context.Context, tenantID, id string, objectType model.SpecReferenceObjectType) (bool, error) {
	return r.existQuerier.Exists(ctx, objectType.GetResourceType(), tenantID, repo.Conditions{repo.NewEqualCondition("id", id)})
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

// SpecCollection missing godoc
type SpecCollection []Entity

// Len missing godoc
func (r SpecCollection) Len() int {
	return len(r)
}
