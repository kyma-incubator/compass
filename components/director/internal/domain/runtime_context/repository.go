package runtimectx

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const runtimeContextsTable string = `public.runtime_contexts`

var (
	runtimeContextColumns = []string{"id", "runtime_id", "key", "value"}
	updatableColumns      = []string{"key", "value"}
	orderByColumns        = repo.OrderByParams{repo.NewAscOrderBy("runtime_id"), repo.NewAscOrderBy("id")}
)

type pgRepository struct {
	existQuerier       repo.ExistQuerier
	singleGetter       repo.SingleGetter
	singleGetterGlobal repo.SingleGetterGlobal
	deleter            repo.Deleter
	pageableQuerier    repo.PageableQuerier
	unionLister        repo.UnionLister
	creator            repo.Creator
	updater            repo.Updater
	conv               entityConverter
}

//go:generate mockery --exported --name=entityConverter --output=automock --outpkg=automock --case=underscore
type entityConverter interface {
	ToEntity(in *model.RuntimeContext) *RuntimeContext
	FromEntity(entity *RuntimeContext) *model.RuntimeContext
}

// NewRepository missing godoc
func NewRepository(conv entityConverter) *pgRepository {
	return &pgRepository{
		existQuerier:       repo.NewExistQuerier(runtimeContextsTable),
		singleGetter:       repo.NewSingleGetter(runtimeContextsTable, runtimeContextColumns),
		singleGetterGlobal: repo.NewSingleGetterGlobal(resource.RuntimeContext, runtimeContextsTable, runtimeContextColumns),
		deleter:            repo.NewDeleter(runtimeContextsTable),
		pageableQuerier:    repo.NewPageableQuerier(runtimeContextsTable, runtimeContextColumns),
		unionLister:        repo.NewUnionLister(runtimeContextsTable, runtimeContextColumns),
		creator:            repo.NewCreator(runtimeContextsTable, runtimeContextColumns),
		updater:            repo.NewUpdater(runtimeContextsTable, updatableColumns, []string{"id"}),
		conv:               conv,
	}
}

