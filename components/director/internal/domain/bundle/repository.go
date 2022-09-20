package bundle

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"
)

const (
	bundleTable    string = `public.bundles`
	correlationIDs string = "correlation_ids"
	appIDColumn    string = "app_id"
)

var (
	bundleColumns    = []string{"id", appIDColumn, "name", "description", "instance_auth_request_json_schema", "default_instance_auth", "ord_id", "short_description", "links", "labels", "credential_exchange_strategies", "ready", "created_at", "updated_at", "deleted_at", "error", correlationIDs, "documentation_labels"}
	updatableColumns = []string{"name", "description", "instance_auth_request_json_schema", "default_instance_auth", "ord_id", "short_description", "links", "labels", "credential_exchange_strategies", "ready", "created_at", "updated_at", "deleted_at", "error", "correlation_ids", "documentation_labels"}
	orderByColumns   = repo.OrderByParams{repo.NewAscOrderBy(appIDColumn), repo.NewAscOrderBy("id")}
)

// EntityConverter missing godoc
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityConverter interface {
	ToEntity(in *model.Bundle) (*Entity, error)
	FromEntity(entity *Entity) (*model.Bundle, error)
}

type pgRepository struct {
	existQuerier       repo.ExistQuerier
	singleGetter       repo.SingleGetter
	singleGlobalGetter repo.SingleGetterGlobal
	deleter            repo.Deleter
	lister             repo.Lister
	globalLister       repo.ListerGlobal
	unionLister        repo.UnionLister
	creator            repo.Creator
	updater            repo.Updater
	conv               EntityConverter
}

// NewRepository missing godoc
func NewRepository(conv EntityConverter) *pgRepository {
	return &pgRepository{
		existQuerier:       repo.NewExistQuerier(bundleTable),
		singleGetter:       repo.NewSingleGetter(bundleTable, bundleColumns),
		singleGlobalGetter: repo.NewSingleGetterGlobal(resource.Bundle, bundleTable, bundleColumns),
		deleter:            repo.NewDeleter(bundleTable),
		lister:             repo.NewLister(bundleTable, bundleColumns),
		globalLister:       repo.NewListerGlobal(resource.Bundle, bundleTable, bundleColumns),
		unionLister:        repo.NewUnionLister(bundleTable, bundleColumns),
		creator:            repo.NewCreator(bundleTable, bundleColumns),
		updater:            repo.NewUpdater(bundleTable, updatableColumns, []string{"id"}),
		conv:               conv,
	}
}

// BundleCollection missing godoc
type BundleCollection []Entity

// Len missing godoc
func (r BundleCollection) Len() int {
	return len(r)
}

// Create missing godoc
func (r *pgRepository) Create(ctx context.Context, tenant string, model *model.Bundle) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}

	bndlEnt, err := r.conv.ToEntity(model)
	if err != nil {
		return errors.Wrap(err, "while converting to Bundle entity")
	}

	log.C(ctx).Debugf("Persisting Bundle entity with id %s to db", model.ID)
	return r.creator.Create(ctx, resource.Bundle, tenant, bndlEnt)
}

// Update missing godoc
func (r *pgRepository) Update(ctx context.Context, tenant string, model *model.Bundle) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}

	bndlEnt, err := r.conv.ToEntity(model)

	if err != nil {
		return errors.Wrap(err, "while converting to Bundle entity")
	}

	return r.updater.UpdateSingle(ctx, resource.Bundle, tenant, bndlEnt)
}

