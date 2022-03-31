# Security

There are multiple ways you can access Compass from security perspective - it supports the following ways of authentication: 
 - OAuth 2.0
 - Client certificates (mTLS)
    - issued by the Compass Connector
    - issued by external certificate authority
 - JWT token issued by identity service
 - One-time token
 
 There are also **two consumer-provider** flows where an external multi-tenant system wants to manage resources (e.g. bundle instance auths) on behalf of a tenant present in Compass. :
 - SaaS applications, which are modeled like runtimes, can access Compass resources with an externally-issued certificate, and a JWT where the provider tenant is present.
 - Integration Systems which are configured with externally-issued certificate, can once again consume resources from tenants which are provided as a `Tenant` header.

Compass is integrated with [ORY Hydra](https://www.ory.sh/hydra/) and [ORY OathKeeper](https://github.com/ory/oathkeeper).

## Architecture

The following diagram represents the architecture of the security in Compass:

![](./assets/security-architecture.svg)

### Tenant Mapping Handler

The Tenant Mapping Handler is an OathKeeper [hydrator](https://github.com/ory/docs/blob/525608c65694539384b785355d293bc0ad00da27/docs/oathkeeper/pipeline/mutator.md#hydrator) handler responsible for mapping authentication session data to the tenant. It is built into the Director itself, as it uses the same database. It is implemented as a separate endpoint, such as `/tenant-mapping`.

To unify the approach for mapping, we introduce Authorization ID - `auth_id`, widely used by multiple authentication flows. Each authentication flow has it's own context provider, which takes care of extracting tenant information, and granted scopes.

The `auth_id` is equal to:
- `client_id` in the OAuth 2.0 authentication flow
- Client ID in the one-time token authentication flow
- Common Name (CN) in the Connector-issued certificates authentication flow
- External Tenant ID of type `Subaccount` for externally-issued certificates flow extracted from the certificate's organizational unit (OU) by default

While generating one of the following - a one-time token, a `client_id`/`client_secret` pair, or a client certificate issued by Connector for Runtime/Application/Integration System (using proper GraphQL mutation on the Director) - an entry in the `system_auths` table in the Director database is created. The `system_auths` table is used for tenant mapping.

In the certificates' authentication flow (both Connector-issued and externally-issued), Tenant Mapping Handler puts fixed `scopes` into an authentication session. The `scopes` are fixed in code, and they depend on the type of the caller object (Application/Runtime/Integration System). They are currently listed [here](https://github.com/kyma-incubator/compass/blob/1105695797f74eba8d8a86ee3d4d65809ef6abb7/chart/compass/charts/director/config.yaml#L120). In the future we may introduce another database table for storing generic Application/Runtime/Integration System scopes.

In the identity service flow, where JWT token is used, from the `scopes` are loaded from a ConfigMap, where a static `user_group: scopes` mapping is done. The user group is present in the JWT.

#### `system_auths` table

The table is used by the Director and Tenant Mapping Handler. It contains the following fields:
- `id`, which is the `authorization_id`
- `tenant` (optional field - used for Application and Runtime, not used for Integration System)
- `app_id` foreign key of type UUID
- `runtime_id` foreign key of type UUID
- `integration_system_id` foreign key of type UUID
- `value` of type JSON, with authentication details, such as `client_id/client_secret` in the OAuth 2.0 authentication flow, Common Name in case of the certificates flow, and token in the one-time-token flow.

#### Custom Authenticator
One can configure custom JWT-based authentication for Compass with trusted issuers, and locations for tenant and scopes in the token. The flow is the following:
- The authentication-mapping handler hydrator checks for trusted issuer, and enriches  and then provides the tenant-mapping handler with locations for tenant, scopes and clientID
- The tenant-mapping handler extracts them from the token's claims and verifies the tenant exists

#### Consumer-Provider Flows
As mentioned above, there are use cases where a user might want to manage Compass resources from a different multi-tenant system which shares the same tenancy model, and is aware of the external tenant IDs, which Compass also uses.
Compass should be able to trust those systems to manage resources on behalf of the user. There are two flavors of this flow - the multi-tenant system can be either represented as an Integration System, or as a Runtime.

Both cases will match two context providers which should be used in pair, one of which will be the externally-issued certificate context provider - it will extract the consumer tenant ID from the certificate.

In the integration system case, the provider tenant is specified in the `Tenant` header. The scopes which will be granted to the request are the ones assigned to integration systems by default (will be granted by the context provider mentioned above).

In the runtime case, the second context provider that will match is a custom authenticator one, where the provider tenant is part of a JWT, along with the scopes that will be granted - we take the scopes from this context provider in this case.
Later, Compass checks if the consumer tenant has access to the provider tenant. More details can be found in the *Authentication Flows* fragment below.

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

There are two ways of creating a `client_id` and `client_secret` pair in the Hydra, using Hydra's [oauth client](https://github.com/kyma-project/kyma/blob/ab3d8878d013f8cc34c3f549dfa2f50f06502f14/docs/security/03-06-oauth2-server.md#register-an-oauth2-client) or [simple POST request](https://github.com/kyma-incubator/examples/tree/main/ory-hydra/scenarios/client-credentials#setup-an-oauth2-client).

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

1. Authenticator validates the token using keys provided by the identity service. 
1. If the token is valid, OathKeeper sends the request to Hydrator.
1. Hydrator calls Tenant Mapping Handler hosted by `Director`, which, in production environment, returns **the same** authentication session (as the `tenant` is already in place) even for local development. See the [installation document](https://github.com/kyma-incubator/compass/blob/main/docs/compass/04-01-installation.md#local-minikube-installation) for more details on how to configure OIDC Authentication Server.
1. Hydrator passes response to ID_Token mutator which constructs a JWT token with scopes and `tenant` in the payload.
1. The request is then forwarded to the desired component (such as `Director` or `Connector`) through the `Gateway` component.

![Auth](./assets/dex-security-diagram.svg)

**Scopes**

Scopes are loaded from a ConfigMap, where static `user group -  scopes` mapping is done.

**Example ConfigMap**

```yaml
kind: ConfigMap
apiVersion: v1
metadata:
   name: compass-director-static-groups
   namespace: compass-system
data:
   static-groups.yaml: |
      - groupname: "application-superadmin"
        scopes:
          - "application:read"
          - "application:write"
          - "application.auths:read"
          - "application.webhooks:read"
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

### Externally-issued client certificates

**Used by:** Runtime/Integration System

**Compass Director Flow:**

1. Runtime/Application makes a call to the Director to the externally-issued certificate-secured subdomain.
2. Istio verifies the client certificate. If the certificate is invalid, Istio rejects the request.
3. The certificate info (subject and certificate hash) is added to the `Certificate-Data` header.
4. The OathKeeper uses the Certificate Resolver as a mutator, which turns the `Certificate-Data` header into the `Client-Certificate-Hash` header and the `Client-Id-From-Certificate` header. If the certificate is expired, the two headers are empty. Additionally, an `extra` field is added to the *Auth Session*, if the subject matches one of the subjects in a predefined configuration. The extra will contain a contains consumer type (integration system), access levels (contains a set of tenant types, which the consumer can access, e.g. `account` only), and optional internal consumer ID, which can be to the GUID of an existing integration system.
5. the Certificate Resolver also sets the Authentication ID (`auth_id`) to one of the OUs in the subject. If there are many OUs, Connector can be configured which ones to skip. The auth ID should represent the external ID of a tenant of type `subaccount`
6. The call is then proxied to the Tenant Mapping Handler, where:
   1. In case of static match to an externally issued certificate, the `Tenant` header is mapped to a `tenant`
   2. Otherwise, the `auth_id` is mapped onto the `tenant`
   3. If the tenant does not exist, an empty tenant info is returned.
7. Hydrator passes the response to ID_Token mutator which constructs a JWT token with scopes and tenant in the payload.
8. The OathKeeper proxies the request further to the Compass Gateway.
9. The request is forwarded to the Director.

The diagram is the same as the one for client certificate above.

**Scopes**

Scopes are added to the authentication session in Tenant Mapping Handler. The handler gets not only `tenant`, but also `scopes`, which are defined per object type (Application / Runtime). The default type is Runtime.

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

### Consumer-provider flow

**Used by:** Runtime/Integration System

##### Integration Systems
Currently, an integration system is any multi-tenant application, which has access to all tenants available in Compass.
The Integration Systems can use OAuth credentials for authentication (created via the `requestClientCredentialsForIntegrationSystem` mutation), or client certificate, issued by an external issuer, which is trusted by Compass.

The second authentication mechanism is plugged into Compass Connector - it can be configured to trust a specific certificate subject, and based on that subject, to set the Consumer type to _Integration System_, and grant it access to a set of tenant types - it is a bit more restrictive than the one that uses OAuth client.
The provider tenant is specified in the `Tenant` header.

**Compass Director Flow:**

The flow is basically the one described above for externally issued certificate. The only difference is that the Tenant Mapping handler will match two object context providers, hence the tenants in this request will be 4 in total - internal and external provider tenants, and internal and external consumer tenants.
If the provider tenant has access to the type of tenant the consumer belongs to, then the request is authorized successfully.

**Scopes**
Since there will be two context providers, there will also be two pair of scopes sets - one from the matched external certificate context provider, and one from the access level context provider.
In this case, the scopes from the external certificate context provider (for consumer type Integration System) will be taken into account.

##### Runtimes
As Integration Systems are usually too powerful, and we should be careful who gets Integration System access, a more flexible and secured way of achieving the same scenarios was introduced.
First, the multi-tenant applications should be registered as a "special" runtime in Compass, which contains a special label, which Compass uses for distinguishing multi-tenant and ordinary runtimes.

When the multi-tenant system is represented as a Runtime, its tenant access is managed from an outside service, which communicates with the Tenant Fetcher deployment, and grants the provider tenant (the tenant where the multi-tenant runtime is registered) access to resources in the consumer tenant, which will be of type `subaccount`.

Once the trust between the tenants is established, the user can access any resource which resides in in its tenant from the external multi-tenant system.

**Compass Director Flow:**

_Not productive yet_

**Compass ORD Service Flow:**

_Used in production_

This flow is also almost the same as the externally issued certificate one.
Just like above, two context providers are matched - the externally issued certificate one, and a custom authentication one with JWT.
Compass will validate that the consumer tenant has access to the provider tenant.

**Scopes**
The first context provider will return Runtime scopes, and the second one will return the scopes present in the JWT. We should use the ones from the token.

## Internal Authentication

Compass has a separate security layer to take care of internal component communication (e.g. from ORY Oathkeeper to Director). 

It is based on Istio `RequestAuthentications` and `AuthorizationPolicies` that require service account token from the caller as an additional X-Authorization header. The service account in question is the one bound to the source workload.
The `RequestAuthentication` verifies that the token is correctly signed by the Kubernetes JWKS. 
The `AuthorizationPolicy` makes sure that routes such as `/healthz` are accessible without any authentication, and all other routes require properly signed tokens.

Workloads that previously required some form of authentication, e.g. ORY Oathkeeper tokens for Director, or one-time tokens for Connector, still require them as a separate header.

![](./assets/internal-auth.svg)

* Flows going through ORY Oathkeeper pass through an Envoy filter (matching outbound requests) that auto-injects the service account token in the request. This is secure because going through ORY successfully means that the request already has valid authn/authz.
* Internal calls to Director require an ORY Oathkeeper token on top of the service account token. A new ORY Oathkeeper gateway (`gateway-int`) and a respective rule that accepts a service account token as Authorization header and produces an ORY Token with all the required scopes were introduced to simplify Director clients (such as Operations Controller). On top of that, the Oathkeeper Envoy filter injects again the service account token as `X-Authorization` header when the call passes through Oathkeeper, so that the Istio Policy will pass successfully.