// Exists returns true if a RuntimeContext with the provided `id` exists in the database and is visible for `tenant`
func (r *pgRepository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, resource.RuntimeContext, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// Delete deletes the RuntimeContext with the provided `id` from the database if `tenant` has the appropriate access to it
func (r *pgRepository) Delete(ctx context.Context, tenant string, id string) error {
	return r.deleter.DeleteOne(ctx, resource.RuntimeContext, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// GetByID retrieves the RuntimeContext with the provided `id` from the database if it exists and is visible for `tenant`
func (r *pgRepository) GetByID(ctx context.Context, tenant, id string) (*model.RuntimeContext, error) {
	var runtimeCtxEnt RuntimeContext
	if err := r.singleGetter.Get(ctx, resource.RuntimeContext, tenant, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &runtimeCtxEnt); err != nil {
		return nil, err
	}

	return r.conv.FromEntity(&runtimeCtxEnt), nil
}

// GetForRuntime retrieves RuntimeContext with the provided `id` associated to Runtime with id `runtimeID` from the database if it exists and is visible for `tenant`
func (r *pgRepository) GetForRuntime(ctx context.Context, tenant, id, runtimeID string) (*model.RuntimeContext, error) {
	var runtimeCtxEnt RuntimeContext

	conditions := repo.Conditions{
		repo.NewEqualCondition("id", id),
		repo.NewEqualCondition("runtime_id", runtimeID),
	}

	if err := r.singleGetter.Get(ctx, resource.RuntimeContext, tenant, conditions, repo.NoOrderBy, &runtimeCtxEnt); err != nil {
		return nil, err
	}

	return r.conv.FromEntity(&runtimeCtxEnt), nil
}

// GetByFiltersAndID retrieves RuntimeContext with the provided `id` matching the provided filters from the database if it exists and is visible for `tenant`
func (r *pgRepository) GetByFiltersAndID(ctx context.Context, tenant, id string, filter []*labelfilter.LabelFilter) (*model.RuntimeContext, error) {
	tenantID, err := uuid.Parse(tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing tenant as UUID")
	}

	additionalConditions := repo.Conditions{repo.NewEqualCondition("id", id)}

	filterSubquery, args, err := label.FilterQuery(model.RuntimeContextLabelableObject, label.IntersectSet, tenantID, filter)
	if err != nil {
		return nil, errors.Wrap(err, "while building filter query")
	}
	if filterSubquery != "" {
		additionalConditions = append(additionalConditions, repo.NewInConditionForSubQuery("id", filterSubquery, args))
	}

	var runtimeCtxEnt RuntimeContext
	if err := r.singleGetter.Get(ctx, resource.RuntimeContext, tenant, additionalConditions, repo.NoOrderBy, &runtimeCtxEnt); err != nil {
		return nil, err
	}

	return r.conv.FromEntity(&runtimeCtxEnt), nil
}

// GetByFiltersGlobal retrieves RuntimeContext matching the provided filters from the database if it exists
func (r *pgRepository) GetByFiltersGlobal(ctx context.Context, filter []*labelfilter.LabelFilter) (*model.RuntimeContext, error) {
	filterSubquery, args, err := label.FilterQueryGlobal(model.RuntimeContextLabelableObject, label.IntersectSet, filter)
	if err != nil {
		return nil, errors.Wrap(err, "while building filter query")
	}

	var additionalConditions repo.Conditions
	if filterSubquery != "" {
		additionalConditions = append(additionalConditions, repo.NewInConditionForSubQuery("id", filterSubquery, args))
	}

	var runtimeCtxEnt RuntimeContext
	if err := r.singleGetterGlobal.GetGlobal(ctx, additionalConditions, repo.NoOrderBy, &runtimeCtxEnt); err != nil {
		return nil, err
	}

	return r.conv.FromEntity(&runtimeCtxEnt), nil
}

// RuntimeContextCollection represents collection of RuntimeContext
type RuntimeContextCollection []RuntimeContext

// Len returns the count of RuntimeContexts in the collection
func (r RuntimeContextCollection) Len() int {
	return len(r)
}

// List retrieves a page of RuntimeContext objects associated to Runtime with id `runtimeID` that are matching the provided filters from the database that are visible for `tenant`
func (r *pgRepository) List(ctx context.Context, runtimeID string, tenant string, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.RuntimeContextPage, error) {
	var runtimeCtxsCollection RuntimeContextCollection
	tenantID, err := uuid.Parse(tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing tenant as UUID")
	}
	filterSubquery, args, err := label.FilterQuery(model.RuntimeContextLabelableObject, label.IntersectSet, tenantID, filter)
	if err != nil {
		return nil, errors.Wrap(err, "while building filter query")
	}

	conditions := repo.Conditions{
		repo.NewEqualCondition("runtime_id", runtimeID),
	}
	if filterSubquery != "" {
		conditions = append(conditions, repo.NewInConditionForSubQuery("id", filterSubquery, args))
	}

	page, totalCount, err := r.pageableQuerier.List(ctx, resource.RuntimeContext, tenant, pageSize, cursor, "id", &runtimeCtxsCollection, conditions...)

	if err != nil {
		return nil, err
	}

	items := make([]*model.RuntimeContext, 0, len(runtimeCtxsCollection))

	for _, runtimeCtxEnt := range runtimeCtxsCollection {
		items = append(items, r.conv.FromEntity(&runtimeCtxEnt))
	}
	return &model.RuntimeContextPage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page,
	}, nil
}

// ListByRuntimeIDs retrieves a page of RuntimeContext objects for each runtimeID from the database that are visible for `tenantID`
func (r *pgRepository) ListByRuntimeIDs(ctx context.Context, tenantID string, runtimeIDs []string, pageSize int, cursor string) ([]*model.RuntimeContextPage, error) {
	var runtimeCtxsCollection RuntimeContextCollection

	counts, err := r.unionLister.List(ctx, resource.RuntimeContext, tenantID, runtimeIDs, "runtime_id", pageSize, cursor, orderByColumns, &runtimeCtxsCollection)
	if err != nil {
		return nil, err
	}

	runtimeContextByID := map[string][]*model.RuntimeContext{}
	for _, runtimeContextEntity := range runtimeCtxsCollection {
		rc := r.conv.FromEntity(&runtimeContextEntity)
		runtimeContextByID[runtimeContextEntity.RuntimeID] = append(runtimeContextByID[runtimeContextEntity.RuntimeID], rc)
	}

	offset, err := pagination.DecodeOffsetCursor(cursor)
	if err != nil {
		return nil, errors.Wrap(err, "while decoding page cursor")
	}

	runtimeContextPages := make([]*model.RuntimeContextPage, 0, len(runtimeIDs))
	for _, runtimeID := range runtimeIDs {
		totalCount := counts[runtimeID]
		hasNextPage := false
		endCursor := ""
		if totalCount > offset+len(runtimeContextByID[runtimeID]) {
			hasNextPage = true
			endCursor = pagination.EncodeNextOffsetCursor(offset, pageSize)
		}

		page := &pagination.Page{
			StartCursor: cursor,
			EndCursor:   endCursor,
			HasNextPage: hasNextPage,
		}

		runtimeContextPages = append(runtimeContextPages, &model.RuntimeContextPage{Data: runtimeContextByID[runtimeID], TotalCount: totalCount, PageInfo: page})
	}

	return runtimeContextPages, nil
}

// Create stores RuntimeContext entity in the database using the values from `item`
func (r *pgRepository) Create(ctx context.Context, tenant string, item *model.RuntimeContext) error {
	if item == nil {
		return apperrors.NewInternalError("item can not be empty")
	}
	return r.creator.Create(ctx, resource.RuntimeContext, tenant, r.conv.ToEntity(item))
}

// Update updates the existing RuntimeContext entity in the database with the values from `item`
func (r *pgRepository) Update(ctx context.Context, tenant string, item *model.RuntimeContext) error {
	if item == nil {
		return apperrors.NewInternalError("item can not be empty")
	}
	return r.updater.UpdateSingle(ctx, resource.RuntimeContext, tenant, r.conv.ToEntity(item))
}
