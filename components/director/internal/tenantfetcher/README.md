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

Every endpoints must return a specific payload and accept the following query parameters:
- **ts** - specifies a timestamp in Unix time, that is the date from which events are fetched
- **page** - specifies the number of the page to be fetched, starting from `1`
- **resultsPerPage** - specifies the number of results included on a single page

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
| **global.tenantFetcher.enabled** | Parameter that enables the Tenant Fetcher CronJob | `false` |
| **global.tenantFetcher.providerName** | Name of the tenants provider | `"compass"` |
| **global.tenantFetcher.schedule** | Parameter that specifies how often Tenant Fetcher fetches information about tenants | `"*/5 * * * *"` |
| **global.tenantFetcher.oauth.client** | OAuth 2.0 client ID | None |
| **global.tenantFetcher.oauth.secret** | OAuth 2.0 client Secret | None |
| **global.tenantFetcher.oauth.tokenURL** | Endpoint for fetching an OAuth 2.0 access token to the Tenant Events API | None |
| **global.tenantFetcher.endpoints.tenantCreated** | Tenant Events API endpoint for fetching created tenants | `"127.0.0.1/events?type=created"` |
| **global.tenantFetcher.endpoints.tenantDeleted** | Tenant Events API endpoint for fetching deleted tenants | `"127.0.0.1/events?type=deleted"` |
| **global.tenantFetcher.endpoints.tenantUpdated** | Tenant Events API endpoint for fetching updated tenants | `"127.0.0.1/events?type=updated"` |
| **global.tenantFetcher.fieldMapping.idField** | Name of the field in the event data payload that contains the tenant ID | `"id"` |
| **global.tenantFetcher.fieldMapping.nameField** | Name of the field in the event data payload that contains the tenant name | `"name"` |
| **global.tenantFetcher.fieldMapping.discriminatorField** | Optional name of the field in the event data payload used to filter created tenants. If provided, only the events that contain this field with the value specified in **discriminatorValue** are used. | None |
| **global.tenantFetcher.fieldMapping.discriminatorValue** | Optional value of the discriminator field used to filter  created tenants. It is used only if **discriminatorField** is provided. | None |
