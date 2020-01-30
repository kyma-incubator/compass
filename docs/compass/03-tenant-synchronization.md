# Tenant synchronization

## Overview

This document describes the configuration of the Tenant Fetcher that fetches information about tenants from an external API.
It also contains details about the API schema that should be implemented to allow integration with the Tenant Fetcher.

## Configuring the Tenant Fetcher

The Tenant Fetcher CronJob can be configured with the following Helm values overrides:

| Parameter | Default value | Description |
|-----------|-------------|---------------|
| **global.tenantFetcher.enabled** | `false` | Tenant Fetcher CronJob |
| **global.tenantFetcher.providerName** | `"compass"` | Name of the tenants provider |
| **global.tenantFetcher.schedule** | `"*/5 * * * *"` | CronJob schedule |
| **global.tenantFetcher.oauth.client** | None | OAuth 2.0 client ID |
| **global.tenantFetcher.oauth.secret** | None | OAuth 2.0 client secret |
| **global.tenantFetcher.oauth.tokenURL** | None | Endpoint for fetching the OAuth 2.0 access token to the Tenant Events API |
| **global.tenantFetcher.endpoints.tenantCreated** | `"127.0.0.1/events?type=created"` | Tenant Events API endpoint for fetching created tenants |
| **global.tenantFetcher.endpoints.tenantDeleted** | `"127.0.0.1/events?type=deleted"` | Tenant Events API endpoint for fetching deleted tenants |
| **global.tenantFetcher.endpoints.tenantUpdated** | `"127.0.0.1/events?type=updated"` | Tenant Events API endpoint for fetching updated tenants |
| **global.tenantFetcher.fieldMapping.idField** | `"id"` | Name of the field in the event data payload containing the tenant name |
| **global.tenantFetcher.fieldMapping.nameField** | `"name"` | Name of the field in the event data payload containing the tenant ID |
| **global.tenantFetcher.fieldMapping.discriminatorField** | None | Optional name of the field in the event data payload used to filter created tenants. If provided, only events containing this field with the value specified in **discriminatorValue** will be used. |
| **global.tenantFetcher.fieldMapping.discriminatorValue** | None | Optional value of the discriminator field used to filter created tenants. It is used only if **discriminatorField** is provided. |

## Tenant Events API

### Authorization
Tenant Events API should use the OAuth 2.0 client credentials authorization flow.

### Endpoints

Three endpoints are supported, each for different type of events.
All endpoints should accept the same parameters specified in the **Query parameters** section.
Every endpoint should return a specific payload. 

- **Query parameters**

Every endpoint should support the following query parameters:
- `ts` - specifies a timestamp in Unix time, that is the date since which events should we fetched.
- `page` - specifies the page number that should be fetched, starting from 1.
- `resultsPerPage` - specifies how many results should be included in a single page. 

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
