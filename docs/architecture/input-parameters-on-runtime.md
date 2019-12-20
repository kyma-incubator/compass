# Input parameters for API, Event Definitions or API Packages in the Runtime

## Overview

Previously, in Kyma Runtime, while provisioning Service Class created from API/Event Definition, there were no possibility to pass input parameters. The ability may be useful in some specific use cases. This document describes passing input parameters from Kyma Runtime to Application or Integration System.

## Assumptions
- Multiple Service Instances can be created from a given API, Event Definition or API Package
- During Service Class provisioning, the provided input has to be sent back to Integration System or Application via Compass. The reason is that there is no trusted connection between Integration System and Runtime.

## Solution

### Auths per Service Instance

Previously, for every API Definition there was a list of credentials per Runtime:

```graphql
type APIDefinition {
    # (...)
	auth(runtimeID: ID!): APIRuntimeAuth!
	auths: [APIRuntimeAuth!]!
	defaultAuth: Auth
}

type APIRuntimeAuth {
	runtimeID: ID!
	auth: Auth
}

mutation {
    setAPIAuth(apiID: ID!, runtimeID: ID!, in: AuthInput!): APIRuntimeAuth!
    deleteAPIAuth(apiID: ID!, runtimeID: ID!): APIRuntimeAuth!
}
```

As there could be multiple Service Instances on Runtime with different credentials, the API is modified to take into the account this fact:

```graphql
type APIDefinition {
    # (...)
	auth(runtimeID: ID!, usageID: ID!): APIUsageAuth!
	auths(runtimeID: ID): [APIUsageAuth!]!
	defaultAuth: Auth
}

type APIUsageAuth {
    id: ID! # Usage ID, which equals to Service Instance ID on a given Runtime
	runtimeID: ID!
	auth: Auth
}

mutation {
    setAPIUsageAuth(runtimeID: ID!, usageID: ID!, in: AuthInput!): APIUsageAuth!
    deleteAPIUsageAuth(runtimeID: ID!, usageID: ID!): APIUsageAuth!
}
```

### Instance Create Parameter Schema

> **NOTE:** There is no final decision - we need to figure out the final approach.

#### Option #1: Instance Create Parameter Schema in LabelDefinition

The solution utilizes LabelDefinitions and labels on Application level. It requires implementation of LabelDefinition extension: an ability to create LabelDefinitions for a specific prefix. All labels starting with the specified prefix are validated against JSON schema provided in the LabelDefinition.

1. Integration System or Application provides JSON schema for Service Class `instanceCreateParameterSchema` parameter by creating new Label Definition in the following format:

    **Key**: 
    - For API Definition: `input-params.{app-id}.apis.{api-def-id}/*`
    - For Event Definition: `input-params.{app-id}.events.{event-def-id}/*`
    - For API Package: `input-params.{app-id}.api-packages.{api-package-id}/*`
        
    **Value**: JSON schema for the `instanceCreateParameterSchema` for given Service Class

1. On Kyma Runtime, Runtime Agent queries Applications for given Runtime. It reads all LabelDefinition with prefix `input-params.{app-id}` and updates `instanceCreateParameterSchema` properties in Service Classes.

1. During Service Instance creation, instance parameters are set as label for Application. For example, for API Definition, a key `input-params.{app-id}.apis.{api-def-id}/{instance-id}` is used.

1. To list all Service Instance params, Application or Extension Service can query Application labels with prefix `input-params.{app-id}.apis.{api-def-id}/{instance-id}`.

**Pros:**

- Generic approach; ability to reuse this feature in other use cases
- Static validation of JSON Schema

**Cons:**

- Introduce new feature on Compass side
    - Need to resolve LabelDefinition conflicts (for example, LabelDefinition `foo/*` and LabelDefinition `/foo/bar/*`: we can block it or enable overrides)
- Runtime Agent has to read LabelDefinitions for a specific prefixes
    - This can be solved with an optional parameter with for `labelDefinitions` query, like `labelDefinitions(prefix: String)`

#### Option #2: Instance Create Parameter Schema in Label

1. Application or Integration System creates Application label with `instanceCreateParameterSchema` for given API, Event Definition or API Package. For example, in case of API Definition, it is `input-params.apis.{api-def-name}.schema`
1. During Service Instance creation, instance parameters are set as label `input-params.apis/{api-def-name}.usages` for given Application. For example:

    ```json
    {
        "{instance-1-uuid}": {"foo": "bar", "baz": 3},
        "{instance-2-uuid}": {"foo": null, "baz": 10},
    }
    ```

1. To list parameters for Service Instances of given API Definition, Application or Extension Service reads Application label with key `input-params.apis.{api-def-name}.usages`.

**Pros:**

- No changes in Compass API (apart from Auths)

**Cons:**

- No static validation for input parameters
    - To validate Service Instance params, Integration System or Application can create upfront LabelDefinition for `input-params.apis.{api-def-name}.usages`.
