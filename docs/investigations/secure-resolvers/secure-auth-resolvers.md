# Limit access of GraphQL objects.

## Introduction
Currently in compass runtimeAgent can read information about every runtime and can read `auths` without any problem.

## Terminology

API Consumer - Application / Runtime / Integration System / User 

## Problems we want to solve

### Restrict access to resolvers for specific types of API consumers (by ID). 

We should be able to restrict access to any resolver that takes API Consumer ID as one of its parameters by adding directive.
For example we could restrict access to `applicationsForRuntime(runtimeID: ID!, first: Int = 100, after: PageCursor): ApplicationPage!` for Runtimes with ID different than provided `runtimeID`.
Anothe

Every object should be able to read information only about itself.

We should restrict access to other objects of the same type, for example: runtime musn't read other runtime configuraion.
We have two kinds of such resolvers:
* `runtime(ID)` runtime should be able to only read about itself
* `runtimes` - this data should be limited for `RuntimeAgents` with `hasScopes` directive, which restricts access to the resolver to the certain scopes.
Also we have resolvers with 
TODO: divide those examples

The runtime agent execute query `applicationsForRuntime` with `runtimeID` param.
Runtime should't be able to call this query with different `runtimeID`. 
In case of different ID, we should return  error such as `Access Denied`.

### Restrict access to `auths`:

RuntimeAgent wants to read all Applications.
The application has field `auths`.
The RuntimeAgent shouldn't have access to reading `auths`.
We can restrict the access by current `hasScopes` directive.

### Restrict access to `auth` in `api`:

RuntimeAgent wants to read all Applications with their APIs. 
`Apis` contain field `auths` which contains `APIRuntimeAuth` and field `auth` with `runtimeID` param.
The RuntimeAgent shouldn't have access to reading `auths`.
We can restrict the access by `hasScopes` directive.

RuntimeAgent should be allowed to read field `auth` with owned RuntimeID parameter.

## Solution


### Tenant Mapping service
Tenant Mapping service adds 2 items to the header
1. Object Type, service is able to determine who calls the API, whether it is Application, Runtime or IntegrationSystem
2. Object ID

The enriched request is sent to the compass API.

### Directive

Po co ta dyrektywa itp

The directive does following things:
* get object type and ID from request header
* check if object is the same type as`Owner` argument
* if yes, then the directive compare the ID field. If the ID mismatches, the following error is returned `Access Denied`.

```graphql
directive @limitAccessFor(type: objectType, idField: String) #can be used for query.
```

Object Types:
* Runtime
* Application
* Integration System
* User

Currently we cannot use this directive on query param ( https://github.com/99designs/gqlgen/issues/760).

## Example flow for ApplicationForRuntime with resourceOwner directive
Tutaj opisać to lepiej, flow w jakiś podpunktach
We have following `graphql` query.
```graphql
applicationsForRuntime(runtimeID: ID!, first: Int = 100, after: PageCursor): ApplicationPage! 
@hasScopes(path: "graphql.query.applicationsForRuntime")
@limitAccessFor(type: RUNTIME, idField: "runtimeID")  
```

When Runtime Agent with ID `ABCD` executes query `applicationsForRuntime` with param `runtimeID` equals to `DCBA`, 
the Runtime Agent will get an error `Access Denied`.

When IntegrationSystem with ID `ABCD` executes query `applicationsForRuntime` with param `runtimeID` equals to `DCBA`, 
The directive doesn't check anything, because it's only turn on for the `RUNTIME`.

## Proposed solution applied on graphql
Dodaj jak to bedzie wygladało w rzeczywistosci
New graphql schema after implementing and applying new directives:

```graphql
type Application {
    ...
    auths: [SystemAuth!]! @hasScopes(path:"graphql.query.application.auths")
}

type APIDefinition {
    ...
	auth(runtimeID: ID!): APIRuntimeAuth!@limitAccessFor(type: RUNTIME, idField: "runtimeID")  
	auths: [APIRuntimeAuth!]! @hasScopes(path:"graphql.query.application.write")???
}

type Runtime {
    ...
	auths: [SystemAuth!]! @hasScopes(path:"graphql.query.runtime.auths")
}

type IntegrationSystem {
    ...
	auths: [SystemAuth!]! @hasScopes(path:"graphql.query.integrationSystem.auths")
}

type Query {
    ...   
    integrationSystem(id: ID!): IntegrationSystem @hasScopes(path: "graphql.query.integrationSystem") @limitAccessFor(type: INTEGRATION_SYSTEM, idField: "id")
    runtime(id: ID!): Runtime @hasScopes(path: "graphql.query.runtime") @limitAccessFor(type: APPLICATION, idField: "id")
    application(id: ID!): Application @hasScopes(path: "graphql.query.application") @limitAccessFor(type: RUNTIME, idField: "id")

    applicationsForRuntime(runtimeID: ID!, first: Int = 100, after: PageCursor): ApplicationPage! 
        @hasScopes(path: "graphql.query.applicationsForRuntime")
        @limitAccessFor(type: RUNTIME, idField: "runtimeID")
}
```

TODO: Remove this below
### 

List of resolvers which should be secured with `@limitAccessFor` directive:
* integrationSystem
* runtime
* application

also all `auths` field should be secured with `@hasScopes` in following types:


### `hasScopes` directive



* IntegrationSystem
* Runtime
* Application
* APIDefinition

dodaj przyklady z graphql
