# Tenant Fetcher

## Overview

This application fetches events containing information about created, moved, and deleted tenants. Then, based on the event type, it performs the corresponding create, update, or delete operation on the tenant event in the Director database.

## Details

This section describes the API schema that a server must implement to integrate with the Tenant Fetcher.

### Authorization

Tenant Events API uses the OAuth 2.0 client credentials authorization flow or mTLS, and comprises preconfigured trust for the externally issued client certificate of Compass.

### Endpoints

There are three types of supported endpoints that receive different types of events:
- Tenant creation endpoint
- Tenant deletion endpoint
- Tenant update endpoint

Every endpoint returns a specific payload and accepts the following query parameters:
- **global.tenantFetchers.*<job_name>*.queryMapping.timestampField** - Specifies a timestamp in Unix time format. That is, the date, from which events are fetched.
- **global.tenantFetchers.*<job_name>*.queryMapping.pageNumField** - Specifies the number of the page to be fetched. To specify a starting page of the query, you can use the parameter **global.tenantFetchers.*<job_name>*.query.startPage** to preconfigure a starting page number.
- **global.tenantFetchers.*<job_name>*.queryMapping.pageSizeField** - Specifies the number of results on a single page of the response.

#### Response

Most of the top-level data that is expected by the tenant fetcher is configurable. Consider the following example response for the tenant creation endpoint:

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
- The inner field `eventData` contains the details of an event and it is configured by: **global.tenantFetchers.*<job_name>*.fieldMapping.detailsField**. The details field is expected to be either a JSON object or a string containing JSON. For example:

```json
{
  "details": {
    "key": "value"
  }
}

or

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

The **eventData** field contains an escaped JSON string with the following fields that you can configure by using values overrides:
- `$id` - Specifies a unique tenant ID.
- `$parent_id` - Specifies a unique tenant ID belonging to the parent tenant. If the created tenant is a `subaccount`, the parent tenant must be of type `account`. If the created tenant is of type `account`, the parent tenant must be of type `customer`. 
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

The **eventData** field contains an escaped JSON string with the following fields that you can configure by using values overrides:
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

The **eventData** field contains an escaped JSON string with the following fields that you can configure by using values overrides:
- `$id` - Specifies a unique tenant ID.
- `$name` - Specifies the tenant name.

## Configuration

The Tenant Fetcher binary allows you to override some configuration parameters. To get a list of the configurable parameters, see [main.go](https://github.com/kyma-incubator/compass/blob/75aff5226d4a105f4f04608416c8fa9a722d3534/components/director/cmd/tenantfetcher-job/main.go#L34)

## Local Development

### Prerequisites

The Tenant Fetcher requires access to:
1. A Configured PostgreSQL database with the imported Director's database schema.
1. An API that can be called to fetch tenant events. For details about implementing the Tenant Events API that the Tenant Fetcher can consume, see the [Endpoints](#endpoints) section of this document. 

### Run
As the Tenant Fetcher job is a short-lived process, it is useful to start and debug it directly from your IDE. Make sure that you provide all required configuration properties as environment variables.
