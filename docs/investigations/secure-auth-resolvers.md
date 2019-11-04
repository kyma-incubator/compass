# Limit access of GraphQL objects or fields for specific resource owners

### Case 1:

Every object should be able to read information only about itself.
We should restrict access to other objects of the same type, for example: runtime musn't read other runtime configuraion.
We have two kinds of such resolvers:
* `runtime(ID)` runtime should be able to only read about itself
* `runtimes` - this data is limited for `RuntimeAgents` with `hasScopes` directive, which restricts access to the resolver to the certain scopes.

### Case 2:

The runtime agent execute query `applicationsForRuntime` with `runtimeID` param.
Runtime should't be able to call this query with different `runtimeID`. 
In case of different ID, we should return  error such as `Access Denied`.

### Case 3:

RuntimeAgent wants to read all Applications.
The application has field `auths`.
The RuntimeAgent shouldn't have access to reading `auths`.
We can restrict the access by current `hasScopes` directive.

### Case 4:

RuntimeAgent wants to read all Applications with their APIs. 
`Apis` contain field `auths` which contains `APIRuntimeAuth` and field `auth` with `runtimeID` param.
The RuntimeAgent shouldn't have access to reading `auths`.
We can restrict the access by `hasScopes` directive.

RuntimeAgent should be allowed to read field `auth` with owned RuntimeID parameter.

## Proposed solution for resolvers with parameter ID

### Tenant Mapping service
Tenant Mapping service adds 2 items to the header
1. Object Type, service is able to determine who calls the API, whether it is Application, Runtime or IntegrationSystem
2. Object ID

The enriched request is sent to the compass API.

### Directive
Object Types:
* Runtime
* Application
* Integration System

```graphql
directive @resourceOwner(Owner: objectType, IDField: String) #can be used for query.
```

Currently we cannot use this directive on query param ( https://github.com/99designs/gqlgen/issues/760).
The directive does following things:
* get object type and ID from request header
* check if object is the same type as`Owner` argument
* if yes, then the directive compare the ID field. If the ID mismatches, the following error is returned `Access Denied`.

## Example flow for ApplicationForRuntime with resourceOwner directive

```graphql
applicationsForRuntime(runtimeID: ID!, first: Int = 100, after: PageCursor): ApplicationPage! 
@hasScopes(path: "graphql.query.applicationsForRuntime")
@resourceOwner(Owner: RUNTIME, Id_Field: "runtimeID")  
```

When Runtime Agent with ID `ABCD` executes query `applicationsForRuntime` with param `runtimeID` equals to `DCBA`, 
the Runtime Agent will get an error `Access Denied`.

When IntegrationSystem with ID `ABCD` executes query `applicationsForRuntime` with param `runtimeID` equals to `DCBA`, 
The directive doesn't check anything, because it's only turn on for the `RUNTIME`.

### Resources to secure

List of resolvers which should be secured with `@resourceOwner` directive:
* integrationSystem
* runtime
* application

also all `auths` field should be secured with `@hasScopes` in following types:
* IntegrationSystem
* Runtime
* Application
* APIDefinition
