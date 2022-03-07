# Tenant Fetcher

## Overview

This application fetches events containing information about created, moved, and deleted tenants. Then based on the event type, does the corresponding create, update or move operation on the tenant from the event in the Director database.

## Details

This section describes the API schema that a server must implement to integrate with the Tenant Fetcher.

### Authorization

Tenant Events API should use the OAuth 2.0 client credentials authorization flow or mTLS and trust the Compass client certificate.

### Endpoints

There are three types of supported endpoints that receive different types of events:
- Tenant creation endpoint
- Tenant deletion endpoint
- Tenant update endpoint

Every endpoint must return a specific payload and accept the following type of query parameters:
- **global.tenantFetchers.*job_name*.queryMapping.timestampField** - specifies a timestamp in Unix time format, that is the date from which events are fetched
- **global.tenantFetchers.*job_name*.queryMapping.pageNumField** - specifies the number of the page to be fetched, starting from a preconfigured number via **global.tenantFetchers.*job_name*.query.startPage**
- **global.tenantFetchers.*job_name*.queryMapping.pageSizeField** - specifies the number of results included on a single page

#### Response

Almost every top-level data that is expected by the tenant fetcher is configurable. Consider the following example response for the creation endpoint:

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

- The top-level `events` array is configured with: **global.tenantFetchers.*job_name*.fieldMapping.tenantEventsField**
- The top-level `totalResults` is configured with: **global.tenantFetchers.*job_name*.fieldMapping.totalResultsField**
- The top-level `totalPages` is configured with: **global.tenantFetchers.*job_name*.fieldMapping.totalPagesField**
- The inner field `eventData` contains the details of an event and it is configured by: **global.tenantFetchers.*job_name*.fieldMapping.detailsField**. The details field is expected to be either a JSON object or a string containing JSON. Example:

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

On success, the endpoint returns the following JSON payload:
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

The **eventData** field contains an escaped JSON string with the following fields that you can configure using values overrides:
- `$id` - specifies a unique tenant ID
- `$parent_id` - specifies a unique tenant ID belonging to the parent tenant - should be of type `account` if the created tenant is a `subaccount`, and of type `customer` if the created tenant is of type `account`
- `$name` - specifies the tenant name
- `$discriminator` - specifies an optional field that can be used to distinguish different types of tenants

##### Tenant deletion endpoint

On success, the endpoint returns the following JSON payload:
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

The **eventData** field contains an escaped JSON string with the following fields that you can configure using values overrides:
- `$id` - specifies a unique tenant ID
- `$name` - specifies the tenant name

##### Tenant update endpoint

On success, the endpoint returns the following JSON payload:
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

The **eventData** field contains an escaped JSON string with the following fields that you can configure using values overrides:
- `$id` - specifies a unique tenant ID
- `$name` - specifies the tenant name

## Configuration

Tenant Fetcher binary allows you to override some configuration parameters. Up-to-date list of the configurable parameters of Tenant Fetcher can be found [here](https://github.com/kyma-incubator/compass/blob/75aff5226d4a105f4f04608416c8fa9a722d3534/components/director/cmd/tenantfetcher-job/main.go#L34)

## Local Development

### Prerequisites

Tenant Fetcher requires access to:
1. Configured PostgreSQL database with the imported Director's database schema.
1. API that can be called to fetch tenant events. For details about implementing the Tenant Events API that the Tenant Fetcher can consume, see [the Endpoints section](#endpoints) of this document. 

### Run
As the job is a short-lived process, one can simply use an IDE for running and debugging the Tenant Fetcher job. Remember to provide all required configuration as environment variables.
