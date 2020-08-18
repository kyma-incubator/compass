# Tenant Fetcher

## Overview

Tenant Fetcher fetches information about tenants from external APIs.

## Details

This section describes the API schema that a server must implement to integrate with the Tenant Fetcher.

### Authorization

Tenant Events API should use the OAuth 2.0 client credentials authorization flow.

### Endpoints

There are three types of supported endpoints that receive different types of events:
- Tenant creation endpoint
- Tenant deletion endpoint
- Tenant update endpoint

Every endpoint must return a specific payload and accept the following type of query parameters:
- **global.tenantFetchers.*job_name*.queryMapping.timestampField** - specifies a timestamp in Unix time format, that is the date from which events are fetched
- **global.tenantFetchers.*job_name*.queryMapping.pageNumField** - specifies the number of the page to be fetched, starting from a preconfigured number via **global.tenantFetchers.*job_name*.query.startPage**
- **global.tenantFetchers.*job_name*.queryMapping.pageSizeField** - specifies the number of results included on a single page

### Response

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


#### Tenant creation endpoint

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
- `$name` - specifies the tenant name
- `$discriminator` - specifies an optional field that can be used to distinguish different types of tenants

#### Tenant deletion endpoint

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

#### Tenant update endpoint

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

Tenant Fetcher binary allows you to override some configuration parameters. You can specify the following environment variables:

| Parameter | Description |  Default value |
|-----------|-------------|---------------|
| **global.tenantFetchers.*job_name*.enabled** | Parameter that enables the Tenant Fetcher CronJob | `false` |
| **global.tenantFetchers.*job_name*.providerName** | Name of the tenants provider | `"compass"` |
| **global.tenantFetchers.*job_name*.schedule** | Parameter that specifies how often Tenant Fetcher fetches information about tenants | `"*/5 * * * *"` |
| **global.tenantFetchers.*job_name*.oauth.client** | OAuth 2.0 client ID | None |
| **global.tenantFetchers.*job_name*.oauth.secret** | OAuth 2.0 client Secret | None |
| **global.tenantFetchers.*job_name*.oauth.tokenURL** | Endpoint for fetching an OAuth 2.0 access token to the Tenant Events API | None |
| **global.tenantFetchers.*job_name*.endpoints.tenantCreated** | Tenant Events API endpoint for fetching created tenants | `"127.0.0.1/events?type=created"` |
| **global.tenantFetchers.*job_name*.endpoints.tenantDeleted** | Tenant Events API endpoint for fetching deleted tenants | `"127.0.0.1/events?type=deleted"` |
| **global.tenantFetchers.*job_name*.endpoints.tenantUpdated** | Tenant Events API endpoint for fetching updated tenants | `"127.0.0.1/events?type=updated"` |
| **global.tenantFetchers.*job_name*.fieldMapping.idField** | Name of the field in the event data payload that contains the tenant ID | `"id"` |
| **global.tenantFetchers.*job_name*.fieldMapping.nameField** | Name of the field in the event data payload that contains the tenant name | `"name"` |
| **global.tenantFetchers.*job_name*.fieldMapping.discriminatorField** | Optional name of the field in the event data payload used to filter created tenants. If provided, only the events that contain this field with the value specified in **discriminatorValue** are used. | None |
| **global.tenantFetchers.*job_name*.fieldMapping.discriminatorValue** | Optional value of the discriminator field used to filter  created tenants. It is used only if **discriminatorField** is provided. | None |
| **global.tenantFetchers.*job_name*.fieldMapping.tenantEventsField** | Mandatory value of the field name of the top-level events array |
| **global.tenantFetchers.*job_name*.fieldMapping.totalPagesField** | Mandatory value of the field name of the top-level property showing the number of pages |
| **global.tenantFetchers.*job_name*.fieldMapping.totalResultsField** | Mandatory value of the field name of the top-level property showing the number of total results |
| **global.tenantFetchers.*job_name*.fieldMapping.detailsField** | Mandatory value of the field name of the inner property showing the event details |
| **global.tenantFetchers.*job_name*.queryMapping.pageNumField** | Mandatory value of the query parameter name for the page number |
| **global.tenantFetchers.*job_name*.queryMapping.pageSizeField** | Mandatory value of the query parameter name for the page size |
| **global.tenantFetchers.*job_name*.queryMapping.timestampField** | Mandatory value of the query parameter name for the timestamp |
| **global.tenantFetchers.*job_name*.query.startPage** | Mandatory value of the query parameter value for the starting page from which to fetch events |
| **global.tenantFetchers.*job_name*.query.pageSize** | Mandatory value of the query parameter value for the page size |

