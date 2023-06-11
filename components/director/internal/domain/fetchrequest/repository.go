package fetchrequest

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"
)

const fetchRequestTable string = `public.fetch_requests`

const documentIDColumn = "document_id"
const specIDColumn = "spec_id"

var (
	fetchRequestColumns = []string{"id", documentIDColumn, "url", "auth", "mode", "filter", "status_condition", "status_message", "status_timestamp", specIDColumn}
)

// Converter missing godoc
//
//go:generate mockery --name=Converter --output=automock --outpkg=automock --case=underscore --disable-version-string
type Converter interface {
	ToEntity(in *model.FetchRequest) (*Entity, error)
	FromEntity(in *Entity, objectType model.FetchRequestReferenceObjectType) (*model.FetchRequest, error)
}

type repository struct {
	creator       repo.Creator
	creatorGlobal repo.CreatorGlobal
	singleGetter  repo.SingleGetter
	lister        repo.Lister
	deleter       repo.Deleter
	deleterGlobal repo.DeleterGlobal
	updater       repo.Updater
	updaterGlobal repo.UpdaterGlobal
	conv          Converter
}

// NewRepository missing godoc
func NewRepository(conv Converter) *repository {
	return &repository{
		creator:       repo.NewCreator(fetchRequestTable, fetchRequestColumns),
		creatorGlobal: repo.NewCreatorGlobal(resource.FetchRequest, fetchRequestTable, fetchRequestColumns),
		singleGetter:  repo.NewSingleGetter(fetchRequestTable, fetchRequestColumns),
		lister:        repo.NewLister(fetchRequestTable, fetchRequestColumns),
		deleter:       repo.NewDeleter(fetchRequestTable),
		deleterGlobal: repo.NewDeleterGlobal(resource.FetchRequest, fetchRequestTable),
		updater:       repo.NewUpdater(fetchRequestTable, []string{"status_condition", "status_message", "status_timestamp"}, []string{"id"}),
		updaterGlobal: repo.NewUpdaterGlobal(resource.FetchRequest, fetchRequestTable, []string{"status_condition", "status_message", "status_timestamp"}, []string{"id"}),
		conv:          conv,
	}
}

// Create missing godoc
func (r *repository) Create(ctx context.Context, tenant string, item *model.FetchRequest) error {
	if item == nil {
		return apperrors.NewInternalError("item can not be empty")
	}

	entity, err := r.conv.ToEntity(item)
	if err != nil {
		return errors.Wrap(err, "while creating FetchRequest entity from model")
	}

	return r.creator.Create(ctx, item.ObjectType.GetResourceType(), tenant, entity)
}

// CreateGlobal creates a fetch request without tenant isolation
func (r *repository) CreateGlobal(ctx context.Context, item *model.FetchRequest) error {
	if item == nil {
		return apperrors.NewInternalError("item can not be empty")
	}

	entity, err := r.conv.ToEntity(item)
	if err != nil {
		return errors.Wrap(err, "while creating FetchRequest entity from model")
	}

	return r.creatorGlobal.Create(ctx, entity)
}

// GetByReferenceObjectID missing godoc
func (r *repository) GetByReferenceObjectID(ctx context.Context, tenant string, objectType model.FetchRequestReferenceObjectType, objectID string) (*model.FetchRequest, error) {
	fieldName, err := r.referenceObjectFieldName(objectType)
	if err != nil {
		return nil, err
	}

	var entity Entity
	if err := r.singleGetter.Get(ctx, objectType.GetResourceType(), tenant, repo.Conditions{repo.NewEqualCondition(fieldName, objectID)}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}

	frModel, err := r.conv.FromEntity(&entity, objectType)
	if err != nil {
		return nil, errors.Wrap(err, "while getting FetchRequest model from entity")
	}

	return frModel, nil
}

