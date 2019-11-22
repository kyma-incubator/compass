# Limit access to GraphQL resources

This document describes proposed solution to the problem of unlimited resource access.
Limiting access to certain resolvers is important especially in case of accessing authorization details.

## Terminology

API Consumer - Application / Runtime / Integration System / User 

## Problems we want to solve

### Restrict access to `application { api { auth(runtimeID) } }` for Runtimes with ID different than passed `runtimeID`

RuntimeAgent wants to read auth details for application's API.
The RuntimeAgent should be allowed to read field `auth` only with his own ID.
In this case we will use new directive `limitAccessFor` and `limitManagerAccessTo`.

### Restrict access to `application { auths }`

RuntimeAgent wants to read applications but shouldn't have access to their auth details.
We can restrict the access by current `hasScopes` directive.

### Restrict access to resolvers for specific types of API consumers (by ID)

We should be able to restrict access to any resolver that takes API Consumer ID as one of its parameters by adding directive.

We want to cover following cases:

1. Runtime musn't read other Runtimes.
2. Application musn't read other Applications.
3. Integration System musn't read other Integration Systems.
4. Runtime musn't read Applications which are not in the same scenario (`applicationsForRuntime`).

### Restrict access to Application/Runtime for IntegrationSystem is which not managed
When Integration System request specific Application/Runtime, it should be able to read only object managed by itself.

## Solution
To achieve those restrictions the following solution is proposed:

### Add information about API consumer in Tenant Mapping service
We can use information which is added in Tenant Mapping Service to JWT [PR](https://github.com/kyma-incubator/compass/pull/475):
1. Object Type, service is able to determine who calls the API, whether it is Application, Runtime or IntegrationSystem
2. Object ID of caller

The enriched JWT is sent to the Compass API.

### Directive `limitAccessFor`
This directive will limit access to resolvers for specific types of API consumers (by ID).

Proposed directive:
```graphql
enum consumerType {
    RUNTIME
    APPLICATION
    INTEGRATION_SYSTEM
    USER
}

directive @limitAccessFor(consumerType: consumerType!, idField: String!) on FIELD_DEFINITION
```

The proposed directive does following things:
1. get object type and ID from JWT token
2. check if object is the same type as `consumerType` argument
3. if yes, then the directive compare the ID field. If the ID mismatches, the following error is returned `Access Denied`.
4. if no, the request is allowed

Currently we cannot use this directive on query param: [issue](https://github.com/99designs/gqlgen/issues/760).

### Directive `limitManagerAccessTo`
This directive will limit access to application/runtime resolvers managed by integration system

Proposed directive:
```graphql
enum managedResourceType {
    RUNTIME
    APPLICATION
}

directive @limitManagerAccessTo(managedResourceType: managedResourceType!, idField: String!) on FIELD_DEFINITION
```

The proposed directive does following things:
1. get object type and ID from JWT token
2. check if object is the `integration system`
3. if yes, then the directive check in database if given application/runtime is managed by integration system which called the query.
If integration system doesn't manage the application/runtime, the following error is returned `Access Denied`

Currently we cannot use this directive on query param ( https://github.com/99designs/gqlgen/issues/760).

### Database schema change
Every resource managed by integration system have to be reachable from integration system perspective.

### New scopes

We will introduce new scopes:
* applicationForRuntime:list - for runtimes and admin user
* runtime:list - for admin user
* application:list - for admin user
* application:auth:read - for application

We should change all **read** scopes on collections to **list** for consistency.

### Proposed solution applied on graphql
New graphql schema after implementing and applying new directives:

```graphql
type Application {
    ...
    auths: [SystemAuth!]! @hasScopes(path:"graphql.field.application.auths")
}

type APIDefinition {
    ...
    auth(runtimeID: ID!): APIRuntimeAuth! @limitAccessFor(consumerType: RUNTIME, idField: "runtimeID") @hasScopes(path:"graphql.query.application.apidefinition.apis")  
    auths: [APIRuntimeAuth!]! @hasScopes(path:"graphql.field.apidefinition.auths")
}

type Runtime {
    ...
    auths: [SystemAuth!]! @hasScopes(path:"graphql.field.runtime.auths")
}

type IntegrationSystem {
    ...
    auths: [SystemAuth!]! @hasScopes(path:"graphql.field.integrationSystem.auths")
}

type Query {
    ...   
    integrationSystem(id: ID!): IntegrationSystem @hasScopes(path: "graphql.query.integrationSystem") @limitAccessFor(consumerType: INTEGRATION_SYSTEM, idField: "id")
    runtime(id: ID!): Runtime @hasScopes(path: "graphql.query.runtime") @limitAccessFor(consumerType: RUNTIME, idField: "id")
    application(id: ID!): Application @hasScopes(path: "graphql.query.application") 
        @limitAccessFor(consumerType: APPLICATION, idField: "id")
        @limitManagerAccessTo(managedResourceType: APPLICATION, idField: "id")


    applicationsForRuntime(runtimeID: ID!, first: Int = 100, after: PageCursor): ApplicationPage! 
        @hasScopes(path: "graphql.query.applicationsForRuntime")
        @limitAccessFor(consumerType: RUNTIME, idField: "runtimeID")
}
```

## Examples

### Example flow for ApplicationsForRuntime with limitAccessFor directive
Example based on `applicationsForRuntime` flow.

We add `limitAccessFor` directive to `applicationsForRuntime` query:

```graphql
type Query {
    applicationsForRuntime(runtimeID: ID!, first: Int = 100, after: PageCursor): ApplicationPage!
    @hasScopes(path: "graphql.query.applicationsForRuntime")
    @limitAccessFor(consumerType: RUNTIME, idField: "runtimeID")  
}
```

Execution flow:
1. Tenant Mapping Service recognize API consumer, then add `ID` and `consumerType` to JWT token.
2. Directive `limitAccessFor` is executed
    * When Runtime Agent with ID `ABCD` executes query `applicationsForRuntime` with param `runtimeID` equal to `DCBA`, 
the Runtime Agent will get an error `Access Denied`, because the IDs are different.
    * When IntegrationSystem with ID `ABCD` executes query `applicationsForRuntime` with param `runtimeID` equal to `DCBA`, 
The directive doesn't compare anything, because it's only turn on for the `RUNTIME`.

### Example flow for applications with limitManagerAccessTo directive
Example based on `application(id)` flow.
In this case besides `limitAccessFor` directive we are also going to apply `limitManagerAccessTo`.

We add `limitManagerAccessTo` directive to `application` query:

```graphql
type Query {
    application(id: ID!): Application!
    @hasScopes(path: "graphql.query.applicationsForRuntime")
    @limitAccessFor(consumerType: APPLICATION, idField: "id")  
    @limitManagerAccessTo(managedResourceType: APPLICATION, idField: "id")  
}
```

Execution flow:
1. Integration system asks for application with id `DCBA`
1. Tenant Mapping Service recognize API consumer, then add `ID` and `consumerType` to JWT token.
2. Directive `limitAccessFor` allows the request because consumer is not an application.
3. Directive `limitManagerAccessTo` is executed.
   The directive checks if given application is managed by integration system, if yes the request is allowed, in other case the `Access Denied` error is returned.
