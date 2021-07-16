package ownertenant

import (
	"context"

	gqlgen "github.com/99designs/gqlgen/graphql"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

const MutationObject = "Mutation"

//go:generate mockery --name=TenantIndexRepository --output=automock --outpkg=automock --case=underscore
type TenantIndexRepository interface {
	GetOwnerTenantByResourceID(ctx context.Context, callingTenant, resourceId string) (string, error)
}

//go:generate mockery --name=TenantRepository --output=automock --outpkg=automock --case=underscore
type TenantRepository interface {
	Get(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
}

type middleware struct {
	transact        persistence.Transactioner
	tenantIndexRepo TenantIndexRepository
	tenantRepo      TenantRepository
}

// NewMiddleware creates a new handler struct responsible for deducing the owner tenant of the requested resource
func NewMiddleware(transact persistence.Transactioner, tenantIndexRepo TenantIndexRepository, tenantRepo TenantRepository) *middleware {
	return &middleware{
		transact:        transact,
		tenantIndexRepo: tenantIndexRepo,
		tenantRepo:      tenantRepo,
	}
}

// ExtensionName should be a CamelCase string version of the extension which may be shown in stats and logging.
func (m *middleware) ExtensionName() string {
	return "OwnerTenantExtension"
}

// Validate is called when adding an extension to the server, it allows validation against the servers schema.
func (m *middleware) Validate(_ gqlgen.ExecutableSchema) error {
	return nil
}

// InterceptField intercepts mutation fields in order to find the right tenant to use for the mutation.
//
// Since business tenant mappings can contain tenants and customers, one customer can has multiple child tenants.
// Furthermore the customer can access all the information across his tenants.
//
// However there are some mutations (for example addBundle to application) that requires the bundle to be inserted in the same
// tenant as the owning application. Therefore we should use the application owner's tenant and not the request caller tenant.
//
// The caller in this case is the customer who has access to the application, but when he creates a bundle the bundle should be
// added to the child tenant which owns the application. This is valid for all other mutations for child entities as well.
//
// The validation that the customer has access to the entity as well as the child tenant owning it  is baked in the GetOwnerTenantByResourceID method.
// If the customer has no access to the tenant/resource he will not find the ownerTenant and
// the next resolver in the chain will fail with tenant isolation error.
func (m *middleware) InterceptField(ctx context.Context, next gqlgen.Resolver) (res interface{}, err error) {
	fieldCtx := gqlgen.GetFieldContext(ctx)

	if fieldCtx == nil || fieldCtx.Object != MutationObject {
		return next(ctx)
	}

	var id string
	for _, arg := range fieldCtx.Field.Field.Arguments {
		if arg.Value.Definition.Name == "ID" && arg.Value.Definition.BuiltIn == true { // The argument is of graphql's built-in ID type
			if len(arg.Value.Raw) > 0 { // The ID argument is populated
				if len(id) > 0 { // We already found an ID argument that is populated. More than one ID arguments provided is not supported for now.
					return nil, errors.New("More than one argument with type ID is provided for the mutation")
				}
				id = arg.Value.Raw
			}
		}
	}

	if len(id) > 0 { // If the mutation has an ID argument - potentially parent entity
		tnt, err := tenant.LoadFromContext(ctx)
		if err != nil {
			log.C(ctx).Infof("Mutation call does not have a valid tenant in the context: %s. Proceeding without modifications...", err)
			return next(ctx)
		}

		tx, err := m.transact.Begin()
		if err != nil {
			log.C(ctx).WithError(err).Errorf("An error has occurred while opening transaction: %v", err)
			return nil, err
		}
		defer m.transact.RollbackUnlessCommitted(ctx, tx)

		ctxWithDB := persistence.SaveToContext(ctx, tx)

		ownerTenant, err := m.tenantIndexRepo.GetOwnerTenantByResourceID(ctxWithDB, tnt, id)
		if err != nil {
			if apperrors.IsNotFoundError(err) { // A tenant_id of the resource is not found in the index - potentially a global resource or the tenant has no access to it -> do nothing and let the resolver decide
				if err := tx.Commit(); err != nil {
					log.C(ctx).WithError(err).Errorf("An error has occurred while committing transaction: %v", err)
					return nil, err
				}
				return next(ctx)
			}
			return nil, err
		}

		if ownerTenant != tnt {
			// Now we have the ownerTenant of the resource and we know that the calling tenant is different and has access to that tenant (it is his parent).
			// Proceeding with the ownerTenant of the resource in the context.
			tenantMapping, err := m.tenantRepo.Get(ctxWithDB, ownerTenant)
			if err != nil {
				return nil, err
			}

			log.C(ctx).Infof("Mutation call by tenant %s for resource with id %s owned by child tenant %s. Proceeding with the owner tenant...", tnt, id, ownerTenant)

			ctx = tenant.SaveToContext(ctx, tenantMapping.ID, tenantMapping.ExternalTenant)
		}

		if err := tx.Commit(); err != nil {
			log.C(ctx).WithError(err).Errorf("An error has occurred while committing transaction: %v", err)
			return nil, err
		}
	}

	return next(ctx)
}
