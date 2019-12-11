# Simplify usage of Compass API for paired Application

## Overview

This document discusses how to simplify usage of Compass API for paired Application.

As of right now, mutations that are related to Application type require `applicationID` parameter, for example:
As of right now, there are some mutations that are related to a base entity (e.g. Application or Runtime) and they require the ID of base entity, for example:
 
`addAPI(applicationID: ID!, in: APIDefinitionInput! @validate): APIDefinition! @hasScopes(path: "graphql.mutation.addAPI")` 

needs `applicationID`, and

`setRuntimeLabel(runtimeID: ID!, key: String!, value: Any!): Label! @hasScopes(path: "graphql.mutation.setRuntimeLabel")`
 
 needs `runtimeID`.

So when application wants to add API for itself, it needs to provide it's own ID. That behaviour is inconvenient and should be changed.
Although the option to provide an consumerType ID should stay. Take for example a case when Integration System creates an Application.

There is another case when Integration System creates an application, it also has to provide it's own ID to the `integrationSystemID` parameter inside `ApplicationCreateInput`.

## Requirements
* `consumer id (e.g. applicationID)` not required

## Possible solutions

### 1. `consumer ID` should be optional for operations related to Application/Runtime
 
The `applicationID`/`runtimeID` fields could be simply changed to be optional in mutations listed below:
 - addWebhook
 - addAPI
 - addEventAPI
 - addDocument
 - setApplicationLabel
 - applicationsForRuntime
 - setAPIAuth
 - setRuntimeLabel
 
**Work that has to be done**
* prepare directive which will retrieve `ID` depending on consumer type (from request context)
* change `applicationID`/`runtimeID` parameter to optional in input types inside GraphQL schema
* adjust resolvers

**Pros**
* relatively small amount of work to be done
* no tests need to change

**Cons**
* domains dependant on another mechanism which will fetch `applicationID`

### 2. Separated mutations where `applicationID` input parameter is not needed
The `applicationID` field could be simply changed to be optional in domains listed below:
There could be another mutation which doesn't accept `applicationID` as a parameter, so we would have two mutations:

`addAPI(applicationID: ID, in: APIDefinitionInput! @validate): APIDefinition! @hasScopes(path: "graphql.mutation.addAPI")`

and 

`addAPIForApplication(in: APIDefinitionInput! @validate): APIDefinition! @hasScopes(path: "graphql.mutation.addAPI")`

**Work that has to be done**
* prepare mechanism to retrieve `applicationID` when the caller is an application
* add new mutation to schema
* implement the new mutation (for example for addAPIForApplication mutation we could reuse addAPI implementation)
* test the new mutation(unit and integration)

**Pros**
* seperated mutation for a special use case

**Cons**
* requires more work than the first solution
* has to be maintained

## Decision

The decision made by the team is that we will implement the first option, and maybe in the future we will add more features to it.
Right now it will still improve the newcomer experience.

The second option was rejected due to the high effort to customize the tools to our needs.

