# Tenant fetching

## Overview

This document describes configuration of Tenant Fetcher that fetches information about tenants from external API.
It also contains details about API schema that should be implemented to allow integration with Tenant Fetcher.

## Configuring Tenant Fetcher

Tenant Fetcher CRON Job can be configured with helm value overrides listed below:

| Parameter | Default value | Description |
|-----------|-------------|---------------|
| **global.tenantFetcher.enabled** | false | Enables Tenant Fetcher CRON Job |
| **global.tenantFetcher.providerName** | "compass" | Name of tenants provider |
| **global.tenantFetcher.schedule** | "*/5 * * * *" | CRON Job schedule |
| **global.tenantFetcher.oauth.client** | None | OAuth 2.0 client id, used to connect to Tenant Events API |
| **global.tenantFetcher.oauth.secret** | None | OAuth 2.0 client secret, used to connect to Tenant Events API |
| **global.tenantFetcher.oauth.tokenURL** | None | Endpoint for fetching OAuth 2.0 access token to Tenant Events API |
| **global.tenantFetcher.endpoints.tenantCreated** | "127.0.0.1/events?type=created" | Tenant Events API endpoint for fetching created tenants |
| **global.tenantFetcher.endpoints.tenantDeleted** | "127.0.0.1/events?type=deleted" | Tenant Events API endpoint for fetching created tenants |
| **global.tenantFetcher.endpoints.tenantUpdated** | "127.0.0.1/events?type=updated" | Tenant Events API endpoint for fetching created tenants |
| **global.tenantFetcher.fieldMapping.idField** | "id" | Name of field in event data payload containing tenant name |
| **global.tenantFetcher.fieldMapping.nameField** | "name" | Name of field in event data payload containing tenant id |
| **global.tenantFetcher.fieldMapping.discriminatorField** | None | Optional name of field in event data payload used to filter created tenants, if provided only events containing this field with value specified in discriminatorValue will be used |
| **global.tenantFetcher.fieldMapping.discriminatorValue** | None | Optional value of discriminator field used to filter created tenants, used only if discriminatorField is provided  |

## Tenant Events API

### Authorization
Tenant Events API should be using OAuth 2.0 Client Credentials authorization flow.

### Endpoints

Three endpoints are supported, for different types of events.
All endpoints should accept the same query parameters specified below.
Each endpoint should return specific payload. 

#### Query params

Every endpoint should support following query parameters:
- `ts` - specifies timestamp in Unix time, specifies date since which events should we fetched.
- `page` - specifies page number that should be fetched (starting from 1).
- `resultsPerPage` - specifies how many results should be included on a single page. 

#### Tenant creation endpoint

Creation event payload:
```json
{

}
```
