package tenant

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/pkg/errors"
)

// BusinessTenantMappingService is responsible for the service-layer tenant operations.
//go:generate mockery --name=BusinessTenantMappingService --output=automock --outpkg=automock --case=underscore
type BusinessTenantMappingService interface {
	List(ctx context.Context) ([]*model.BusinessTenantMapping, error)
	ListLabels(ctx context.Context, tenantID string) (map[string]*model.Label, error)
	GetTenantByExternalID(ctx context.Context, externalID string) (*model.BusinessTenantMapping, error)
	UpsertMany(ctx context.Context, tenantInputs ...model.BusinessTenantMappingInput) error
	Update(ctx context.Context, id string, tenantInput model.BusinessTenantMappingInput) error
	DeleteMany(ctx context.Context, tenantInputs []string) error
}

// BusinessTenantMappingConverter is used to convert the internally used tenant representation model.BusinessTenantMapping
// into the external GraphQL representation graphql.Tenant.
//go:generate mockery --name=BusinessTenantMappingConverter --output=automock --outpkg=automock --case=underscore
type BusinessTenantMappingConverter interface {
	MultipleToGraphQL(in []*model.BusinessTenantMapping) []*graphql.Tenant
	MultipleInputFromGraphQL(in []*graphql.BusinessTenantMappingInput) []model.BusinessTenantMappingInput
	ToGraphQL(in *model.BusinessTenantMapping) *graphql.Tenant
}

// Resolver is the resolver responsible for tenant-related GraphQL requests.
type Resolver struct {
	transact persistence.Transactioner

	srv  BusinessTenantMappingService
	conv BusinessTenantMappingConverter
}

// Tenants transactionally retrieves all tenants present in the Compass storage.
func (r *Resolver) Tenants(ctx context.Context) ([]*graphql.Tenant, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	tenants, err := r.srv.List(ctx)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	gqlTenants := r.conv.MultipleToGraphQL(tenants)
	return gqlTenants, nil
}

// Tenant retrieves a tenant with the provided external ID from the Compass storage.
func (r *Resolver) Tenant(ctx context.Context, externalID string) (*graphql.Tenant, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	t, err := r.srv.GetTenantByExternalID(ctx, externalID)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	gqlTenant := r.conv.MultipleToGraphQL([]*model.BusinessTenantMapping{t})
	return gqlTenant[0], nil
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

// Write creates new tenants
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

// NewResolver returns the GraphQL resolver for tenants.
func NewResolver(transact persistence.Transactioner, srv BusinessTenantMappingService, conv BusinessTenantMappingConverter) *Resolver {
	return &Resolver{
		transact: transact,
		srv:      srv,
		conv:     conv,
	}
}
