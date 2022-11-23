package bundlereferences

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/scope"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

const (
	// BundleReferenceTable represents the db name of the BundleReference table
	BundleReferenceTable string = `public.bundle_references`
	// APIDefTable represents the db name of the API Definitions table
	APIDefTable string = `api_definitions`
	// EventDefTable represents the db name of the Event Definitions table
	EventDefTable string = `event_api_definitions`

	// APIDefIDColumn represents the db column of the APIDefinition ID
	APIDefIDColumn string = "api_def_id"
	// APIDefURLColumn represents the db column of the APIDefinition default url
	APIDefURLColumn string = "api_def_url"
	// EventDefIDColumn represents the db column of the EventDefinition ID
	EventDefIDColumn string = "event_def_id"

	bundleIDColumn          string = "bundle_id"
	visibilityColumn        string = "visibility"
	internalVisibilityScope string = "internal_visibility:read"
	publicVisibilityValue   string = "public"
)

var (
	bundleReferencesColumns         = []string{"api_def_id", "event_def_id", "bundle_id", "api_def_url", "id", "is_default_bundle"}
	updatableColumns                = []string{"api_def_id", "event_def_id", "bundle_id", "api_def_url", "is_default_bundle"}
	updatableColumnsWithoutBundleID = []string{"api_def_id", "event_def_id", "api_def_url"}
)

// BundleReferenceConverter converts BundleReferences between the model.BundleReference service-layer representation and the repo-layer representation Entity.
//go:generate mockery --name=BundleReferenceConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type BundleReferenceConverter interface {
	ToEntity(in model.BundleReference) Entity
	FromEntity(in Entity) (model.BundleReference, error)
}

type repository struct {
	creator            repo.CreatorGlobal
	unionLister        repo.UnionListerGlobal
	lister             repo.ListerGlobal
	getter             repo.SingleGetterGlobal
	deleter            repo.DeleterGlobal
	updater            repo.UpdaterGlobal
	queryBuilderAPIs   repo.QueryBuilderGlobal
	queryBuilderEvents repo.QueryBuilderGlobal
	conv               BundleReferenceConverter
}

// NewRepository returns a new entity responsible for repo-layer BundleReference operations.
func NewRepository(conv BundleReferenceConverter) *repository {
	return &repository{
		creator:            repo.NewCreatorGlobal(resource.BundleReference, BundleReferenceTable, bundleReferencesColumns),
		unionLister:        repo.NewUnionListerGlobal(resource.BundleReference, BundleReferenceTable, []string{}),
		lister:             repo.NewListerGlobal(resource.BundleReference, BundleReferenceTable, bundleReferencesColumns),
		getter:             repo.NewSingleGetterGlobal(resource.BundleReference, BundleReferenceTable, bundleReferencesColumns),
		deleter:            repo.NewDeleterGlobal(resource.BundleReference, BundleReferenceTable),
		updater:            repo.NewUpdaterGlobal(resource.BundleReference, BundleReferenceTable, updatableColumns, []string{}),
		queryBuilderAPIs:   repo.NewQueryBuilderGlobal(resource.API, APIDefTable, []string{"id"}),
		queryBuilderEvents: repo.NewQueryBuilderGlobal(resource.EventDefinition, EventDefTable, []string{"id"}),
		conv:               conv,
	}
}

// BundleReferencesCollection is an array of Entities
type BundleReferencesCollection []Entity

// Len returns the length of the collection
func (r BundleReferencesCollection) Len() int {
	return len(r)
}

// GetByID retrieves the BundleReference with matching objectID/objectID and bundleID from the Compass storage.
func (r *repository) GetByID(ctx context.Context, objectType model.BundleReferenceObjectType, objectID, bundleID *string) (*model.BundleReference, error) {
	fieldName, err := r.referenceObjectFieldName(objectType)
	if err != nil {
		return nil, err
	}

	var bundleReferenceEntity Entity
	var conditions repo.Conditions

	if bundleID == nil {
		conditions = repo.Conditions{repo.NewEqualCondition(fieldName, objectID)}
	} else {
		conditions = repo.Conditions{
			repo.NewEqualCondition(fieldName, objectID),
			repo.NewEqualCondition(bundleIDColumn, bundleID),
		}
	}
	err = r.getter.GetGlobal(ctx, conditions, repo.NoOrderBy, &bundleReferenceEntity)
	if err != nil {
		return nil, err
	}

	bundleReferenceModel, err := r.conv.FromEntity(bundleReferenceEntity)
	if err != nil {
		return nil, err
	}

	return &bundleReferenceModel, nil
}

// GetBundleIDsForObject retrieves all BundleReference IDs for matching objectID from the Compass storage.
func (r *repository) GetBundleIDsForObject(ctx context.Context, objectType model.BundleReferenceObjectType, objectID *string) (ids []string, err error) {
	fieldName, err := r.referenceObjectFieldName(objectType)
	if err != nil {
		return nil, err
	}

	var objectBundleIDs IDs

	lister := r.lister.Clone()
	lister.SetSelectedColumns([]string{bundleIDColumn})

	conditions := repo.Conditions{
		repo.NewEqualCondition(fieldName, objectID),
	}

	err = lister.ListGlobal(ctx, &objectBundleIDs, conditions...)
	if err != nil {
		return nil, err
	}

	return objectBundleIDs, nil
}

// Create adds the provided BundleReference into the Compass storage.
func (r *repository) Create(ctx context.Context, item *model.BundleReference) error {
	if item == nil {
		return apperrors.NewInternalError("item can not be empty")
	}

	entity := r.conv.ToEntity(*item)

	return r.creator.Create(ctx, entity)
}

