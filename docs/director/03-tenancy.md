The Compass Director is a multitenancy service.
Tenant is an object which is owner of resources.

# Tenancy
Tenancy in Director is implemented on database level.
Every object which belongs to tenant has a special column `tenant_id`.
Tenant is mainly described by two properties: 
* the global tenant (can be any string, treated like identifier from external system) 
* the internal tenant which is used as internal technical identifier (UUID) to which columns `tenant_id` has references to.
Those properties are stored in table `business_tenant_mapping` with metadata.
Internal tenant as a way to implement and unify tenancy in Compass Director.

Application and Runtimes are bound to tenants.
Integration System is not bound to any tenant and can work within multiple tenants.

# Tenants seeding
The Compass Director has two components which are responsible for seeding database with tenants:
* Tenant importer - is one time job for importing tenants from files at the first instalation of Compass.
Technical details can be found [here](../../components/director/cmd/tenantloader/README.md).
* Tenant fetcher - is a periodic job which synchronize global tenants from external system.
Technical details can be found [here](../../components/director/cmd/tenantfetcher/README.md).

# Authentication flow
Director has endpoint called `tenant-mapping-service`, which is a [hydrator](https://www.ory.sh/oathkeeper/docs/pipeline/mutator/#hydrator).
The responsibility of this component is to map external tenant to internal tenant and enrich the request with mapping result.
More technical details can be found [here](../../components/director/internal/).

# Tenants Query in GraphQL API
The Compass Director GraphQL API exposes query`Tenants` which return list of all tenants with global tenant, internal tenant and metadata available in the Compass Director. 
