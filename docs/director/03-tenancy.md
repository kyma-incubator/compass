The Compass Director is a multi tenant service. Tenant is an object which owns the resources. The tenant has its own entity stored in the database.

# Tenancy
Tenancy in Director is implemented on the database level. Every object which belongs to a tenant has a column `tenant_id` which points to the actual tenant entity.
Tenant is mainly described by two properties: 
* the global tenant identifier (can be any string, treated like identifier from external system) 
* the internal tenant identifier which is used as internal technical identifier (UUID) to which columns `tenant_id` references to.
Those properties are stored in table `business_tenant_mapping` with metadata.
We introduced the internal tenant identifier to implement unified tenant identification in the Director. Thanks to this approach external system can describe the tenant in their own way without any impact on the Director internals.

Compass Director has 3 main objects:
* Application
* Runtime
* Integration System

Application and Runtimes and theirs child resources like APIs and etc are bound to tenants.
Integration Systems are not bound to a tenant and can work representing multiple tenants.

# Importing Tenants
The Compass Director has two components which are responsible for importing tenants:
* Tenant importer - is one time job for importing tenants from files at the first instalation of Compass.
Technical details can be found [here](../../components/director/cmd/tenantloader/README.md).
* Tenant fetcher - is a periodic job which synchronize tenants from external system.
Technical details can be found [here](../../components/director/cmd/tenantfetcher/README.md).

# Authentication flow
Tenant information takes important role during the authentication and authorization phase. The tenant mapping phase is done on the incomming request that is routed to the component called tenant mapping handler. The tenant mapping handler is a HTTP endpoint configured as the Oathkeeper [hydrator](https://www.ory.sh/oathkeeper/docs/pipeline/mutator/#hydrator). On of its responsibility is to map external tenant identifier to the internal one.

More technical details can be found [here](../../components/director/internal/).

# Tenants Query in GraphQL API
The Compass Director GraphQL API exposes `tenants` query. The query return list of all tenants with their external identifier, internal identifier, and additional metadata. 
