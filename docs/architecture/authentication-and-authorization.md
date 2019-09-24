# Authentication and Authorization

## Introduction
Currently, communication between the Compass and both Runtimes and Applications is not secured. We need to provide some security possibilities.
We want to secure the Compass using ORY's Hydra and OathKeeper. There will be three ways of authentication:
 - OAuth 2.0 
 - Client Certificates (mTLS)
 - JWT token issued by identity service

 To achieve that, first, we need to integrate Hydra and OathKeeper into the Compass. We also need to implement additional supporting components to make our scenarios valid.

## Architecture

The following diagram represents the architecture of the security in Compass:

![](./assets/security-architecture.svg)

### Tenant Mapping Handler

It is an OathKeeper [hydrator](https://github.com/ory/docs/blob/525608c65694539384b785355d293bc0ad00da27/docs/oathkeeper/pipeline/mutator.md#hydrator) handler responsible for mapping authentication session data to tenant. It is built into Director itself, as it uses the same database. It is implemented as a separate endpoint, such as `/tenant-mapping`.

To unify the approach for mapping, we introduce `authorization_id`, widely used by multiple authentication flows.
The `authorization_id` is equal to:
- `client_id` in OAuth 2.0 authentication flow 
- `username` in Basic authentication flow
- Common Name (CN) in Certificates authentication flow

While generating one-time token, `client_id`/`client_secret` pair or basic authentication details for Runtime/Application/Integration System (using proper GraphQL mutation on Director), an entry in `system_auths` table in Director database is created. The `system_auths` table is used for tenant mapping.

In certificates authentication flow, Tenant Mapping Handler puts fixed `scopes` into authentication session. The `scopes` are fixed in code, and they depends of type of the object (Application / Runtime / Integration System). In future we may introduce another database table for storing generic Application / Runtime / Integration System scopes.

In JWT token from identity service flow, for local development, user `tenant` and `scopes` are loaded from ConfigMap for given user (email), where static `user: tenant and scopes` mapping is done.

#### `system_auths` table

The table is used by Director and Tenant Mapping Handler. It contains the following fields:
- `id`, which is the `authorization_id`
- `tenant` (optional field - used for Application and Runtime, not used for Integration System)
- `app_id` foreign key of type UUID
- `runtime_id` foreign key of type UUID
- `integration_system_id` foreign key of type UUID
- `value` of type JSONB (with authentication details, such as `client_id/client_secret` in OAuth 2.0 authentication flow, `username/password` in Basic authentication flow; in case of certificates flow it is empty)

### GraphQL security

The Gateway passes request along with JWT token to Compass GraphQL services, such as Director or Connector. The GraphQL components have authentication middleware and GraphQL [directives](https://graphql.org/learn/queries/#directives) set up for all GraphQL operations (and some specific type fields, if necessary).

#### HTTP middleware

In GraphQL servers, such as Director or Connector, there is a HTTP authentication middleware set up, which validates and decodes JWT token. It puts user scopes and tenant in request context 
(`context.Context`).

![](./assets/graphql-security.svg)

#### GraphQL Directives

When GraphQL operation is processed, an authorization directive is triggered, before actual GraphQL resolver. It checks if the client has required scopes to do the requested operation. To avoid defining permissions statically in GraphQL schema, a YAML file is loaded with needed requests. In fact, it is a ConfigMap injected to Director/Connector. 

The following example illustrates how we can implement dynamic comparison between required scopes and request scopes for example mutations, queries and type fields:

```graphql

type Mutation {
    createApplication(in: ApplicationInput!): Application! @secureWithScopes(path: "mutations.createApplication")
    updateApplication(id: ID!, in: ApplicationInput!): Application! @secureWithScopes(path: "mutations.updateApplication")
    deleteApplication(id: ID!): Application @secureWithScopes(path: "mutations.deleteApplication")
}


type Query {
    runtimes(filter: [LabelFilter!], first: Int = 100, after: PageCursor): RuntimePage! @secureWithScopes(path: "queries.runtimes")
    runtime(id: ID!): Runtime @secureWithScopes(path: "queries.runtime")
}

type Application {
    id: ID! 
    name: String!
    description: String
    labels(key: String): Labels!
    status: ApplicationStatus!
    webhooks: [Webhook!]! @secureWithScopes(path: "types.Application.webhooks")
    healthCheckURL: String
    apis(group: String, first: Int = 100, after: PageCursor): APIDefinitionPage! @secureWithScopes(path: "types.Application.apis")
    eventAPIs(group: String, first: Int = 100, after: PageCursor): EventAPIDefinitionPage! @secureWithScopes(path: "types.Application.eventAPIs")
    documents(first: Int = 100, after: PageCursor): DocumentPage! @secureWithScopes(path: "types.Application.documents")
}
```

Instead of defining manually these directives in GraphQL schema, we can automate it using [gqlgen](https://gqlgen.com/reference/plugins/) plugins.

The `path` parameter specifies the path in YAML file (ConfigMap) with required scopes for a given resolver. For example:
```yaml
queries:
    runtimes: "runtime:view"
    runtime: "runtime:view"
mutations:
    createApplication: "application:admin"
    updateApplication: "application:admin"
    deleteApplication: "application:admin"
types:
    Application:
        webhooks: "webhook:view"
        apis: "api:view"
        eventAPIs: "eventapi:view"
        documents: "document:view"
```

The actual scopes will be defined later.

#### Limiting Application/Runtime modifications

Application/Runtime shouldn't be able to modify other Applications or Runtimes. In future, to limit the functionality, we will introduce another GraphQL directive.

```graphql
type Mutation {
    updateApplication(id: ID!, in: ApplicationInput!): Application! @secureWithScopes(path: "mutations.updateApplication") @limitModifications(type: APPLICATION, idParamName: "id")
}
```

The `limitModifications` mutation compares ID provided for the `updateApplication` mutation with Application ID saved in the context by Tenant Mapping Handler.

## Authentication flows

Each authentication flow is handled on a separate host via different VirtualService, as currently OathKeeper doesn't support certificates and multiple `Bearer` authenticators.

### OAuth 2.0 Access Token

**Used by:** Integration System / Application / Runtime

There are two ways of creating a `client_id` and `client_secret` pair in the Hydra, using Hydra's [oauth client](https://github.com/kyma-project/kyma/blob/ab3d8878d013f8cc34c3f549dfa2f50f06502f14/docs/security/03-06-oauth2-server.md#register-an-oauth2-client) or [simple POST request](https://github.com/kyma-incubator/examples/tree/master/ory-hydra/scenarios/client-credentials#setup-an-oauth2-client).

**Obtaining token:**

1. Runtime/Application/IntegrationSystem requests `client_id` and `client_credentials` pair from Director by separate GraphQL mutation. Director generates the pair, registers it in Hydra with proper scopes (defined by object type) and writes it in database.
1. Runtime/Application/IntegrationSystem calls Hydra with encoded credentials (`client_id` is the ID of `SystemAuth` entry related to given Runtime/Application/IntegrationSystem) and requested scopes.
1. If the requested scopes are valid, Runtime/Application/IntegrationSystem receives in response an access token, otherwise receives an error.

**Request flow:**

1. Authenticator calls Hydra for introspection of the token.
1. If the token is valid, OathKeeper sends the request to Hydrator. 
1. Hydrator calls Tenant Mapping Handler hosted by `Director` to get `tenant` based on a `client_id` (`client_id` is the ID of `SystemAuth` entry related to given Runtime/Application/IntegrationSystem) .
1. Hydrator passes response to ID_Token mutator which constructs a JWT token with scopes and `tenant` in the payload.
1. The request is then forwarded to the desired component (such as `Director` or `Connector`) through the `Gateway` component.
 
![Auth](./assets/oauth2-security-diagram.svg)

**Scopes**

In this authentication flow, scopes are read from OAuth 2.0 access token and written directly in output JWT token. Hydra validates if user can request access token with given scopes.

**Proof of concept:** [kyma-incubator/compass#287](https://github.com/kyma-incubator/compass/pull/287)

### JWT token issued by identity service

**Used by:** User

**Obtaining token:**

User logs in to Compass UI 

**Request flow:**

1. Authenticator validates the token using keys provided by identity service. In production environment, tenant **must be** included in token payload. For local development, the `tenant` property is missing from token issued by Dex.
1. If the token is valid, OathKeeper sends the request to Hydrator.
1. Hydrator calls Tenant Mapping Handler hosted by `Director`, which, in production environment, returns **the same** authentication session (as the `tenant` is already in place). For local development, user `tenant` and `scopes` are loaded from ConfigMap, where static `user - tenant and scopes` mapping is done.
1. Hydrator passes response to ID_Token mutator which constructs a JWT token with scopes and `tenant` in the payload.
1. The request is then forwarded to the desired component (such as `Director` or `Connector`) through the `Gateway` component.
 
![Auth](./assets/dex-security-diagram.svg)

**Scopes**

For local development, user scopes are loaded from ConfigMap, where static `user - tenant and scopes` mapping is done.

**Example ConfigMap for local development**

```yaml
admin@kyma.cx:
    tenant: edf2e0c0-58b1-45c6-b345-fabc9774600c
    scopes:
        - application:admin
        - runtime:admin
foo@bar.com:
    tenant: c862d791-2735-4ffb-ae2d-3ace408d6cff
    scopes:
        - application:view
        - runtime:view
```

### Client certificates

**Used by:** Runtime/Application

**Compass Connector flow:**

1. Runtime/Application makes a call to the Connector to the certificate-secured subdomain. 
2. Istio verifies the client certificate. If the certificate is invalid, Istio rejects the request. 
3. The certificate info (subject and certificate hash) is added to the `Certificate-Data` header.
4. The OathKeeper uses the Certificate Resolver as a mutator, which turns the `Certificate-Data` header into the `Client-Certificate-Hash` header and the `Client-Id-From-Certificate` header. If the certificate has been revoked, the two headers are empty. 
5. The request is forwarded to the Connector through the Compass Gateway.

![Auth](./assets/certificate-security-diagram-connector.svg)

**Compass Director Flow:**

1. Runtime/Application makes a call to the Director to the certificate-secured subdomain. 
2. Istio verifies the client certificate. If the certificate is invalid, Istio rejects the request. 
3. The certificate info (subject and certificate hash) is added to the `Certificate-Data` header.
4. The OathKeeper uses the Certificate Resolver as a mutator, which turns the `Certificate-Data` header into the `Client-Certificate-Hash` header and the `Client-Id-From-Certificate` header. If the certificate has been revoked, the two headers are empty. 
5. The call is then proxied to the Tenant mapping handler, where the `client_id` is mapped onto the `tenant` and returned to the OathKeeper. If the Common Name is invalid, the `tenant` will be empty. 
6. Hydrator passes the response to ID_Token mutator which constructs a JWT token with scopes and tenant in the payload.
7. The OathKeeper proxies the request further to the Compass Gateway.
8. The request is forwarded to the Director. 

![Auth](./assets/certificate-security-diagram-director.svg)

**Scopes**

Scopes are added to the authentication session in Tenant Mapping Handler. The handler gets not only `tenant`, but also `scopes`, which are fixed regarding of type of the object (Application / Runtime). Application and Runtime are always have the same scopes defined.