// Update updates the provided BundleReference.
func (r *repository) Update(ctx context.Context, item *model.BundleReference) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	fieldName, err := r.referenceObjectFieldName(item.ObjectType)
	if err != nil {
		return err
	}

	updater := r.updater.Clone()

	idColumns := make([]string, 0)
	if item.BundleID == nil {
		idColumns = append(idColumns, fieldName)
		updater.SetUpdatableColumns(updatableColumnsWithoutBundleID)
	} else {
		idColumns = append(idColumns, fieldName, bundleIDColumn)
	}

	updater.SetIDColumns(idColumns)

	entity := r.conv.ToEntity(*item)

	return updater.UpdateSingleGlobal(ctx, entity)
}

// DeleteByReferenceObjectID removes a BundleReference with matching objectID and bundleID from the Compass storage.
func (r *repository) DeleteByReferenceObjectID(ctx context.Context, bundleID string, objectType model.BundleReferenceObjectType, objectID string) error {
	fieldName, err := r.referenceObjectFieldName(objectType)
	if err != nil {
		return err
	}

	conditions := repo.Conditions{
		repo.NewEqualCondition(fieldName, objectID),
		repo.NewEqualCondition(bundleIDColumn, bundleID),
	}

	return r.deleter.DeleteOneGlobal(ctx, conditions)
}

// ListByBundleIDs retrieves all BundleReferences matching an array of bundleIDs from the Compass storage.
func (r *repository) ListByBundleIDs(ctx context.Context, objectType model.BundleReferenceObjectType, bundleIDs []string, pageSize int, cursor string) ([]*model.BundleReference, map[string]int, error) {
	objectTable, objectIDCol, columns, err := getDetailsByObjectType(objectType)
	if err != nil {
		return nil, nil, err
	}

	unionLister := r.unionLister.Clone()
	unionLister.SetSelectedColumns(columns)

	objectFieldName, err := r.referenceObjectFieldName(objectType)
	if err != nil {
		return nil, nil, err
	}

	isInternalVisibilityScopePresent, err := scope.Contains(ctx, internalVisibilityScope)
	if err != nil {
		log.C(ctx).Infof("No scopes are present in the context meaning the flow is not user-initiated. Processing %ss without visibility check...", objectType)
		isInternalVisibilityScopePresent = true
	}

	queryBuilder := r.queryBuilderAPIs
	if objectTable == EventDefTable {
		queryBuilder = r.queryBuilderEvents
	}

	var conditions repo.Conditions
	if !isInternalVisibilityScopePresent {
		log.C(ctx).Infof("No internal visibility scope is present in the context. Processing only public %ss...", objectType)

		query, args, err := queryBuilder.BuildQueryGlobal(false, repo.NewEqualCondition(visibilityColumn, publicVisibilityValue))
		if err != nil {
			return nil, nil, err
		}
		conditions = append(conditions, repo.NewInConditionForSubQuery(objectIDCol, query, args))
	}

	log.C(ctx).Infof("Internal visibility scope is present in the context. Processing %ss without visibility check...", objectType)
	conditions = append(conditions, repo.NewNotNullCondition(objectFieldName))

	orderByColumns, err := getOrderByColumnsByObjectType(objectType)
	if err != nil {
		return nil, nil, err
	}

	var objectBundleIDs BundleReferencesCollection
	counts, err := unionLister.ListGlobal(ctx, bundleIDs, bundleIDColumn, pageSize, cursor, orderByColumns, &objectBundleIDs, conditions...)
	if err != nil {
		return nil, nil, err
	}

	bundleReferences := make([]*model.BundleReference, 0, len(objectBundleIDs))
	for _, d := range objectBundleIDs {
		entity, err := r.conv.FromEntity(d)
		if err != nil {
			return nil, nil, err
		}
		bundleReferences = append(bundleReferences, &entity)
	}

	return bundleReferences, counts, nil
}

func (r *repository) referenceObjectFieldName(objectType model.BundleReferenceObjectType) (string, error) {
	switch objectType {
	case model.BundleAPIReference:
		return APIDefIDColumn, nil
	case model.BundleEventReference:
		return EventDefIDColumn, nil
	}
	return "", apperrors.NewInternalError("Invalid type of the BundleReference object")
}

func getDetailsByObjectType(objectType model.BundleReferenceObjectType) (string, string, []string, error) {
	switch objectType {
	case model.BundleAPIReference:
		return APIDefTable, APIDefIDColumn, []string{APIDefIDColumn, bundleIDColumn, APIDefURLColumn}, nil
	case model.BundleEventReference:
		return EventDefTable, EventDefIDColumn, []string{EventDefIDColumn, bundleIDColumn}, nil
	}
	return "", "", []string{""}, apperrors.NewInternalError("Invalid type of the BundleReference object")
}

func getOrderByColumnsByObjectType(objectType model.BundleReferenceObjectType) (repo.OrderByParams, error) {
	switch objectType {
	case model.BundleAPIReference:
		return repo.OrderByParams{repo.NewAscOrderBy(APIDefIDColumn), repo.NewAscOrderBy(bundleIDColumn), repo.NewAscOrderBy(APIDefURLColumn)}, nil
	case model.BundleEventReference:
		return repo.OrderByParams{repo.NewAscOrderBy(EventDefIDColumn), repo.NewAscOrderBy(bundleIDColumn)}, nil
	}
	return nil, apperrors.NewInternalError("Invalid type of the BundleReference object")
}

// IDs keeps IDs retrieved from the Compass storage.
type IDs []string

// Len returns the length of the IDs
func (i IDs) Len() int {
	return len(i)
}
