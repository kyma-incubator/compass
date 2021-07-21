package mp_bundle

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"


	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"
)

const bundleTable string = `public.bundles`
const bundleInstanceAuthTable string = `public.bundle_instance_auths`
const bundleInstanceAuthBundleRefField string = `bundle_id`

var (
	bundleColumns    = []string{"id", "tenant_id", "app_id", "name", "description", "instance_auth_request_json_schema", "default_instance_auth", "ord_id", "short_description", "links", "labels", "credential_exchange_strategies", "ready", "created_at", "updated_at", "deleted_at", "error"}
	tenantColumn     = "tenant_id"
	updatableColumns = []string{"name", "description", "instance_auth_request_json_schema", "default_instance_auth", "ord_id", "short_description", "links", "labels", "credential_exchange_strategies", "ready", "created_at", "updated_at", "deleted_at", "error"}
	orderByColumns = repo.OrderByParams{repo.NewAscOrderBy("app_id"), repo.NewAscOrderBy("id")}
)

//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore
type EntityConverter interface {
	ToEntity(in *model.Bundle) (*Entity, error)
	FromEntity(entity *Entity) (*model.Bundle, error)
}

type pgRepository struct {
	existQuerier    repo.ExistQuerier
	singleGetter    repo.SingleGetter
	deleter         repo.Deleter
	pageableQuerier repo.PageableQuerier
	lister          repo.Lister
	unionLister     repo.UnionLister
	creator         repo.Creator
	updater         repo.Updater
	conv            EntityConverter
}

func NewRepository(conv EntityConverter) *pgRepository {
	return &pgRepository{
		existQuerier:    repo.NewExistQuerier(resource.Bundle, bundleTable, tenantColumn),
		singleGetter:    repo.NewSingleGetter(resource.Bundle, bundleTable, tenantColumn, bundleColumns),
		deleter:         repo.NewDeleter(resource.Bundle, bundleTable, tenantColumn),
		pageableQuerier: repo.NewPageableQuerier(resource.Bundle, bundleTable, tenantColumn, bundleColumns),
		lister:          repo.NewLister(resource.Bundle, bundleTable, tenantColumn, bundleColumns),
		unionLister:     repo.NewUnionLister(resource.Bundle, bundleTable, tenantColumn, bundleColumns),
		creator:         repo.NewCreator(resource.Bundle, bundleTable, bundleColumns),
		updater:         repo.NewUpdater(resource.Bundle, bundleTable, updatableColumns, tenantColumn, []string{"id"}),
		conv:            conv,
	}
}

type BundleCollection []Entity

func (r BundleCollection) Len() int {
	return len(r)
}

func (r *pgRepository) Create(ctx context.Context, model *model.Bundle) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}

	bndlEnt, err := r.conv.ToEntity(model)
	if err != nil {
		return errors.Wrap(err, "while converting to Bundle entity")
	}

	log.C(ctx).Debugf("Persisting Bundle entity with id %s to db", model.ID)
	return r.creator.Create(ctx, bndlEnt)
}

func (r *pgRepository) Update(ctx context.Context, model *model.Bundle) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}

	bndlEnt, err := r.conv.ToEntity(model)

	if err != nil {
		return errors.Wrap(err, "while converting to Bundle entity")
	}

	return r.updater.UpdateSingle(ctx, bndlEnt)
}

func (r *pgRepository) Delete(ctx context.Context, tenant, id string) error {
	return r.deleter.DeleteOne(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *pgRepository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *pgRepository) GetByID(ctx context.Context, tenant, id string) (*model.Bundle, error) {
	var bndlEnt Entity
	if err := r.singleGetter.Get(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &bndlEnt); err != nil {
		return nil, err
	}

	bndlModel, err := r.conv.FromEntity(&bndlEnt)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Bundle from Entity")
	}

	return bndlModel, nil
}

func (r *pgRepository) GetForApplication(ctx context.Context, tenant string, id string, applicationID string) (*model.Bundle, error) {
	var ent Entity

	conditions := repo.Conditions{
		repo.NewEqualCondition("id", id),
		repo.NewEqualCondition("app_id", applicationID),
	}
	if err := r.singleGetter.Get(ctx, tenant, conditions, repo.NoOrderBy, &ent); err != nil {
		return nil, err
	}

	bndlModel, err := r.conv.FromEntity(&ent)
	if err != nil {
		return nil, errors.Wrap(err, "while creating Bundle model from entity")
	}

	return bndlModel, nil
}

func (r *pgRepository) ListByApplicationID(ctx context.Context, tenantID string, applicationID string, pageSize int, cursor string) (*model.BundlePage, error) {
	conditions := repo.Conditions{
		repo.NewEqualCondition("app_id", applicationID),
	}

	var bundleCollection BundleCollection
	page, totalCount, err := r.pageableQuerier.List(ctx, tenantID, pageSize, cursor, "id", &bundleCollection, conditions...)
	if err != nil {
		return nil, err
	}

	var items []*model.Bundle

	for _, bndlEnt := range bundleCollection {
		m, err := r.conv.FromEntity(&bndlEnt)
		if err != nil {
			return nil, errors.Wrap(err, "while creating Bundle model from entity")
		}
		items = append(items, m)
	}

	return &model.BundlePage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page,
	}, nil
}

func (r *pgRepository) ListByApplicationIDs(ctx context.Context, tenantID string, applicationIDs []string, pageSize int, cursor string) ([]*model.BundlePage, error) {
	var bundleCollection BundleCollection
	counts, err := r.unionLister.List(ctx, tenantID, applicationIDs, "app_id", pageSize, cursor, orderByColumns, &bundleCollection)
	if err != nil {
		return nil, err
	}

	bundleByID := map[string][]*model.Bundle{}
	for _, bundleEnt := range bundleCollection {
		m, err := r.conv.FromEntity(&bundleEnt)
		if err != nil {
			return nil, errors.Wrap(err, "while creating Bundle model from entity")
		}
		bundleByID[bundleEnt.ApplicationID] = append(bundleByID[bundleEnt.ApplicationID], m)
	}

	offset, err := pagination.DecodeOffsetCursor(cursor)
	if err != nil {
		return nil, errors.Wrap(err, "while decoding page cursor")
	}

	bundlePages := make([]*model.BundlePage, len(applicationIDs))
	for i, appID := range applicationIDs {
		totalCount := counts[appID]
		hasNextPage := false
		endCursor := ""
		if totalCount > offset+len(bundleByID[appID]) {
			hasNextPage = true
			endCursor = pagination.EncodeNextOffsetCursor(offset, pageSize)
		}

		page := &pagination.Page{
			StartCursor: cursor,
			EndCursor:   endCursor,
			HasNextPage: hasNextPage,
		}

		bundlePages[i] = &model.BundlePage{Data: bundleByID[appID], TotalCount: totalCount, PageInfo: page}
	}

	return bundlePages, nil
}

func (r *pgRepository) ListByApplicationIDNoPaging(ctx context.Context, tenantID, appID string) ([]*model.Bundle, error) {
	bundleCollection := BundleCollection{}
	if err := r.lister.List(ctx, tenantID, &bundleCollection, repo.NewEqualCondition("app_id", appID)); err != nil {
		return nil, err
	}
	bundles := make([]*model.Bundle, 0, bundleCollection.Len())
	for _, bundle := range bundleCollection {
		bundleModel, err := r.conv.FromEntity(&bundle)
		if err != nil {
			return nil, err
		}
		bundles = append(bundles, bundleModel)
	}
	return bundles, nil
}
