# Limit access of GraphQL objects or fields for specific resource owners

### Case 1:

Every object should has access to read information only about itself.
We should restrict access to other objects of the same type, for example: runtime musn't read other runtime configuraion.
We have two kind of such resolvers:
* `runtime(ID)` runtime should be able to only read about itself
* `runtimes` - this data is limited for `RuntimeAgents` with `hasScopes` directive, so it's not our case.

### Case 2:

The runtime agent execute query `applicationsForRuntime` with `runtimeID` param.
Runtime should't be able to call this query with different `runtimeID`. 
In case of different ID, we should return  error such as `Access Denied`.

### Case 3:

RuntimeAgent want to read all Applications.
The application has field `auths`.
The RuntimeAgent should't has access to reading `auths`.
We can restrict the access by current `hasScopes` directive.

### Case 4:

RuntimeAgent want to read all Application with their Apis. 
`Apis` contain field `auths` which contains `APIRuntimeAuth` and field `auth` with `runtimeID` param.
The RuntimeAgent should't has access to reading `auths`.
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

resourceOwner(Owner: []objectType, Id_Field: string) - can be used for query.
Currently we cannot use this directive on query param ( https://github.com/99designs/gqlgen/issues/760).
The directive does following things:
* get object type and ID from request header
* check if object type exist in `For` argument
* if yes, then the directive compare the ID field. If the ID mismatches, the following error is returned `Access Denied`.

## Example flow for ApplicationForRuntime with resourceOwner directive

```graphql
applicationsForRuntime(runtimeID: ID!, first: Int = 100, after: PageCursor): ApplicationPage! 
@hasScopes(path: "graphql.query.applicationsForRuntime")
@resourceOwner(For: [RUNTIME], Id_Field: "runtimeID")  
```

When runtimeAgent with ID `ABCD` execute query `applicationsForRuntime` with param `runtimeID` equal to `DCBA`, 
the runtime agent will get error `Access Denied`.

When IntegrationSystem with ID `ABCD` execute query `applicationsForRuntime` with param `runtimeID` equal to `DCBA`, 
The directive dont' check anything, because it's only turn on for `RUNTIME`.

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
