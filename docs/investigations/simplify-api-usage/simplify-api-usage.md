# Simplify usage of Compass API

## Terminology:
`consumer` - client of a Compass API (Application, Runtime, Integration System)

## Overview

This document discusses how to simplify usage of Compass API

As of right now, there are some mutations that are related to a base entity (e.g. Application or Runtime) and they require the ID of the base entity:
 
`addAPI(applicationID: ID!, in: APIDefinitionInput! @validate): APIDefinition! @hasScopes(path: "graphql.mutation.addAPI")` 

needs `applicationID`, and

`setRuntimeLabel(runtimeID: ID!, key: String!, value: Any!): Label! @hasScopes(path: "graphql.mutation.setRuntimeLabel")`
 
 needs `runtimeID`.

So when an application wants to add API for itself, it needs to provide its own ID. That behaviour is inconvenient and should be changed.
Although the option to provide an consumerType ID should stay. Take for example a case when an Integration System creates an Application.

There is another case when Integration System creates an application, it also has to provide its own ID to the `integrationSystemID` parameter inside `ApplicationCreateInput`.


## Requirements
* `consumer id (e.g. applicationID)` not required

## Possible solutions

### 1. `consumer ID` should be optional for operations related to itself
 
The `applicationID`/`runtimeID` fields could be changed to be optional in mutations in the schema
 
The idea is to create a directive which will fetch the ID of the consumer when it's not present.
It is possible due to the fact that it can be retrieved from the request context.
The directive could be used both on parameter(e.g. `addAPI` mutation) and field inside input(`integrationSystemID` case).
Also, it would be possible for a consumer to update itself (`updateApplication` mutation) and not provide it's `ID`
 
**Work that has to be done**
* prepare directive which will retrieve `ID` depending on consumer type
* change `ID` parameters to optional in mutations in GraphQL schema
* adjust resolvers

**Pros**
* no existing code needs to change

**Cons**
* A LOT work with tests, see that example: 

considering `updateApplication` mutation, we would have to write tests to cover these cases:

|                    | Is `applicationID` provided? | Should pass? |
|--------------------|:---------------:|--------------|
| User               | YES             | YES          |
| User               | NO              | NO           |
| Application        | YES             | YES          |
| Application        | NO              | YES          |
| Runtime            | YES             | NO           |
| Runtime            | NO              | NO           |
| Integration system | YES             | YES          |
| Integration system | NO              | NO           |

There are 8 test cases in just one mutation, having in mind we would change around 20 mutations, it would be around 160 tests. 

* API becomes less readable due to optional `ID` parameter (the user could be confused what happens if we don't provide one)
* domains dependant on another mechanism which will fetch `consumer ID`

### 2. Separated mutations where `consumer ID` input parameter is not needed
There could be another mutation which doesn't accept `applicationID`/`runtimeID` as a parameter, so we would have doubled mutations. Take this for example:

`addAPI(applicationID: ID, in: APIDefinitionInput! @validate): APIDefinition! @hasScopes(path: "graphql.mutation.addAPI")`

and 

`addAPIForApplication(in: APIDefinitionInput! @validate): APIDefinition! @hasScopes(path: "graphql.mutation.addAPI")`

**Work that has to be done**
* add new mutations to schema
* implement the new mutations (for example for addAPIForApplication mutation we could reuse addAPI implementation)
* test the new mutations (unit and integration)

**Pros**
* separated mutation for a special use case

**Cons**
* API becomes a lot bigger - around 20 new mutations
* the same situation with tests as described in first solution

### 3. The Viewer pattern
The alternative to solutions presented above is the Viewer pattern.
It is a GraphQL query that looks like:
```graphql
viewer: Viewer!
type Viewer {
  id: ID!
  type: ViewerType!
}
```
Usage:
```graphql
viewer {
   id 
   type // Application/Runtime/Integration System
}

```

> The viewer field represents the currently logged-in user; its subfields expose data that are contextual to the user.

The consumer ID is present in request context, so it can be retrieved from there.

**Work that has to be done**
* prepare and implement the Viewer query
* test the query

**Pros**
* small amount of work
* API doesn't get much bigger(only one query)

**Cons**
* consumer has to make an additional call to get its ID for other GraphQL operations 

Sources:

https://medium.com/workflowgen/graphql-schema-design-the-viewer-field-aeabfacffe72
https://codeahoy.com/2019/10/13/graphql-practical-tutorial/

The Viewer pattern is used in e.g. Facebook and GitHub API:

https://developer.github.com/v4/query
https://github.com/graphql/graphql-js/issues/571

## Decision

The first and second option doesn't seem to be good due to:
- a lot of work
- significant increase of Compass API
- Compass API get less readable

The best option is the third one - the Viewer pattern. Not only we achieve the requirement with the low workload, but it is also less painful to test it.
