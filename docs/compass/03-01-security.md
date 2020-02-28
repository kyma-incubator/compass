# Security

Compass is secured with ORY [Hydra](https://www.ory.sh/hydra/) and [OathKeeper](https://github.com/ory/oathkeeper). There are the following ways of authentication: 
 - OAuth 2.0
 - Client certificates (mTLS)
 - JWT token issued by identity service
 - One-time token

 To achieve that, first, we need to integrate Hydra and OathKeeper into Compass. We also need to implement additional supporting components to make our scenarios valid.

## Architecture

The following diagram represents the architecture of the security in Compass:

![](./assets/security-architecture.svg)

### Tenant Mapping Handler

It is an OathKeeper [hydrator](https://github.com/ory/docs/blob/525608c65694539384b785355d293bc0ad00da27/docs/oathkeeper/pipeline/mutator.md#hydrator) handler responsible for mapping authentication session data to the tenant. It is built into the Director itself, as it uses the same database. It is implemented as a separate endpoint, such as `/tenant-mapping`.

To unify the approach for mapping, we introduce `authorization_id`, widely used by multiple authentication flows.
The `authorization_id` is equal to:
- `client_id` in the OAuth 2.0 authentication flow
- `username` in the Basic authentication flow
- Common Name (CN) in the certificates authentication flow
- Client ID in the one-time token authentication flow

While generating the one-time token, the `client_id`/`client_secret` pair, or basic authentication details for Runtime/Application/Integration System (using proper GraphQL mutation on the Director), an entry in the `system_auths` table in the Director database is created. The `system_auths` table is used for tenant mapping.

In the certificates authentication flow, Tenant Mapping Handler puts fixed `scopes` into an authentication session. The `scopes` are fixed in code, and they depend on the type of the object (Application/Runtime/Integration System). In the future we may introduce another database table for storing generic Application/Runtime/Integration System scopes.

In JWT token from identity service flow, for local development, user `tenant` and `scopes` are loaded from ConfigMap for given user (email), where static `user: tenant and scopes` mapping is done.

#### `system_auths` table

The table is used by the Director and Tenant Mapping Handler. It contains the following fields:
- `id`, which is the `authorization_id`
- `tenant` (optional field - used for Application and Runtime, not used for Integration System)
- `app_id` foreign key of type UUID
- `runtime_id` foreign key of type UUID
- `integration_system_id` foreign key of type UUID
- `value` of type JSON, with authentication details, such as `client_id/client_secret` in the OAuth 2.0 authentication flow, or `username/password` in the Basic authentication flow. In the case of the certificates flow it is empty.

### GraphQL security

The Gateway passes the request to Compass GraphQL services, such as the Director or the Connector. Additionally, the request contains authentication data. In the Director it is a JWT token, in the Connector it is the `Client-Id-From-Token` header or the `Client-Id-From-Certificate` and `Client-Certificate-Hash` headers. The GraphQL components have authentication middleware and GraphQL [directives](https://graphql.org/learn/queries/#directives) set up for all GraphQL operations (and some specific type fields, if necessary). 

#### HTTP middleware

In GraphQL servers, such as the Director or the Connector, there is an HTTP authentication middleware set up. In the Director, it validates and decodes the JWT token, and puts user scopes and the tenant in the request context
(`context.Context`). In the Connector, it verifies the `Client-Id-From-Token` header or the `Client-Id-From-Certificate` and `Client-Certificate-Hash` headers.

![](./assets/graphql-security.svg)

#### GraphQL Directives

When GraphQL operation is processed, an authorization directive is triggered, before actual GraphQL resolver. It checks if the client has required scopes to perform the requested operation. To avoid defining permissions statically in the GraphQL schema, a YAML file is loaded with needed requests. In fact, it is a ConfigMap injected to the Director/Connector.

The following example illustrates how we can implement dynamic comparison between required scopes and request scopes for example mutations, queries, and type fields:

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

Instead of manually defining these directives in the GraphQL schema, we can automate it using [gqlgen](https://gqlgen.com/reference/plugins/) plugins.

The `path` parameter specifies the path in the YAML file (ConfigMap) with required scopes for a given resolver. For example:
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

Application/Runtime shouldn't be able to modify other Applications or Runtimes. In the future, to limit the functionality, we will introduce another GraphQL directive.

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

1. Runtime/Application/IntegrationSystem requests the `client_id` and `client_credentials` pair from the Director by a separate GraphQL mutation. the Director generates the pair, registers it in Hydra with proper scopes (defined by object type), and writes it to the database.
1. Runtime/Application/IntegrationSystem calls Hydra with encoded credentials (`client_id` is the ID of the `SystemAuth` entry related to the given Runtime/Application/IntegrationSystem) and requested scopes.
1. If the requested scopes are valid, Runtime/Application/IntegrationSystem receives an access token in response. Otherwise, it receives an error.

**Request flow:**

1. Authenticator calls Hydra for introspection of the token.
1. If the token is valid, OathKeeper sends the request to Hydrator.
1. Hydrator calls Tenant Mapping Handler hosted by `Director` to get `tenant` based on a `client_id` (`client_id` is the ID of `SystemAuth` entry related to the given Runtime/Application/IntegrationSystem) .
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

1. Authenticator validates the token using keys provided by the identity service. In the production environment, the tenant **must be** included in the token payload. For local development, the `tenant` property is missing from the token issued by Dex.
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
5. The call is then proxied to Tenant Mapping Handler, where the `client_id` is mapped onto the `tenant` and returned to the OathKeeper. If the Common Name is invalid, the `tenant` will be empty.
6. Hydrator passes the response to ID_Token mutator which constructs a JWT token with scopes and tenant in the payload.
7. The OathKeeper proxies the request further to the Compass Gateway.
8. The request is forwarded to the Director.

![Auth](./assets/certificate-security-diagram-director.svg)

**Scopes**

Scopes are added to the authentication session in Tenant Mapping Handler. The handler gets not only `tenant`, but also `scopes`, which are defined per object type (Application / Runtime). 

### One-time token

**Used by:** Runtime/Application

**Connector flow:**

1. Runtime/Application makes a call to the Connector's internal API.
1. The OathKeeper uses the Token Resolver as a mutator. The Token Resolver extracts the Client ID from the one-time token's `Connector-Token` header or from the `token` query parameter, and writes it to the `Client-Id-From-Token` header. In the case of failure, the header is empty.
1. The OathKeeper proxies the request further to the Compass Gateway.
1. The request is forwarded to the Connector.

![Auth](./assets/token-security-diagram-connector.svg)

**Director Flow:**

1. Runtime/Application makes a call to the Director.
1. The OathKeeper uses the Token Resolver as a mutator. The Client ID is extracted from the one-time token's `Connector-Token` header or from the `token` query parameter, and is then written to the `Client-Id-From-Token` header. In the case of failure, the header is empty.
1. The call is then proxied to the Tenant Mapping Handler mutator, where the Client ID is mapped onto the `tenant` and returned to the OathKeeper. 
1. Hydrator passes the response to ID_Token mutator which constructs a JWT token with scopes and tenant in the payload.
1. The OathKeeper proxies the request further to the Compass Gateway.
1. The request is forwarded to the Director.

![Auth](./assets/token-security-diagram-director.svg)

**Scopes**

Scopes are added to the authentication session in Tenant Mapping Handler. The handler gets not only `tenant`, but also `scopes`, which are defined per object type (Application/Runtime).
