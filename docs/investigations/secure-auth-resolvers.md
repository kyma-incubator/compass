# Limit access of GraphQL objects or fields for specific resource owners

### Case 1:

Every object should has access to read information only about itself.
We should restrict access to other objects of the same type, for example: runtime musn't read other runtime configuraion.
We have two kind of such resolvers:
* `runtime(ID)` runtime should be able to only read about itself
* `runtimes` runtime should get list only with itself

Also we have fields with custom resolvers like this:
In `APIDefinition`, there is a field `auth` with param `runtimeID`, runtime which ask for 
this field should be validated if ask with own `ID`.


#### Example:
RuntimeAgent wants `api` with field `auth`.
The `auth` field has custom resolver with 1 parameter `runtime_id`.
Before executing this resolver, the directive is executed.
Runtime with ID `ABCD` wants to fetch `auth` for runtime ID `DCBA`.
It's should be forbidden.

We have 2 options:
- return error
- return empty auth

For me more suitable is to return error such as `Access Denied`.

### Case 2:

RuntimeAgent want to read all Application.
The application has field `auths`.
The runtime should be able to read `auths` only connected with given runtime.

#### Example:

RuntimeAgent wants `api` with field `auths`.

Field `Auths` contains all `auths` for all runtimes.
We want to return `auths` connected with the asking runtime.

### Case 3:

RuntimeAgent want to read all Application with their Apis. 
`Apis` contain field `auths` which contain `APIRuntimeAuth`
The data in `auths` field should be filtered to contains `auths` only connected to the given RuntimeAgent.

### Questions
`Fetch Request`, `Webhook` has also  field `auth`, who does have the ability to fetch this field?

Connected [Issue](https://github.com/kyma-incubator/compass/issues/306)

## Proposed solution

### Tenant Mapping service
Tenant Mapping service adds 2 items to the header
1. Object Type, service is able to determine who calls the API, whether it is Application, Runtime or IntegrationSystem
2. Object ID

The enriched request is sent to the compass.

### Directives
Object Types:
*Runtime
*Application
*Integration System

Then we create two directives:
* limitScope(For: []objectTypes) - this directive does following things:
    * get object type and id from request header
    * check if object type exist in `For` argument
    * if exist, the directive put to context following information:
        * key: column_name/property, depends on object type
        * value: object ID
      * if not exist, the query underneath is not modified
     Then in query helpers we modify the query, by adding additional filtering conditions.

* restrict(For: []objectType) - does following things:
    * get object type and id from request header
    * check if object type exist in `For` argument
    * if yes, then the directive compare the IDs. If the ID mismatches, the following error is returned `Access Denied`.

Those directives works like `blacklists`.

Second proposition:

The second proposition is to use `whitelists` in `limitScope` directive.

Example:
* limitScope(FullAccess: []objectTypes):
    * get object type and it from request header
    * check if object exist in full access array
    * if not exists, the directive put to context following information:
        * key: column_name/property, depends on object type
        * value: object ID
    * if exist, the query underneath is not modified

## Example Flow

Our directive is set on `auths` field in `APIDefinition` object and parameter `For` contains object type `Runtime`.
RuntimeAgent fetches Application with `apis` and `apis` with `auths`.
Directive put in context `type` and `id` of the RuntimeAgent.
In `Get`  repository helpers we need following condition to the query:
`runtime_id == ID`
By adding this condition we filtered `auths` which doesn't belong to given runtime.
