# Simplify usage of Compass API using ID injection

## Terminology:
`consumer` - client of a Compass API (Application, Runtime, Integration System)

## Overview

This document discusses how to simplify usage of Compass API using ID injection mechanism.

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
* relatively small amount of work to be done
* no code needs to change

**Cons**
* API becomes less readable due to optional `ID` parameter (the user could be confused what happens if we don't provide one)
* domains dependant on another mechanism which will fetch `consumer ID`

### 2. Separated mutations where `consumer ID` input parameter is not needed
There could be another mutation which doesn't accept `applicationID`/`runtimeID` as a parameter, so we would have doubled mutations. Take this for example:

`addAPI(applicationID: ID, in: APIDefinitionInput! @validate): APIDefinition! @hasScopes(path: "graphql.mutation.addAPI")`

and 

`addAPIForApplication(in: APIDefinitionInput! @validate): APIDefinition! @hasScopes(path: "graphql.mutation.addAPI")`

**Work that has to be done**
* prepare directive which will retrieve `ID` depending on consumer type
* add new mutations to schema
* implement the new mutation (for example for addAPIForApplication mutation we could reuse addAPI implementation)
* test the new mutation(unit and integration)

**Pros**
* separated mutation for a special use case

**Cons**
* requires more work than the first solution

## Decision

The first option seems to be better due to less work effort and easier to maintain(there could be more consumer types in future). However, it has to be well documented, to not get users confused about why the `ID` parameter is optional.

The second option was rejected due to greater workload.
