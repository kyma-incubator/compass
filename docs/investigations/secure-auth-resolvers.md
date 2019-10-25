# Limit access of GraphQL objects or fields for specific resource owners

### Case 1:

Every object should has access to read information only about itself.
We should restrict access to other objects of the same type, for example: runtime musn't read other runtime configuraion.
We have two kind of such resolvers:
* `runtime(ID)` runtime should be able to only read about itself
* `runtimes` - this data is limited for `RuntimeAgrnts` with `hasScopes` directive, so it's not our case.

The runtime agent execute query `applicationForRuntime` 
which returns ApplicationPage, Application has collection of APIDefinition and APIDefinition has field `auth` and `auths`.

RuntimeAgent should be able to call `auth` resolver with own `ID`.
In case of different ID, we should return  error such as `Access Denied`.

RuntimeAgent should be able to see `auths` only connected with itself.

### Case 2:

RuntimeAgent want to read all Application.
The application has field `auths`.
The RuntimeAgent should't has access to reading `auths`.
We can restrict the access by current `hasScopes` directive.

#### Example:

RuntimeAgent wants `api` with field `auths`.

Field `Auths` contains all `auths` for all runtimes.
We want to return `auths` connected with the asking runtime.

### Case 3:

RuntimeAgent want to read all Application with their Apis. 
`Apis` contain field `auths` which contain `APIRuntimeAuth`
The RuntimeAgent should't has access to reading `auths`.
We can restrict the access by current `hasScopes` directive.
RuntimeAgent should read field `auth` with RuntimeID parameter.

### Questions
`Fetch Request`, `Webhook` has also  field `auth`, who does have the ability to fetch this field?

## Proposed solution

### Tenant Mapping service
Tenant Mapping service adds 2 items to the header
1. Object Type, service is able to determine who calls the API, whether it is Application, Runtime or IntegrationSystem
2. Object ID

The enriched request is sent to the compass.

### Directive
Object Types:
* Runtime
* Application
* Integration System

resourceOwner(Owner: []objectType) - can be used for query parameter. The directive does following things:
* get object type and id from request header
* check if object type exist in `For` argument
* if yes, then the directive compare the IDs. If the ID mismatches, the following error is returned `Access Denied`.

## Example flow for ApplicationForRuntime with resourceOwner directive

applicationsForRuntime(runtimeID: ID!, first: Int = 100, after: PageCursor): ApplicationPage! 
@hasScopes(path: "graphql.query.applicationsForRuntime")
@resourceOwner(For: [RUNTIME])

When runtimeAgent with ID `ABCD` execute query `applicationsForRuntime` with param `runtimeID` equal to `DCBA`, 
the runtime agent will get error `Access Denied`.