// Delete missing godoc
func (r *pgRepository) Delete(ctx context.Context, tenant, id string) error {
	return r.deleter.DeleteOne(ctx, resource.Bundle, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// Exists missing godoc
func (r *pgRepository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, resource.Bundle, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// GetByID missing godoc
func (r *pgRepository) GetByID(ctx context.Context, tenant, id string) (*model.Bundle, error) {
	var bndlEnt Entity
	if err := r.singleGetter.Get(ctx, resource.Bundle, tenant, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &bndlEnt); err != nil {
		return nil, err
	}

	return convertToBundle(r, &bndlEnt)
}

func convertToBundle(r *pgRepository, bndlEnt *Entity) (*model.Bundle, error) {
	bndlModel, err := r.conv.FromEntity(bndlEnt)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Bundle from Entity")
	}

	return bndlModel, nil
}

// GetForApplication missing godoc
func (r *pgRepository) GetForApplication(ctx context.Context, tenant string, id string, applicationID string) (*model.Bundle, error) {
	var ent Entity

	conditions := repo.Conditions{
		repo.NewEqualCondition("id", id),
		repo.NewEqualCondition(appIDColumn, applicationID),
	}
	if err := r.singleGetter.Get(ctx, resource.Bundle, tenant, conditions, repo.NoOrderBy, &ent); err != nil {
		return nil, err
	}

	bndlModel, err := r.conv.FromEntity(&ent)
	if err != nil {
		return nil, errors.Wrap(err, "while creating Bundle model from entity")
	}

	return bndlModel, nil
}

// ListByApplicationIDs missing godoc
func (r *pgRepository) ListByApplicationIDs(ctx context.Context, tenantID string, applicationIDs []string, pageSize int, cursor string) ([]*model.BundlePage, error) {
	var bundleCollection BundleCollection
	counts, err := r.unionLister.List(ctx, resource.Bundle, tenantID, applicationIDs, appIDColumn, pageSize, cursor, orderByColumns, &bundleCollection)
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

	bundlePages := make([]*model.BundlePage, 0, len(applicationIDs))
	for _, appID := range applicationIDs {
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

		bundlePages = append(bundlePages, &model.BundlePage{Data: bundleByID[appID], TotalCount: totalCount, PageInfo: page})
	}

	return bundlePages, nil
}

// ListByApplicationIDNoPaging missing godoc
func (r *pgRepository) ListByApplicationIDNoPaging(ctx context.Context, tenantID, appID string) ([]*model.Bundle, error) {
	bundleCollection := BundleCollection{}
	if err := r.lister.ListWithSelectForUpdate(ctx, resource.Bundle, tenantID, &bundleCollection, repo.NewEqualCondition(appIDColumn, appID)); err != nil {
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

func (r *pgRepository) ListByDestination(ctx context.Context, tenantID string, destination model.DestinationInput) ([]*model.Bundle, error) {
	bundleCollection := BundleCollection{}

	var appIDInCondition repo.Condition
	if destination.XSystemTenantID == "" {
		appIDInCondition = repo.NewInConditionForSubQuery(appIDColumn, `
			SELECT id
			FROM public.applications
			WHERE id IN (
				SELECT id
				FROM tenant_applications
				WHERE tenant_id=(SELECT parent FROM business_tenant_mappings WHERE id = ? )
			)
			AND name = ? AND base_url = ?
		`, []interface{}{tenantID, destination.XSystemTenantName, destination.XSystemBaseURL})
	} else {
		appIDInCondition = repo.NewInConditionForSubQuery(appIDColumn, `
			SELECT DISTINCT pa.id as id
			FROM public.applications pa JOIN labels l ON pa.id=l.app_id
			WHERE pa.id IN (
				SELECT id
				FROM tenant_applications
				WHERE tenant_id=(SELECT parent FROM business_tenant_mappings WHERE id = ? )
			)
			AND l.key='applicationType'
			AND l.value ?| array[?]
			AND pa.local_tenant_id = ?
		`, []interface{}{tenantID, destination.XSystemType, destination.XSystemTenantID})
	}

	conditions := repo.Conditions{
		appIDInCondition,
		repo.NewJSONArrMatchAnyStringCondition(correlationIDs, destination.XCorrelationID),
	}
	err := r.globalLister.ListGlobal(ctx, &bundleCollection, conditions...)
	if err != nil {
		return nil, err
	}
	bundles := make([]*model.Bundle, 0, bundleCollection.Len())
	for _, bundle := range bundleCollection {
		bundleModel, err := r.conv.FromEntity(&bundle)
		if err != nil {
			return nil, errors.Wrap(err, "while creating Bundle model from entity")
		}
		bundles = append(bundles, bundleModel)
	}
	return bundles, nil
}
