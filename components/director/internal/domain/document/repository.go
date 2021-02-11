package document

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const documentTable = "public.documents"

var (
	documentColumns = []string{"id", "tenant_id", "bundle_id", "title", "display_name", "description", "format", "kind", "data", "ready", "created_at", "updated_at", "deleted_at", "error"}
	tenantColumn    = "tenant_id"
)

//go:generate mockery -name=Converter -output=automock -outpkg=automock -case=underscore
type Converter interface {
	ToEntity(in model.Document) (*Entity, error)
	FromEntity(in Entity) (model.Document, error)
}

type repository struct {
	existQuerier    repo.ExistQuerier
	singleGetter    repo.SingleGetter
	deleter         repo.Deleter
	pageableQuerier repo.PageableQuerier
	creator         repo.Creator

	conv Converter
}

func NewRepository(conv Converter) *repository {
	return &repository{
		existQuerier:    repo.NewExistQuerier(resource.Document, documentTable, tenantColumn),
		singleGetter:    repo.NewSingleGetter(resource.Document, documentTable, tenantColumn, documentColumns),
		deleter:         repo.NewDeleter(resource.Document, documentTable, tenantColumn),
		pageableQuerier: repo.NewPageableQuerier(resource.Document, documentTable, tenantColumn, documentColumns),
		creator:         repo.NewCreator(resource.Document, documentTable, documentColumns),
		conv:            conv,
	}
}

func (r *repository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *repository) GetByID(ctx context.Context, tenant, id string) (*model.Document, error) {
	var entity Entity
	if err := r.singleGetter.Get(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}

	docModel, err := r.conv.FromEntity(entity)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Document entity to model")
	}

	return &docModel, nil
}

func (r *repository) GetForBundle(ctx context.Context, tenant string, id string, bundleID string) (*model.Document, error) {
	var ent Entity

	conditions := repo.Conditions{
		repo.NewEqualCondition("id", id),
		repo.NewEqualCondition("bundle_id", bundleID),
	}
	if err := r.singleGetter.Get(ctx, tenant, conditions, repo.NoOrderBy, &ent); err != nil {
		return nil, err
	}

	documentModel, err := r.conv.FromEntity(ent)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Document entity to model")
	}

	return &documentModel, nil
}

func (r *repository) Create(ctx context.Context, item *model.Document) error {
	if item == nil {
		return apperrors.NewInternalError("Document cannot be empty")
	}

	entity, err := r.conv.ToEntity(*item)
	if err != nil {
		return errors.Wrap(err, "while creating Document entity from model")
	}

	log.C(ctx).Debugf("Persisting Document entity with id %s to db", item.ID)
	return r.creator.Create(ctx, entity)
}

func (r *repository) CreateMany(ctx context.Context, items []*model.Document) error {
	for _, item := range items {
		if item == nil {
			return apperrors.NewInternalError("Document cannot be empty")
		}
		err := r.Create(ctx, item)
		if err != nil {
			return errors.Wrapf(err, "while creating Document with ID %s", item.ID)
		}
	}

	return nil
}

func (r *repository) Delete(ctx context.Context, tenant, id string) error {
	return r.deleter.DeleteOne(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *repository) ListForBundle(ctx context.Context, tenantID string, bundleID string, pageSize int, cursor string) (*model.DocumentPage, error) {
	conditions := repo.Conditions{
		repo.NewEqualCondition("bundle_id", bundleID),
	}
	return r.list(ctx, tenantID, pageSize, cursor, conditions)
}

func (r *repository) list(ctx context.Context, tenant string, pageSize int, cursor string, conditions repo.Conditions) (*model.DocumentPage, error) {
	var documentCollection Collection
	page, totalCount, err := r.pageableQuerier.List(ctx, tenant, pageSize, cursor, "id", &documentCollection, conditions...)
	if err != nil {
		return nil, err
	}

	var items []*model.Document

	for _, documentEnt := range documentCollection {
		m, err := r.conv.FromEntity(documentEnt)
		if err != nil {
			return nil, errors.Wrap(err, "while creating APIDefinition model from entity")
		}
		items = append(items, &m)
	}

	return &model.DocumentPage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page,
	}, nil
}
