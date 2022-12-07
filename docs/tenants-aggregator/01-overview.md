# Tenants Aggregator

The purpose of the Tenants Aggregator component is to manage the tenants in the Compass database. It periodically synchronizes the tenants in the database with an external tenants registry. Additionally, it keeps track of events that involve tenants, which might be new to Compass.

## Tenancy Model

Compass supports the following tenant types that can be used for tenancy context:

- Customer
- Account
- Subaccount

The different tenants have hierarchical structure. That is, all resources that are available in a subordinate tenant are also visible in the
superordinate tenant.

![hierarchical tenants view](./assets/tenancy-model.svg)

Compass stores external ID, external name, and tenant subdomains and regions in case of multi-region scenarios.

## Aggregation Capabilities

The Tenants Aggregator provides the following ways of aggregation:

- Periodic synchronization - Tenants are synchronized periodically against an external system that exposes APIs for fetching tenant-related events, such as, created, updated, deleted, or moved (from one superordinate tenant to another).
- _Tenant On-demand API_ - An API that is used to fetch a specific tenant of type `subaccount` from an external tenants's registry. If there is a tenant with
  the provided ID, it is stored in the Compass database.
- _Subscription Callback API_ - An API that creates any tenants that are missing in the hierarchy of the tenant from the request. The API is used when the request is for a new subscription.

## Tenant Fetchers

A tenant fetcher is a configurable, periodic, Go routine that runs on the Tenants Aggregator. Its purpose is to
synchronize tenants that are managed by an external system. Note that the external system must expose an API with tenants-related events, for example, when a new tenant is created, updated, deleted, or moved.

### Region-Based Tenancy

In cases when SaaS applications or integration systems from multiple regions use Compass, the supported tenants must
be associated with regions, too. That is, consumers from one region can only interact with tenants from the same
region.

By default, the tenant fetcher is region-based because at least one central region is always available.

The tenant fetcher first tries to fetch the tenant events from the central region, and then, fills out any missing events
with the events from each available additional region.

![regional tenancy](./assets/tenant-regions.svg)

### External Tenant Registry

To integrate with the Tenants Aggregator, an external registry must implement an API schema known as _Tenant Events API_.

#### Authorization

The Tenant Events API uses the OAuth 2.0 client credentials authorization flow or mTLS, and comprises preconfigured trust
for the externally issued client certificate of Compass.

#### Endpoints

The following types of supported endpoints that receive different types of events are available:

- Tenant creation endpoint
- Tenant deletion endpoint
- Tenant update endpoint
- Tenant move endpoint

Every endpoint returns a specific payload and accepts the following query parameters:

- **global.tenantFetchers.*<job_name>*.queryMapping.timestampField** - Specifies a timestamp in Unix time format. That
  is, the date, from which events are fetched.
- **global.tenantFetchers.*<job_name>*.queryMapping.pageNumField** - Specifies the number of the exact page to be fetched. To
  specify a starting page of the query, you can use the parameter **global.tenantFetchers.*<job_name>*.query.startPage**
  to preconfigure a starting page number.
- **global.tenantFetchers.*<job_name>*.queryMapping.pageSizeField** - Specifies the number of results on a single page
  of the response.

#### API Response

Most of the top-level data that is expected by the tenant fetcher is configurable. Consider the following example
response for the tenant creation endpoint:

```json
{
  "events": [
    {
      "eventData": "{\"$id\":\"837d023b-782d-4a97-9d38-fecab47c296a\",\"$name\":\"Tenant 1\",\"$discriminator\":\"default\"}"
    }
  ],
  "totalResults": 27,
  "totalPages": 1
}
```

- The top-level `events` array is configured with: **global.tenantFetchers.*<job_name>*.fieldMapping.tenantEventsField**
- The top-level `totalResults` is configured with: **global.tenantFetchers.*<job_nam>e*.fieldMapping.totalResultsField**
- The top-level `totalPages` is configured with: **global.tenantFetchers.*<job_name>*.fieldMapping.totalPagesField**
- The inner field `eventData` contains the details of an event and it is configured by: **global.tenantFetchers.*<
  job_name>*.fieldMapping.detailsField**. The details field is expected to be either a JSON object or a string
  containing JSON. For example:

```json
{
  "details": {
    "key": "value"
  }
}
```

or

```json
{
  "details": "{\"key\": \"value\"}"
}
```

##### Tenant creation endpoint

When successful, the endpoint returns the following JSON payload:

```json
{
  "events": [
    {
      "eventData": "{\"$id\":\"837d023b-782d-4a97-9d38-fecab47c296a\",\"$name\":\"Tenant 1\",\"$discriminator\":\"default\"}"
    }
  ],
  "totalResults": 27,
  "totalPages": 1
}
```

The **eventData** field contains an escaped JSON string with the following fields that you can configure by using values
overrides:

- `$id` - Specifies a unique tenant ID.
- `$parent_id` - Specifies a unique tenant ID belonging to the superordinate tenant. If the created tenant is a `subaccount`,
  the superordinate tenant must be of type `account`. If the created tenant is of type `account`, the superordinate tenant must be of
  type `customer`.
- `$name` - Specifies the tenant name.
- `$discriminator` - Specifies an optional field that can be used to distinguish different types of tenants.

##### Tenant deletion endpoint

When successful, the endpoint returns the following JSON payload:

```json
{
  "events": [
    {
      "eventData": "{\"$id\":\"837d023b-782d-4a97-9d38-fecab47c296a\",\"$name\":\"Tenant 1\"}"
    }
  ],
  "totalResults": 27,
  "totalPages": 1
}
```

The **eventData** field contains an escaped JSON string with the following fields that you can configure by using values
overrides:

- `$id` - Specifies a unique tenant ID.
- `$name` - Specifies the tenant name.

##### Tenant update endpoint

When successful, the endpoint returns the following JSON payload:

```json
{
  "events": [
    {
      "eventData": "{\"$id\":\"837d023b-782d-4a97-9d38-fecab47c296a\",\"$name\":\"Tenant 1\"}"
    }
  ],
  "totalResults": 27,
  "totalPages": 1
}
```

The **eventData** field contains an escaped JSON string with the following fields that you can configure by using values
overrides:

- `$id` - Specifies a unique tenant ID.
- `$name` - Specifies the tenant name.

##### Tenant move endpoint

When successful, the endpoint returns the following JSON payload:

```json
{
  "events": [
    {
      "eventData": "{\"$id\":\"837d023b-782d-4a97-9d38-fecab47c296a\",\"$sourceParentTenantID\":\"Tenant 1\",\"$targetParentTenantID\":\"Tenant 2\"}"
    }
  ],
  "totalResults": 27,
  "totalPages": 1
}
```

The **eventData** field contains an escaped JSON string with the following fields that you can configure by using values
overrides:

- `$id` - Specifies a unique tenant ID.
- `$sourceParentTenantID` - Specifies the previous of superordinate object of the tenant.
- `$targetParentTenantID` - Specifies the current or the new superodinate object of the tenant.

## Tenant On-demand API

The Tenant On-demand API is called by Compass Director for requests involving subaccounts. That call is a prerequisite for those requests because the subaccount tenant might be missing from the Compass DB.
It might be missing because the tenant fetcher's Go routine for that type of tenant has not been run yet or no subscription
callbacks have been received for the tenant.

Currently there are the following scenarios when a brand new subaccount tenant might be used:

- When a Runtime is being created in the scope of a subaccount tenant.
- When a subaccount tenant is assigned to a scenario.
- When an Automatic Scenario Assignment is created for a subaccount tenant.
- When the Hydrator tries to find a tenant present in the request context.

The Fetch On-demand flow is described in the diagram below:
![tenant on-demand flow](./assets/tenant-on-demand.svg)

### API Security
The On-demand API is used only internally by Director and Hydrator, hence it uses the Internal Authentication mechanism. You can learn more about it [here](https://github.com/kyma-incubator/compass/blob/main/docs/compass/03-01-security.md#internal-authentication).

## Subscription Callback API

The Tenants Aggregator is the entry point for tenant subscriptions. It provides an endpoint that can be registered as
a subscription callback in an external system. This endpoint is called when:
* a tenant is created
* a consumer tenant is **subscribed** to a provider tenant
* a consumer tenant is **unsubscribed** from a provider tenant

### Subscription and Tenant Creation
```
PUT /v1/regional/{region}/callback/{tenantId}
```

Example subscription request body
```
{
	"subscriptionAppId": "compass",
	"subscribedCrmId": "31c48eec-182b-4fa4-847f-0697cb91fc76",
	"accountID": "21c111ec-182b-4tr4-8trf-0i8scb91fc75",
	"subdomain": "desitest",
	"subaccountId": "1573a360-a409-4183-9d38-0e1237cfc191"
}
```

When a call is received at the endpoint, it tries to create all tenants from the tenant hierarchy of the tenant for
that request. Then, if the callback is a subscription callback, it calls the Director to subscribe the tenant.
The subscription logic is handled in Director.

The Tenants Aggregator runs the following Director mutation:

```graphql
subscribeTenant(providerID: String!, subaccountID: String!, providerSubaccountID: String!, consumerTenantID: String!, region: String!, subscriptionAppName: String!): Boolean! @hasScopes(path: "graphql.mutation.subscribeTenant")
```

It distinguishes subscription callbacks and normal tenant creation callbacks by checking the `subscribedSaaSApplication` property in the request
payload.

Compass can be considered as an SaaS application that a tenant can subscribe to. The only required action from Compass side
is to store that tenant.

If the Subscription Callback API is called when a tenant is subscribed to another SaaS application or a Runtime, then it must also provide the subscription tenant (provider tenant) access to the subscriber tenant (consumer tenant). 

### Subscription and Tenant Decomissioning
```
DELETE /v1/regional/{region}/callback/{tenantId}
```
The Tenants Aggregator delegates the unsubscribe logic to Director by calling the following mutation:
```graphql
unsubscribeTenant(providerID: String!, subaccountID: String!, providerSubaccountID: String!, consumerTenantID: String!, region: String!): Boolean! @hasScopes(path: "graphql.mutation.unsubscribeTenant")
```

No tenants are deleted via this call. The only flow where tenants are actually removed from the Compass DB is when the Tenant Fetcher Routine is run and there is a `Delete` event for the tenant.

### API Security
The Subscription API is called from an external system. The Security configuration is managed as ORY Oathkeeper rule [here](https://github.com/kyma-incubator/compass/blob/main/chart/compass/charts/gateway/templates/oathkeeper-authenticator-rules.yaml), and the current scenario uses the JWT Gateway endpoint.
The validation on the Tenants Aggregator side is only for one scope in the token - `Callback`.
