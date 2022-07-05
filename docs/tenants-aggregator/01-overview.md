# Tenants Aggregator

The Tenants Aggregator is responsible for maintaining the currently available tenants in the Compass DB. It takes care
of periodically synchronizing the tenants with an external tenants registry. It also listens for events involving
tenants that might be still unknown to Compass.

## Tenancy Model

Compass supports the following tenant types that can be used for tenancy context:

- Customer
- Account
- Subaccount

The different tenants have hierarchical structure and all resources available in a child tenant, are also visible in the
parent tenant.

![hierarchical tenants view](./assets/tenancy-model.svg)

Along with external ID and external name, Compass stores tenant subdomains and regions in case of multi-region
scenarios.

## Aggregation Capabilities

The _Tenants Aggregator_ provides the following ways of aggregation:

- Periodic synchronization of tenants against an external system that exposes APIs for fetching events related to
  tenants - created, updated, deleted, or moved (moved from one parent tenant to another).
- _On-demand API_ to fetch a specific tenant of type `subaccount` from an external tenants registry - if a tenant with
  the provided ID exist there, it will be stored in the Compass DB.
- _Subscription Callback API_ that creates any missing tenants from the tenant hierarchy of the tenant provided in the
  request, if that request is for a new subscription.

## Tenant Fetchers

A "Tenant Fetcher" is a configurable periodic Go routine that runs on the _Tenants Aggregator_. It is responsible for
synchronizing tenants that are managed by an external system.

That external system should expose an API with events related to the tenants - when a new tenant is created, updated,
deleted or moved.

### Region-based tenancy

In case where SaaS applications or Integration Systems from multiple regions use Compass, the supported tenants should
also be associated with regions. That way, consumers from one region can only interact with tenants from the same
region.

The Tenant Fetcher is region-based by default, since you'll always have at least one central region available.

The Tenant Fetcher first tries to fetch the tenant events from the central region, and then fills out any missing events
with the events from each available additional region.

![regional tenancy](./assets/tenant-regions.svg)

### External Tenant Registry

This section describes the API schema that an External Registry should implement in order to integrate with the _Tenants
Aggregator_. We will call that API _"Tenant Events API"_.

#### Authorization

Tenant Events API uses the OAuth 2.0 client credentials authorization flow or mTLS, and comprises preconfigured trust
for the externally issued client certificate of Compass.

#### Endpoints

There are three types of supported endpoints that receive different types of events:

- Tenant creation endpoint
- Tenant deletion endpoint
- Tenant update endpoint
- Tenant move endpoint - for tenants which have been moved from one parent tenant to another

Every endpoint returns a specific payload and accepts the following query parameters:

- **global.tenantFetchers.*<job_name>*.queryMapping.timestampField** - Specifies a timestamp in Unix time format. That
  is, the date, from which events are fetched.
- **global.tenantFetchers.*<job_name>*.queryMapping.pageNumField** - Specifies the number of the page to be fetched. To
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
- The inner field `eventData` contains the details of an event, and it is configured by: **global.tenantFetchers.*<
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
- `$parent_id` - Specifies a unique tenant ID belonging to the parent tenant. If the created tenant is a `subaccount`,
  the parent tenant must be of type `account`. If the created tenant is of type `account`, the parent tenant must be of
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
- `$sourceParentTenantID` - Specifies the tenant's old parent.
- `$targetParentTenantID` - Specifies the tenant's current/new parent.

## On-demand API

The _Tenant On-demand API_ is called in cases where the tenant from the request context is still not known to Compass.
That might happen because the _Tenant Fetcher_ Go routine for that type of tenant has yet to be run, or no subscription
callbacks were received for that tenant.

We currently have a couple of scenarios where a tenant that might be still unknown is used:

- When a Runtime is being created in the scope of a subaccount tenant.
- When a subaccount tenant is assigned to a scenario.
- When an Automatic Scenario Assignment is created for a subaccount tenant.
- When the Hydrator tries to find a tenant present in the request context.

The "Fetch On-demand" flow is described in the diagram below:
![tenant on-demand flow](./assets/tenant-on-demand.svg)

## Subscriptions API

The entry point for tenant subscriptions is the _Tenants Aggregator_ - it provides an endpoint that can be registered as
a callback in an external system. That endpoint will be called when a tenant is created, or a consumer tenant is
subscribed to a provider tenant.

When a call is received at that endpoint, we first try to create all tenants from the tenant hierarchy of the tenant for
that request. Then, if the callback is a subscription callback, we call Director to subscribe or unsubscribe a tenant.
The subscription logic is handled in Director. For more details see the [subscription document]().

The _Tenants Aggregator_ runs the following _Director_ mutations:

```graphql
type Mutation {
    subscribeTenant(providerID: String!, subaccountID: String!, providerSubaccountID: String!, consumerTenantID: String!, region: String!, subscriptionAppName: String!): Boolean! @hasScopes(path: "graphql.mutation.subscribeTenant")
    unsubscribeTenant(providerID: String!, subaccountID: String!, providerSubaccountID: String!, consumerTenantID: String!, region: String!): Boolean! @hasScopes(path: "graphql.mutation.unsubscribeTenant")
}
```

We distinguish subscription callbacks and normal tenant creation callbacks by checking one property in the request
payload - `subscribedSaaSApplication`.

Compass can be viewed as a SaaS application that a tenant can subscribe to. The only required action from Compass side
is to store that tenant.

If the Subscription Callback API is called when a tenant is subscribed to another SaaS Application or a Runtime, then we
should provide the subscription tenant (provider tenant) access to the subscriber tenant (consumer tenant). 