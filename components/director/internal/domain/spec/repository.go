package spec

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"
)

const (
	specificationsTable = `public.specifications`
	apiDefIDColumn      = "api_def_id"
	eventAPIDefIDColumn = "event_def_id"
	capabilityIDColumn  = "capability_id"
	pageSize            = 1
	cursor              = ""
)

var (
	specificationsColumns = []string{"id", apiDefIDColumn, eventAPIDefIDColumn, capabilityIDColumn, "spec_data", "api_spec_format", "api_spec_type", "event_spec_format", "event_spec_type", "capability_spec_format", "capability_spec_type", "custom_type"}
	updatableColumns      = []string{"spec_data", "api_spec_format", "api_spec_type", "event_spec_format", "event_spec_type", "capability_spec_format", "capability_spec_type"}
	idColumns             = []string{"id"}
	orderByColumns        = repo.OrderByParams{repo.NewAscOrderBy("created_at"), repo.NewAscOrderBy("id")}
)

// Converter missing godoc
//
//go:generate mockery --name=Converter --output=automock --outpkg=automock --case=underscore --disable-version-string
type Converter interface {
	ToEntity(in *model.Spec) *Entity
	FromEntity(in *Entity) (*model.Spec, error)
}

type repository struct {
	creator        repo.Creator
	creatorGlobal  repo.CreatorGlobal
	lister         repo.Lister
	listerGlobal   repo.ListerGlobal
	idLister       repo.Lister
	idListerGlobal repo.ListerGlobal
	unionLister    repo.UnionLister
	getter         repo.SingleGetter
	getterGlobal   repo.SingleGetterGlobal
	deleter        repo.Deleter
	deleterGlobal  repo.DeleterGlobal
	updater        repo.Updater
	updaterGlobal  repo.UpdaterGlobal
	existQuerier   repo.ExistQuerier
	conv           Converter
}

// NewRepository missing godoc
func NewRepository(conv Converter) *repository {
	return &repository{
		creator:       repo.NewCreator(specificationsTable, specificationsColumns),
		creatorGlobal: repo.NewCreatorGlobal(resource.Specification, specificationsTable, specificationsColumns),
		getter:        repo.NewSingleGetter(specificationsTable, specificationsColumns),
		getterGlobal:  repo.NewSingleGetterGlobal(resource.Specification, specificationsTable, specificationsColumns),
		lister: repo.NewListerWithOrderBy(specificationsTable, specificationsColumns, repo.OrderByParams{
			{
				Field: "created_at",
				Dir:   repo.AscOrderBy,
			},
		}),
		listerGlobal: repo.NewListerGlobal(resource.Specification, specificationsTable, specificationsColumns),
		idLister: repo.NewListerWithOrderBy(specificationsTable, idColumns, repo.OrderByParams{
			{
				Field: "created_at",
				Dir:   repo.AscOrderBy,
			},
		}),
		idListerGlobal: repo.NewListerGlobalWithOrderBy(resource.Specification, specificationsTable, idColumns, repo.OrderByParams{
			{
				Field: "created_at",
				Dir:   repo.AscOrderBy,
			},
		}),
		unionLister:   repo.NewUnionLister(specificationsTable, specificationsColumns),
		deleter:       repo.NewDeleter(specificationsTable),
		deleterGlobal: repo.NewDeleterGlobal(resource.Specification, specificationsTable),
		updater:       repo.NewUpdater(specificationsTable, updatableColumns, idColumns),
		updaterGlobal: repo.NewUpdaterGlobal(resource.Specification, specificationsTable, updatableColumns, idColumns),
		existQuerier:  repo.NewExistQuerier(specificationsTable),
		conv:          conv,
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

// GetByIDGlobal gets a Spec by ID without tenant isolation
func (r *repository) GetByIDGlobal(ctx context.Context, id string) (*model.Spec, error) {
	var specEntity Entity
	err := r.getterGlobal.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &specEntity)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Specification with id %q", id)
	}

	specModel, err := r.conv.FromEntity(&specEntity)
	if err != nil {
		return nil, err
	}

	return specModel, nil
}