// Delete missing godoc
func (r *repository) Delete(ctx context.Context, tenant, id string, objectType model.FetchRequestReferenceObjectType) error {
	return r.deleter.DeleteOne(ctx, objectType.GetResourceType(), tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// DeleteByReferenceObjectID missing godoc
func (r *repository) DeleteByReferenceObjectID(ctx context.Context, tenant string, objectType model.FetchRequestReferenceObjectType, objectID string) error {
	fieldName, err := r.referenceObjectFieldName(objectType)
	if err != nil {
		return err
	}

	return r.deleter.DeleteMany(ctx, objectType.GetResourceType(), tenant, repo.Conditions{repo.NewEqualCondition(fieldName, objectID)})
}

// DeleteByReferenceObjectIDGlobal deletes fetch request by model.FetchRequestReferenceObjectType without tenant isolation
func (r *repository) DeleteByReferenceObjectIDGlobal(ctx context.Context, objectType model.FetchRequestReferenceObjectType, objectID string) error {
	fieldName, err := r.referenceObjectFieldName(objectType)
	if err != nil {
		return err
	}

	return r.deleterGlobal.DeleteManyGlobal(ctx, repo.Conditions{repo.NewEqualCondition(fieldName, objectID)})
}

// Update missing godoc
func (r *repository) Update(ctx context.Context, tenant string, item *model.FetchRequest) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}
	entity, err := r.conv.ToEntity(item)
	if err != nil {
		return err
	}
	return r.updater.UpdateSingle(ctx, item.ObjectType.GetResourceType(), tenant, entity)
}

// UpdateGlobal updates a fetch request globally without tenant isolation
func (r *repository) UpdateGlobal(ctx context.Context, item *model.FetchRequest) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}
	entity, err := r.conv.ToEntity(item)
	if err != nil {
		return err
	}
	return r.updaterGlobal.UpdateSingleGlobal(ctx, entity)
}

func (r *repository) ListByReferenceObjectIDs(ctx context.Context, tenant string, objectType model.FetchRequestReferenceObjectType, objectIDs []string) ([]*model.FetchRequest, error) {
	fieldName, err := r.referenceObjectFieldName(objectType)
	if err != nil {
		return nil, err
	}

	var fetchRequestCollection FetchRequestsCollection

	var conditions repo.Conditions
	if len(objectIDs) > 0 {
		conditions = repo.Conditions{
			repo.NewInConditionForStringValues(fieldName, objectIDs),
		}
	}
	if err := r.lister.List(ctx, objectType.GetResourceType(), tenant, &fetchRequestCollection, conditions...); err != nil {
		return nil, err
	}

	fetchRequestsByID := map[string]*model.FetchRequest{}
	for _, fetchRequestEnt := range fetchRequestCollection {
		m, err := r.conv.FromEntity(&fetchRequestEnt, objectType)
		if err != nil {
			return nil, errors.Wrap(err, "while creating FetchRequest model from entity")
		}

		if fieldName == specIDColumn {
			fetchRequestsByID[fetchRequestEnt.SpecID.String] = m
		} else if fieldName == documentIDColumn {
			fetchRequestsByID[fetchRequestEnt.DocumentID.String] = m
		}
	}

	fetchRequests := make([]*model.FetchRequest, 0, len(objectIDs))
	for _, objectID := range objectIDs {
		fetchRequests = append(fetchRequests, fetchRequestsByID[objectID])
	}

	return fetchRequests, nil
}

func (r *repository) referenceObjectFieldName(objectType model.FetchRequestReferenceObjectType) (string, error) {
	switch objectType {
	case model.DocumentFetchRequestReference:
		return documentIDColumn, nil
	case model.EventSpecFetchRequestReference:
		fallthrough
	case model.APISpecFetchRequestReference:
		return specIDColumn, nil
	}

	return "", apperrors.NewInternalError("Invalid type of the Fetch Request reference object")
}

// FetchRequestsCollection missing godoc
type FetchRequestsCollection []Entity

// Len missing godoc
func (r FetchRequestsCollection) Len() int {
	return len(r)
}
