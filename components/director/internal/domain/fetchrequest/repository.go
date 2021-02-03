package fetchrequest

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"
)

const fetchRequestTable string = `public.fetch_requests`

const documentIDColumn = "document_id"
const specIDColumn = "spec_id"

var (
	fetchRequestColumns = []string{"id", "tenant_id", documentIDColumn, "url", "auth", "mode", "filter", "status_condition", "status_message", "status_timestamp", specIDColumn}
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
	updater      repo.Updater
	conv         Converter
}

func NewRepository(conv Converter) *repository {
	return &repository{
		creator:      repo.NewCreator(resource.FetchRequest, fetchRequestTable, fetchRequestColumns),
		singleGetter: repo.NewSingleGetter(resource.FetchRequest, fetchRequestTable, tenantColumn, fetchRequestColumns),
		deleter:      repo.NewDeleter(resource.FetchRequest, fetchRequestTable, tenantColumn),
		updater:      repo.NewUpdater(resource.FetchRequest, fetchRequestTable, []string{"status_condition", "status_message", "status_timestamp"}, tenantColumn, []string{"id"}),
		conv:         conv,
	}
}

func (r *repository) Create(ctx context.Context, item *model.FetchRequest) error {
	if item == nil {
		return apperrors.NewInternalError("item can not be empty")
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

func (r *repository) Update(ctx context.Context, item *model.FetchRequest) error {

	entity, err := r.conv.ToEntity(*item)
	if err != nil {
		return err
	}
	return r.updater.UpdateSingle(ctx, entity)
}

func (r *repository) referenceObjectFieldName(objectType model.FetchRequestReferenceObjectType) (string, error) {
	switch objectType {
	case model.DocumentFetchRequestReference:
		return documentIDColumn, nil
	case model.SpecFetchRequestReference:
		return specIDColumn, nil
	}

	return "", apperrors.NewInternalError("Invalid type of the Fetch Request reference object")
}