// Create creates a spec in the scope of a tenant
func (r *repository) Create(ctx context.Context, tenant string, item *model.Spec) error {
	if item == nil {
		return apperrors.NewInternalError("item can not be empty")
	}

	entity := r.conv.ToEntity(item)

	return r.creator.Create(ctx, item.ObjectType.GetResourceType(), tenant, entity)
}

// CreateGlobal create a spec without a tenant isolation
func (r *repository) CreateGlobal(ctx context.Context, item *model.Spec) error {
	if item == nil {
		return apperrors.NewInternalError("item can not be empty")
	}

	entity := r.conv.ToEntity(item)

	return r.creatorGlobal.Create(ctx, entity)
}

// ListIDByReferenceObjectID retrieves all spec ids by objectType and objectID
func (r *repository) ListIDByReferenceObjectID(ctx context.Context, tenant string, objectType model.SpecReferenceObjectType, objectID string) ([]string, error) {
	conditions, err := r.buildReferenceObjectIDConditions(objectType, objectID)
	if err != nil {
		return nil, err
	}

	var specCollection SpecCollection
	err = r.idLister.List(ctx, objectType.GetResourceType(), tenant, &specCollection, conditions...)
	if err != nil {
		return nil, err
	}

	return extractIDsFromCollection(specCollection), nil
}

// ListIDByReferenceObjectIDGlobal retrieves all spec ids by objectType and objectID
func (r *repository) ListIDByReferenceObjectIDGlobal(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string) ([]string, error) {
	conditions, err := r.buildReferenceObjectIDConditions(objectType, objectID)
	if err != nil {
		return nil, err
	}

	var specCollection SpecCollection
	err = r.idListerGlobal.ListGlobal(ctx, &specCollection, conditions...)
	if err != nil {
		return nil, err
	}

	return extractIDsFromCollection(specCollection), nil
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

// ListByReferenceObjectIDGlobal lists specs by a model.SpecReferenceObjectType without tenant isolation
func (r *repository) ListByReferenceObjectIDGlobal(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string) ([]*model.Spec, error) {
	fieldName, err := r.referenceObjectFieldName(objectType)
	if err != nil {
		return nil, err
	}
	conditions := repo.Conditions{
		repo.NewEqualCondition(fieldName, objectID),
	}

	var specCollection SpecCollection
	err = r.listerGlobal.ListGlobal(ctx, &specCollection, conditions...)
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

// DeleteByReferenceObjectIDGlobal deletes a reference object with a given ID without tenant isolation
func (r *repository) DeleteByReferenceObjectIDGlobal(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string) error {
	fieldName, err := r.referenceObjectFieldName(objectType)
	if err != nil {
		return err
	}

	return r.deleterGlobal.DeleteManyGlobal(ctx, repo.Conditions{repo.NewEqualCondition(fieldName, objectID)})
}

// Update missing godoc
func (r *repository) Update(ctx context.Context, tenant string, item *model.Spec) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(item)

	return r.updater.UpdateSingle(ctx, item.ObjectType.GetResourceType(), tenant, entity)
}

// UpdateGlobal updates a Spec without tenant isolation
func (r *repository) UpdateGlobal(ctx context.Context, item *model.Spec) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(item)

	return r.updaterGlobal.UpdateSingleGlobal(ctx, entity)
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
	case model.CapabilitySpecReference:
		return capabilityIDColumn, nil
	}

	return "", apperrors.NewInternalError("Invalid type of the Specification reference object")
}

func (r *repository) buildReferenceObjectIDConditions(objectType model.SpecReferenceObjectType, objectID string) (repo.Conditions, error) {
	fieldName, err := r.referenceObjectFieldName(objectType)
	if err != nil {
		return nil, err
	}

	return repo.Conditions{
		repo.NewEqualCondition(fieldName, objectID),
	}, nil
}

// SpecCollection missing godoc
type SpecCollection []Entity

// Len missing godoc
func (r SpecCollection) Len() int {
	return len(r)
}

func extractIDsFromCollection(specCollection SpecCollection) []string {
	items := make([]string, 0, len(specCollection))

	for _, specEnt := range specCollection {
		items = append(items, specEnt.ID)
	}

	return items
}
