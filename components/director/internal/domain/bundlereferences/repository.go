package bundlereferences

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

const (
	BundleReferenceTable string = `public.bundle_references`

	APIDefIDColumn   string = "api_def_id"
	EventDefIDColumn string = "event_def_id"
	bundleIDColumn   string = "bundle_id"
)

var (
	tenantColumn                    = "tenant_id"
	bundleReferencesColumns         = []string{"tenant_id", "api_def_id", "event_def_id", "bundle_id", "api_def_url"}
	updatableColumns                = []string{"api_def_id", "event_def_id", "bundle_id", "api_def_url"}
	updatableColumnsWithoutBundleID = []string{"api_def_id", "event_def_id", "api_def_url"}
)

//go:generate mockery --name=BundleReferenceConverter --output=automock --outpkg=automock --case=underscore
type BundleReferenceConverter interface {
	ToEntity(in model.BundleReference) Entity
	FromEntity(in Entity) (model.BundleReference, error)
}

type repository struct {
	creator repo.Creator
	lister  repo.Lister
	getter  repo.SingleGetter
	deleter repo.Deleter
	updater repo.Updater
	conv    BundleReferenceConverter
}

func NewRepository(conv BundleReferenceConverter) *repository {
	return &repository{
		creator: repo.NewCreator(resource.BundleReference, BundleReferenceTable, bundleReferencesColumns),
		lister:  repo.NewLister(resource.BundleReference, BundleReferenceTable, tenantColumn, bundleReferencesColumns),
		getter:  repo.NewSingleGetter(resource.BundleReference, BundleReferenceTable, tenantColumn, bundleReferencesColumns),
		deleter: repo.NewDeleter(resource.BundleReference, BundleReferenceTable, tenantColumn),
		updater: repo.NewUpdater(resource.BundleReference, BundleReferenceTable, updatableColumns, tenantColumn, []string{}),
		conv:    conv,
	}
}

type BundleReferencesCollection []Entity

func (r BundleReferencesCollection) Len() int {
	return len(r)
}

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

func (r *repository) Create(ctx context.Context, item *model.BundleReference) error {
	if item == nil {
		return apperrors.NewInternalError("item can not be empty")
	}

	entity := r.conv.ToEntity(*item)

	return r.creator.Create(ctx, entity)
}

func (r *repository) Update(ctx context.Context, item *model.BundleReference) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	fieldName, err := r.referenceObjectFieldName(item.ObjectType)
	if err != nil {
		return err
	}

	updater := r.updater.Clone()

	idColumns := make([]string, 0, 0)
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

func (r *repository) referenceObjectFieldName(objectType model.BundleReferenceObjectType) (string, error) {
	switch objectType {
	case model.BundleAPIReference:
		return APIDefIDColumn, nil
	case model.BundleEventReference:
		return EventDefIDColumn, nil
	}
	return "", apperrors.NewInternalError("Invalid type of the BundleReference object")
}

type IDs []string

func (i IDs) Len() int {
	return len(i)
}
