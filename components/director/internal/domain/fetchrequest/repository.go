package fetchrequest

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"
)

const fetchRequestTable string = `public.fetch_requests`

const documentIDColumn = "document_id"
const apiDefIDColumn = "api_def_id"
const eventAPIDefIDColumn = "event_api_def_id"

var (
	fetchRequestColumns = []string{"id", "tenant_id", apiDefIDColumn, eventAPIDefIDColumn, documentIDColumn, "url", "auth", "mode", "filter", "status_condition", "status_timestamp"}
	tenantColumn        = "tenant_id"
)

//go:generate mockery -name=Converter -output=automock -outpkg=automock -case=underscore
type Converter interface {
	ToEntity(in model.FetchRequest) (Entity, error)
	FromEntity(in Entity) (model.FetchRequest, error)
}

type repository struct {
	creator      repo.Creator
	singleGetter repo.SingleGetter
	deleter      repo.Deleter
	conv         Converter
}

func NewRepository(conv Converter) *repository {
	return &repository{
		creator:      repo.NewCreator(fetchRequestTable, fetchRequestColumns),
		singleGetter: repo.NewSingleGetter(fetchRequestTable, tenantColumn, fetchRequestColumns),
		deleter:      repo.NewDeleter(fetchRequestTable, tenantColumn),
		conv:         conv,
	}
}

func (r *repository) Create(ctx context.Context, item *model.FetchRequest) error {
	if item == nil {
		return errors.New("item can not be empty")
	}

	entity, err := r.conv.ToEntity(*item)
	if err != nil {
		return errors.Wrap(err, "while creating FetchRequest entity from model")
	}

	return r.creator.Create(ctx, entity)
}

func (r *repository) GetByReferenceObjectID(ctx context.Context, tenant string, objectType model.FetchRequestReferenceObjectType, objectID string) (*model.FetchRequest, error) {
	fieldName, err := r.referenceObjectFieldName(objectType)
	if err != nil {
		return nil, err
	}

	var entity Entity
	if err := r.singleGetter.Get(ctx, tenant, repo.Conditions{repo.NewEqualCondition(fieldName, objectID)}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}

	frModel, err := r.conv.FromEntity(entity)
	if err != nil {
		return nil, errors.Wrap(err, "while getting FetchRequest model from entity")
	}

	return &frModel, nil
}

func (r *repository) Delete(ctx context.Context, tenant, id string) error {
	return r.deleter.DeleteOne(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *repository) DeleteByReferenceObjectID(ctx context.Context, tenant string, objectType model.FetchRequestReferenceObjectType, objectID string) error {
	fieldName, err := r.referenceObjectFieldName(objectType)
	if err != nil {
		return err
	}

	return r.deleter.DeleteMany(ctx, tenant, repo.Conditions{repo.NewEqualCondition(fieldName, objectID)})
}

func (r *repository) referenceObjectFieldName(objectType model.FetchRequestReferenceObjectType) (string, error) {
	switch objectType {
	case model.DocumentFetchRequestReference:
		return documentIDColumn, nil
	case model.APIFetchRequestReference:
		return apiDefIDColumn, nil
	case model.EventAPIFetchRequestReference:
		return eventAPIDefIDColumn, nil
	}

	return "", errors.New("Invalid type of the Fetch Request reference object")
}
