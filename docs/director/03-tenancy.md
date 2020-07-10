#Tenancy

The Compass Director is a multi-tenant service. A tenant is an object that owns resources. The tenant has its own entity stored in the database.

## Tenancy in Director
Tenancy in Director is implemented on the database level. Every object which belongs to a tenant has the `tenant_id` column which points to the actual tenant entity.
A tenant is mainly described by two properties: 
*  Global tenant identifier - can be any string, treated like identifier from an external system 
* Internal tenant identifier - used as an internal technical identifier (UUID) to which `tenant_id` columns refer to
Those properties are stored in the `business_tenant_mapping` table with metadata.
Internal tenant identifier allows for unified tenant identification in Director. Thanks to this approach, external systems can describe their tenants in their own way without any impact on the Director internals.

Compass Director has 3 main objects:
* Application
* Runtime
* Integration System

Applications and Runtimes and their child resources, such as APIs, are bound to tenants.
Integration Systems are not bound to a tenant and can work representing multiple tenants.

## Importing tenants
The Director consists of two components that are responsible for importing tenants:
* Tenant importer - is one time job for importing tenants from files at the first instalation of Compass.
Technical details can be found [here](../../components/director/cmd/tenantloader/README.md).
* [Tenant Fetcher](https://github.com/kyma-incubator/compass/tree/master/components/director/cmd/tenantfetcher) - a periodic job that synchronizes tenants from an external system
Technical details can be found [here](../../components/director/cmd/tenantfetcher/README.md).

## Authentication flow
Tenant information takes important role during the authentication and authorization phase. The tenant mapping phase is done on the incomming request that is routed to the component called tenant mapping handler. The tenant mapping handler is a HTTP endpoint configured as the Oathkeeper [hydrator](https://www.ory.sh/oathkeeper/docs/pipeline/mutator/#hydrator). On of its responsibility is to map external tenant identifier to the internal one.


# Tenants Query in GraphQL API
The Compass Director GraphQL API exposes `tenants` query. The query return list of all tenants with their external identifier, internal identifier, and additional metadata. 
