package document

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const documentTable = "public.documents"

var (
	documentColumns = []string{"id", "bundle_id", "app_id", "title", "display_name", "description", "format", "kind", "data", "ready", "created_at", "updated_at", "deleted_at", "error"}
	bundleIDColumn  = "bundle_id"
	orderByColumns  = repo.OrderByParams{repo.NewAscOrderBy("bundle_id"), repo.NewAscOrderBy("id")}
)

// Converter missing godoc
//go:generate mockery --name=Converter --output=automock --outpkg=automock --case=underscore --disable-version-string
type Converter interface {
	ToEntity(in *model.Document) (*Entity, error)
	FromEntity(in *Entity) (*model.Document, error)
}

type repository struct {
	existQuerier    repo.ExistQuerier
	singleGetter    repo.SingleGetter
	unionLister     repo.UnionLister
	deleter         repo.Deleter
	pageableQuerier repo.PageableQuerier
	creator         repo.Creator

	conv Converter
}

// NewRepository missing godoc
func NewRepository(conv Converter) *repository {
	return &repository{
		existQuerier:    repo.NewExistQuerier(documentTable),
		singleGetter:    repo.NewSingleGetter(documentTable, documentColumns),
		unionLister:     repo.NewUnionLister(documentTable, documentColumns),
		deleter:         repo.NewDeleter(documentTable),
		pageableQuerier: repo.NewPageableQuerier(documentTable, documentColumns),
		creator:         repo.NewCreator(documentTable, documentColumns),
		conv:            conv,
	}
}

// DocumentCollection missing godoc
type DocumentCollection []Entity

// Len missing godoc
func (d DocumentCollection) Len() int {
	return len(d)
}

// Exists missing godoc
func (r *repository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, resource.Document, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// GetByID missing godoc
func (r *repository) GetByID(ctx context.Context, tenant, id string) (*model.Document, error) {
	var entity Entity
	if err := r.singleGetter.Get(ctx, resource.Document, tenant, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}

	docModel, err := r.conv.FromEntity(&entity)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Document entity to model")
	}

	return docModel, nil
}

// GetForBundle missing godoc
func (r *repository) GetForBundle(ctx context.Context, tenant string, id string, bundleID string) (*model.Document, error) {
	var ent Entity

	conditions := repo.Conditions{
		repo.NewEqualCondition("id", id),
		repo.NewEqualCondition("bundle_id", bundleID),
	}
	if err := r.singleGetter.Get(ctx, resource.Document, tenant, conditions, repo.NoOrderBy, &ent); err != nil {
		return nil, err
	}

	documentModel, err := r.conv.FromEntity(&ent)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Document entity to model")
	}

	return documentModel, nil
}

// Create missing godoc
func (r *repository) Create(ctx context.Context, tenant string, item *model.Document) error {
	if item == nil {
		return apperrors.NewInternalError("Document cannot be empty")
	}

	entity, err := r.conv.ToEntity(item)
	if err != nil {
		return errors.Wrap(err, "while creating Document entity from model")
	}

	log.C(ctx).Debugf("Persisting Document entity with id %s to db", item.ID)
	return r.creator.Create(ctx, resource.Document, tenant, entity)
}

// CreateMany missing godoc
func (r *repository) CreateMany(ctx context.Context, tenant string, items []*model.Document) error {
	for _, item := range items {
		if item == nil {
			return apperrors.NewInternalError("Document cannot be empty")
		}
		err := r.Create(ctx, tenant, item)
		if err != nil {
			return errors.Wrapf(err, "while creating Document with ID %s", item.ID)
		}
	}

	return nil
}

// Delete missing godoc
func (r *repository) Delete(ctx context.Context, tenant, id string) error {
	return r.deleter.DeleteOne(ctx, resource.Document, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// ListByBundleIDs missing godoc
func (r *repository) ListByBundleIDs(ctx context.Context, tenantID string, bundleIDs []string, pageSize int, cursor string) ([]*model.DocumentPage, error) {
	var documentCollection DocumentCollection
	counts, err := r.unionLister.List(ctx, resource.Document, tenantID, bundleIDs, bundleIDColumn, pageSize, cursor, orderByColumns, &documentCollection)
	if err != nil {
		return nil, err
	}

	documentByID := map[string][]*model.Document{}
	for _, documentEnt := range documentCollection {
		m, err := r.conv.FromEntity(&documentEnt)
		if err != nil {
			return nil, errors.Wrap(err, "while creating Document model from entity")
		}
		documentByID[documentEnt.BndlID] = append(documentByID[documentEnt.BndlID], m)
	}

	offset, err := pagination.DecodeOffsetCursor(cursor)
	if err != nil {
		return nil, errors.Wrap(err, "while decoding page cursor")
	}

	documentPages := make([]*model.DocumentPage, 0, len(bundleIDs))
	for _, bndlID := range bundleIDs {
		totalCount := counts[bndlID]
		hasNextPage := false
		endCursor := ""
		if totalCount > offset+len(documentByID[bndlID]) {
			hasNextPage = true
			endCursor = pagination.EncodeNextOffsetCursor(offset, pageSize)
		}

		page := &pagination.Page{
			StartCursor: cursor,
			EndCursor:   endCursor,
			HasNextPage: hasNextPage,
		}

		documentPages = append(documentPages, &model.DocumentPage{Data: documentByID[bndlID], TotalCount: totalCount, PageInfo: page})
	}

	return documentPages, nil
}
