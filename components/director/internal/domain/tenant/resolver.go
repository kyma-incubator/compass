package tenant

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/pkg/errors"
)

// BusinessTenantMappingService is responsible for the service-layer tenant operations.
//go:generate mockery --name=BusinessTenantMappingService --output=automock --outpkg=automock --case=underscore
type BusinessTenantMappingService interface {
	List(ctx context.Context) ([]*model.BusinessTenantMapping, error)
	ListPageBySearchTerm(ctx context.Context, searchTerm string, pageSize int, cursor string) (*model.BusinessTenantMappingPage, error)
	ListLabels(ctx context.Context, tenantID string) (map[string]*model.Label, error)
	GetTenantByExternalID(ctx context.Context, externalID string) (*model.BusinessTenantMapping, error)
	GetTenantByID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
	UpsertMany(ctx context.Context, tenantInputs ...model.BusinessTenantMappingInput) error
	Update(ctx context.Context, id string, tenantInput model.BusinessTenantMappingInput) error
	DeleteMany(ctx context.Context, tenantInputs []string) error
	GetLowestOwnerForResource(ctx context.Context, resourceType resource.Type, objectID string) (string, error)
}

// BusinessTenantMappingConverter is used to convert the internally used tenant representation model.BusinessTenantMapping
// into the external GraphQL representation graphql.Tenant.
//go:generate mockery --name=BusinessTenantMappingConverter --output=automock --outpkg=automock --case=underscore
type BusinessTenantMappingConverter interface {
	MultipleToGraphQL(in []*model.BusinessTenantMapping) []*graphql.Tenant
	MultipleInputFromGraphQL(in []*graphql.BusinessTenantMappingInput) []model.BusinessTenantMappingInput
	ToGraphQL(in *model.BusinessTenantMapping) *graphql.Tenant
}

// TenantFetcher calls an API which fetches details for the given tenant from an external tenancy service, stores the tenant in the Compass DB and returns 200 OK if the tenant was successfully created.
//go:generate mockery --name=TenantFetcher --output=automock --outpkg=automock --case=underscore
type TenantFetcher interface {
	FetchOnDemand(tenant string) error
}

// Resolver is the resolver responsible for tenant-related GraphQL requests.
type Resolver struct {
	transact persistence.Transactioner

	srv     BusinessTenantMappingService
	conv    BusinessTenantMappingConverter
	fetcher TenantFetcher
}

// NewResolver returns the GraphQL resolver for tenants.
func NewResolver(transact persistence.Transactioner, srv BusinessTenantMappingService, conv BusinessTenantMappingConverter, fetcher TenantFetcher) *Resolver {
	return &Resolver{
		transact: transact,
		srv:      srv,
		conv:     conv,
		fetcher:  fetcher,
	}
}

// Tenants transactionally retrieves a page of tenants present in the Compass storage by a search term. If the search term is missing it will be ignored in the resulting tenant subset.
func (r *Resolver) Tenants(ctx context.Context, first *int, after *graphql.PageCursor, searchTerm *string) (*graphql.TenantPage, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	var cursor string
	if after != nil {
		cursor = string(*after)
	}
	if first == nil {
		return nil, apperrors.NewInvalidDataError("missing required parameter 'first'")
	}

	searchStr := str.PtrStrToStr(searchTerm)

	ctx = persistence.SaveToContext(ctx, tx)

	tenantsPage, err := r.srv.ListPageBySearchTerm(ctx, searchStr, *first, cursor)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	gqlTenants := r.conv.MultipleToGraphQL(tenantsPage.Data)

	return &graphql.TenantPage{
		Data:       gqlTenants,
		TotalCount: tenantsPage.TotalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(tenantsPage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(tenantsPage.PageInfo.EndCursor),
			HasNextPage: tenantsPage.PageInfo.HasNextPage,
		},
	}, nil
}

