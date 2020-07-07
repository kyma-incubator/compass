The Compass Director is a multitenancy service.
Tenant is an object which is owner of resources.

# Tenancy
Tenancy in Director is implemented on database level.
Every object which belongs to tenant has a special column `tenant_id`.
Tenant is mainly described by two properties: 
* the global tenant (can be any string, treated like identifier from external system) 
* the internal tenant which is used as internal technical identifier (UUID) to which columns `tenant_id` has references to.
Those properties are stored in table `business_tenant_mapping` with metadata.

Application and Runtimes are bound to tenants.
Integration System is not bound to any tenant and can work within multiple tenants.

# Authentication flow
Director has endpoint called `tenant-mapping-service`, which is a [hydrator](https://www.ory.sh/oathkeeper/docs/pipeline/mutator/#hydrator).
The responsibility of this component is to map external tenant to internal tenant and put it in request.
The mapping flow differs in case of:
* User
* Application, Runtime
* Integration System 

User flow:
- the request contains `ExternalTenant`
- the service check if given user is stored in local configuration and has access to given tenant.
- if yes, the application fetch data from `business_tenant_mapping` table and put internal tenant into request

Application and Runtime flow:
- the oauthkeeper put in request `client_id` which is `auth_id` primary key from `SystemAuth` table in the director database. 
- the request also contains `ExternalTenant`. 
- the service check if internal tenant from `SystemAuth` is the same internal tenant which is mapped by `External Tenant`.

Integration System flow:
- the oauthkeeper put in request `client_id` which is `auth_id` primary key from `SystemAuth` table in the director database.
- the request also contains `ExternalTenant`. 
- the service check if `SystemAuth` pointed by `auth_id` exist and then fetch internal tenant from `business_tenant_mapping` table from director database by`ExternalTenant`. 

Flow for: Runtime, Application and Integration system differs:
* Application, Runtime, Integration System - the request from those resources contains `auth_id` which is a primary key in `SystemAuth` table and `external tenant`.

In case of Applications and Runtime, the request has to contain headrer `External Tenant`. 
The service check if given `auth_id` belongs to the same internal tenant as `extenal tenant` if yes, the intenal tenant is added to the request.
In case of Integration System, the intenal tenant is fetched by `extenral tenant`, becasue the `auth` for integration system is not tied to any specific internal tenant.
