# Tenancy
Tenant is a single object that owns resources. Every tenant's resources are isolated from the other tenants, which ensures data security and privacy.
Compass Director is a multi-tenant service, which means that a single instance of Director is shared among different tenants (clients), but their data is isolated.

## Tenancy in Director
Director manages the configuration of the following main objects:
* Application
* Runtime
* Integration System

Applications and Runtimes and their child resources, such as APIs, are bound to tenants, whereas Integration Systems can represent multiple tenants.

Tenancy in Director is implemented on the level of a database in which a given tenant has its own entity.
Every object which belongs to a tenant has the `tenant_id` column which points to the actual tenant entity.
A tenant is mainly described by two properties: 
* Global tenant identifier - can be any string, treated like an identifier from an external system 
* Internal tenant identifier - used as an internal technical identifier (UUID) to which `tenant_id` columns refer to. It allows for unified tenant identification in Director. Thanks to this approach, external systems can describe their tenants in their own way without any impact on the Director internals.

Those properties are stored together with the metadata in the `business_tenant_mapping` table in the database.

The Compass Director GraphQL API exposes `tenants` query. 
The query return list of all tenants with their external identifier, internal identifier, and additional metadata. 

## Creating tenants
To create tenant in Director you can insert them by hand using SQL Statement or use one of the following importing mechanism:
* [Tenant Importer](https://github.com/kyma-incubator/compass/tree/master/components/director/cmd/tenantloader) - a one-time job for importing tenants from files during the first Compass installation
* [Tenant Fetcher](https://github.com/kyma-incubator/compass/tree/master/components/director/cmd/tenantfetcher) - a periodic job that synchronizes tenants from an external system

## Authentication flow
Information about tenants is used during the authentication and authorization phase in Compass. 
Every incoming request is routed to the component called Tenant Mapping Handler. 
The responsibility is to map an external tenant identifier to the internal one.
The service look for external tenant identifier in headers or body under key `headers` or `extra`.
Having external tenant identifier the component can continue tenant mapping using `business_tenant_mapping` and `system_auth` table.
The additional data (internal tenant identifier, identifier of caller) is put into JWT Token and send to Director.
The Director validates the token and extract `tenant` id from token, which is the internal tenant identifier.
Having this information, director is able to work as multi tenant service.