// Tenant retrieves a tenant with the provided external ID from the Compass storage.
func (r *Resolver) Tenant(ctx context.Context, externalID string) (*graphql.Tenant, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	tenant, err := r.srv.GetTenantByExternalID(ctx, externalID)
	if err != nil && apperrors.IsNotFoundError(err) {
		tx, err = r.fetchTenant(&tx, externalID)
	}
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	gqlTenant := r.conv.MultipleToGraphQL([]*model.BusinessTenantMapping{tenant})
	return gqlTenant[0], nil
}

// TenantByID retrieves a tenant with the provided internal ID from the Compass storage.
func (r *Resolver) TenantByID(ctx context.Context, internalID string) (*graphql.Tenant, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	t, err := r.srv.GetTenantByID(ctx, internalID)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	gqlTenant := r.conv.ToGraphQL(t)
	return gqlTenant, nil
}

// TenantByLowestOwnerForResource retrieves a tenant with the provided internal ID from the Compass storage.
func (r *Resolver) TenantByLowestOwnerForResource(ctx context.Context, resourceStr, objectID string) (string, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return "", err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	resourceType := resource.Type(resourceStr)

	tenantID, err := r.srv.GetLowestOwnerForResource(ctx, resourceType, objectID)
	if err != nil {
		return "", err
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}

	return tenantID, nil
}

// Labels transactionally retrieves all existing labels of the given tenant if it exists.
func (r *Resolver) Labels(ctx context.Context, obj *graphql.Tenant, key *string) (graphql.Labels, error) {
	if obj == nil {
		return nil, apperrors.NewInternalError("Tenant cannot be empty")
	}
	log.C(ctx).Infof("getting labels for tenant with ID %s, and internal ID %s", obj.ID, obj.InternalID)

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	itemMap, err := r.srv.ListLabels(ctx, obj.InternalID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, tx.Commit()
		}
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	resultLabels := make(map[string]interface{})
	for _, label := range itemMap {
		if key == nil || label.Key == *key {
			resultLabels[label.Key] = label.Value
		}
	}

	return resultLabels, nil
}

// Write creates new global and subaccounts
func (r *Resolver) Write(ctx context.Context, inputTenants []*graphql.BusinessTenantMappingInput) (int, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return -1, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	tenants := r.conv.MultipleInputFromGraphQL(inputTenants)

	if err := r.srv.UpsertMany(ctx, tenants...); err != nil {
		return -1, errors.Wrap(err, "while writing new tenants")
	}

	if err = tx.Commit(); err != nil {
		return -1, err
	}

	return len(tenants), nil
}

// Delete deletes tenants
func (r *Resolver) Delete(ctx context.Context, externalTenantIDs []string) (int, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return -1, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err := r.srv.DeleteMany(ctx, externalTenantIDs); err != nil {
		return -1, errors.Wrap(err, "while deleting tenants")
	}

	if err = tx.Commit(); err != nil {
		return -1, err
	}

	return len(externalTenantIDs), nil
}

// Update update single tenant
func (r *Resolver) Update(ctx context.Context, id string, in graphql.BusinessTenantMappingInput) (*graphql.Tenant, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)
	tenantModels := r.conv.MultipleInputFromGraphQL([]*graphql.BusinessTenantMappingInput{&in})
	if err := r.srv.Update(ctx, id, tenantModels[0]); err != nil {
		return nil, errors.Wrap(err, "while deleting tenants")
	}

	tenant, err := r.srv.GetTenantByExternalID(ctx, in.ExternalTenant)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting tenant with external id %s", in.ExternalTenant)
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(tenant), nil
}

func (r *Resolver) fetchTenant(tx *persistence.PersistenceTx, externalID string) (persistence.PersistenceTx, error) {
	if err := (*tx).Commit(); err != nil {
		return nil, err
	}
	if err := r.fetcher.FetchOnDemand(externalID); err != nil {
		return nil, errors.Wrapf(err, "while trying to create if not exists tenant %s", externalID)
	}
	tr, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	return tr, nil
}
