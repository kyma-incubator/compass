# Secure resovlers to be able to show only the allowed fields
Purpose:

[Issue](https://github.com/kyma-incubator/compass/issues/306)

Secure auth fields to show values only for the right caller.

# First Scenario resolver with ID parameter
RuntimeAgent want `api` with field `auth`.
The `auth` field has custom resolver with 1 parameter `runtime_id`.
Before executing this resolver, the directive is executed.
When runtime with ID `ABCD` want to get `auth` for runtime ID `DCBA`.
It's should be forbidden.
We have 2 options:
- return error
- return empty auth

For me more suitable is to return error such as `Access Denied`.

## Second scenario resolver without ID parameter
RuntimeAgent want `api` with field `auths`.

Field `Auths` contains all `auths` for all runtime.
We want to return `auths` connected with runtime which ask for this field. 

## Proposed Flow
### First step (tenant mapping service):

Tenant Mapping service add 2 items to header
1. Object Type, service is able to differ who called the API, if it was App, Runtime or IntegrationSystem
2. Object ID

The request is send to compass.

### Second step for resolver without parameters
Compass has such directive:

`limitScope(For: objectType)`

In this directive, we retrieve Object Type and Object ID.
If object type match type from header, then we put it into context such pairs:

- key: column_name/property, depends on object type
- value: object ID

We can name it as a list of restrictions.

### Second step for resolver with parameters
Compass has such directive:

`restrict(For: objectType)`

In this directive, we retrieve Object Type and Object ID.
If object type match type from header, then we compare the ids.
If don't match we return error for this resolver with message `Access Denied`

### Third step
In our repository we have compliant column name (if column refer to application it has name `app_id`. 
Repository helper `Get` and `List` retrieve the value from context and construct the conditions, which are later added to conditons list.

## Summary
Those directives can be used to restrict other resolvers.
