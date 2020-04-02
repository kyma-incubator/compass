# Tenant Fetcher

## Overview

Tenant Fetcher fetches information about tenants from external APIs.

It also contains details about the API schema that should be implemented to allow integration with the Tenant Fetcher.

## Configuration

You can configure Tenant Fetcher CronJob with the following Helm values overrides:

| Parameter | Description |  Default value |
|-----------|-------------|---------------|
| **global.tenantFetcher.enabled** | Enables the Tenant Fetcher CronJob. | `false` |
| **global.tenantFetcher.providerName** | Specifies the name of the tenants provider. | `"compass"` |
| **global.tenantFetcher.schedule** | CronJob schedule | `"*/5 * * * *"` |
| **global.tenantFetcher.oauth.client** | OAuth 2.0 client ID | None |
| **global.tenantFetcher.oauth.secret** | OAuth 2.0 client secret | None |
| **global.tenantFetcher.oauth.tokenURL** | Endpoint for fetching the OAuth 2.0 access token to the Tenant Events API | None |
| **global.tenantFetcher.endpoints.tenantCreated** | Tenant Events API endpoint for fetching created tenants | `"127.0.0.1/events?type=created"` |
| **global.tenantFetcher.endpoints.tenantDeleted** | Tenant Events API endpoint for fetching deleted tenants | `"127.0.0.1/events?type=deleted"` |
| **global.tenantFetcher.endpoints.tenantUpdated** | Tenant Events API endpoint for fetching updated tenants | `"127.0.0.1/events?type=updated"` |
| **global.tenantFetcher.fieldMapping.idField** | Name of the field in the event data payload containing the tenant name | `"id"` |
| **global.tenantFetcher.fieldMapping.nameField** | Name of the field in the event data payload containing the tenant ID | `"name"` |
| **global.tenantFetcher.fieldMapping.discriminatorField** | Optional name of the field in the event data payload used to filter created tenants. If provided, only events containing this field with the value specified in **discriminatorValue** will be used. | None |
| **global.tenantFetcher.fieldMapping.discriminatorValue** | Optional value of the discriminator field used to filter created tenants. It is used only if **discriminatorField** is provided. | None |

## Tenant Events API

### Authorization
Tenant Events API should use the OAuth 2.0 client credentials authorization flow.

### Endpoints

Three endpoints are supported, each for different type of events.
All endpoints should accept the same parameters specified in the **Query parameters** section.
Every endpoint should return a specific payload.

- **Query parameters**

Every endpoint should support the following query parameters:
- `ts` - specifies a timestamp in Unix time, that is the date since which events should we fetched
- `page` - specifies the page number that should be fetched, starting from 1
- `resultsPerPage` - specifies how many results should be included in a single page

- **Tenant creation endpoint**

JSON payload:
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

The `eventData` field contains an escaped JSON string, with the following fields that can be configured in values overrides:
- `$id` - specifies a unique tenant ID
- `$name` - specifies the tenant name
- `$discriminator` - specifies an optional field that can be used to distinguish different types of tenants

- **Tenant deletion endpoint**

JSON payload:
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

The `eventData` field contains an escaped JSON string, with the following fields that can be configured in values overrides:
- `$id` - specifies a unique tenant ID
- `$name` - specifies the tenant name

- **Tenant update endpoint**

JSON payload:
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

The `eventData` field contains an escaped JSON string, with the following fields that can be configured in values overrides:
- `$id` - specifies a unique tenant ID
- `$name` - specifies the tenant name
