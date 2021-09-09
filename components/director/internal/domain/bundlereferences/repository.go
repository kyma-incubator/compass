package bundlereferences

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

const (
	// BundleReferenceTable missing godoc
	BundleReferenceTable string = `public.bundle_references`

	// APIDefIDColumn missing godoc
	APIDefIDColumn string = "api_def_id"
	// APIDefURLColumn missing godoc
	APIDefURLColumn string = "api_def_url"
	// EventDefIDColumn missing godoc
	EventDefIDColumn string = "event_def_id"
	bundleIDColumn   string = "bundle_id"
)

var (
	tenantColumn                    = "tenant_id"
	bundleReferencesColumns         = []string{"tenant_id", "api_def_id", "event_def_id", "bundle_id", "api_def_url", "id"}
	updatableColumns                = []string{"api_def_id", "event_def_id", "bundle_id", "api_def_url"}
	updatableColumnsWithoutBundleID = []string{"api_def_id", "event_def_id", "api_def_url"}
)

// BundleReferenceConverter missing godoc
//go:generate mockery --name=BundleReferenceConverter --output=automock --outpkg=automock --case=underscore
type BundleReferenceConverter interface {
	ToEntity(in model.BundleReference) Entity
	FromEntity(in Entity) (model.BundleReference, error)
}

type repository struct {
	creator     repo.Creator
	unionLister repo.UnionLister
	lister      repo.Lister
	getter      repo.SingleGetter
	deleter     repo.Deleter
	updater     repo.Updater
	conv        BundleReferenceConverter
}

// NewRepository missing godoc
func NewRepository(conv BundleReferenceConverter) *repository {
	return &repository{
		creator:     repo.NewCreator(resource.BundleReference, BundleReferenceTable, bundleReferencesColumns),
		unionLister: repo.NewUnionLister(resource.BundleReference, BundleReferenceTable, tenantColumn, []string{}),
		lister:      repo.NewLister(resource.BundleReference, BundleReferenceTable, tenantColumn, bundleReferencesColumns),
		getter:      repo.NewSingleGetter(resource.BundleReference, BundleReferenceTable, tenantColumn, bundleReferencesColumns),
		deleter:     repo.NewDeleter(resource.BundleReference, BundleReferenceTable, tenantColumn),
		updater:     repo.NewUpdater(resource.BundleReference, BundleReferenceTable, updatableColumns, tenantColumn, []string{}),
		conv:        conv,
	}
}

// BundleReferencesCollection missing godoc
type BundleReferencesCollection []Entity

// Len missing godoc
func (r BundleReferencesCollection) Len() int {
	return len(r)
}

// GetByID missing godoc
func (r *repository) GetByID(ctx context.Context, objectType model.BundleReferenceObjectType, tenantID string, objectID, bundleID *string) (*model.BundleReference, error) {
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
	err = r.getter.Get(ctx, tenantID, conditions, repo.NoOrderBy, &bundleReferenceEntity)
	if err != nil {
		return nil, err
	}

	bundleReferenceModel, err := r.conv.FromEntity(bundleReferenceEntity)
	if err != nil {
		return nil, err
	}

	return &bundleReferenceModel, nil
}

// GetBundleIDsForObject missing godoc
func (r *repository) GetBundleIDsForObject(ctx context.Context, tenantID string, objectType model.BundleReferenceObjectType, objectID *string) (ids []string, err error) {
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

	err = lister.List(ctx, tenantID, &objectBundleIDs, conditions...)
	if err != nil {
		return nil, err
	}

	return objectBundleIDs, nil
}

// Create missing godoc
func (r *repository) Create(ctx context.Context, item *model.BundleReference) error {
	if item == nil {
		return apperrors.NewInternalError("item can not be empty")
	}

	entity := r.conv.ToEntity(*item)

	return r.creator.Create(ctx, entity)
}

// Update missing godoc
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

	return updater.UpdateSingle(ctx, entity)
}

// DeleteByReferenceObjectID missing godoc
func (r *repository) DeleteByReferenceObjectID(ctx context.Context, tenant, bundleID string, objectType model.BundleReferenceObjectType, objectID string) error {
	fieldName, err := r.referenceObjectFieldName(objectType)
	if err != nil {
		return err
	}

	conditions := repo.Conditions{
		repo.NewEqualCondition(fieldName, objectID),
		repo.NewEqualCondition(bundleIDColumn, bundleID),
	}

	return r.deleter.DeleteOne(ctx, tenant, conditions)
}

// ListByBundleIDs missing godoc
func (r *repository) ListByBundleIDs(ctx context.Context, objectType model.BundleReferenceObjectType, tenantID string, bundleIDs []string, pageSize int, cursor string) ([]*model.BundleReference, map[string]int, error) {
	columns, err := getSelectedColumnsByObjectType(objectType)
	if err != nil {
		return nil, nil, err
	}

	unionLister := r.unionLister.Clone()
	unionLister.SetSelectedColumns(columns)

	objectFieldName, err := r.referenceObjectFieldName(objectType)
	if err != nil {
		return nil, nil, err
	}

	conditions := repo.Conditions{
		repo.NewNotNullCondition(objectFieldName),
	}

	orderByColumns, err := getOrderByColumnsByObjectType(objectType)
	if err != nil {
		return nil, nil, err
	}

	var objectBundleIDs BundleReferencesCollection
	counts, err := unionLister.List(ctx, tenantID, bundleIDs, bundleIDColumn, pageSize, cursor, orderByColumns, &objectBundleIDs, conditions...)
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

func getSelectedColumnsByObjectType(objectType model.BundleReferenceObjectType) ([]string, error) {
	switch objectType {
	case model.BundleAPIReference:
		return []string{APIDefIDColumn, bundleIDColumn, APIDefURLColumn}, nil
	case model.BundleEventReference:
		return []string{EventDefIDColumn, bundleIDColumn}, nil
	}
	return []string{""}, apperrors.NewInternalError("Invalid type of the BundleReference object")
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

// IDs missing godoc
type IDs []string

// Len missing godoc
func (i IDs) Len() int {
	return len(i)
}
