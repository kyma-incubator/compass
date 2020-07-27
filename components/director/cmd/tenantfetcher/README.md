# Tenant Fetcher

## Overview

This application fetches events containing information about created, updated, and deleted tenants. It also stores events in the database or removes them from it.

## Prerequisites

Tenant Fetcher requires access to:
1. Configured PostgreSQL database with the imported Director's database schema.
2. API that can be called to fetch tenant events. For details about implementing the Tenant Events API that the Tenant Fetcher can consume, see [this](https://github.com/kyma-incubator/compass/blob/master/components/director/internal/tenantfetcher/README.md) document. 

## Configuration

The Tenant Fetcher binary allows you to override some configuration parameters. You can specify the following environment variables:

| Environment variable                | Default       | Description                                                                                                                                                                                                     |
|-------------------------------------|---------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **APP_DB_USER**                     | `postgres`    | Database username                                                                                                                                                                                               |
| **APP_DB_PASSWORD**                 | `pgsql@12345` | Database password                                                                                                                                                                                               |
| **APP_DB_HOST**                     | `localhost`   | Database host                                                                                                                                                                                                   |
| **APP_DB_PORT**                     | `5432`        | Database port                                                                                                                                                                                                   |
| **APP_DB_NAME**                     | `postgres`    | Database name                                                                                                                                                                                                   |
| **APP_DB_SSL**                      | `disable`     | Database SSL mode (`disable` or `enable`)                                                                                                                                                                       |
| **APP_CLIENT_ID**                   |               | OAuth 2.0 client ID                                                                                                                                                                                             |
| **APP_CLIENT_SECRET**               |               | OAuth 2.0 client secret                                                                                                                                                                                         |
| **APP_OAUTH_TOKEN_ENDPOINT**        |               | Endpoint for fetching the OAuth 2.0 access token                                                                                                                                                                |
| **APP_ENDPOINT_TENANT_CREATED**     |               | Tenant Events API endpoint for fetching created tenants                                                                                                                                                         |
| **APP_ENDPOINT_TENANT_DELETED**     |               | Tenant Events API endpoint for fetching deleted tenants                                                                                                                                                         |
| **APP_ENDPOINT_TENANT_UPDATED**     |               | Tenant Events API endpoint for fetching updated tenants                                                                                                                                                         |
| **APP_MAPPING_FIELD_NAME**          | `name`        | Name of the field in the event data payload containing the tenant name                                                                                                                                          |
| **APP_MAPPING_FIELD_ID**            | `id`          | Name of the field in the event data payload containing the tenant ID                                                                                                                                            |
| **APP_MAPPING_FIELD_DETAILS**       | `details`     | Name of the field in the event data payload containing the details of the event                                                                                                                                 |
| **APP_TENANT_TOTAL_PAGES_FIELD**    |               | Name of the field in the service response pointing the total pages count                                                                                                                                        |
| **APP_TENANT_TOTAL_RESULTS_FIELD**  |               | Name of the field in the service response pointing the total count of events                                                                                                                                    |
| **APP_TENANT_EVENTS_FIELD**         |               | Name of the field in the service response pointing to the array of events                                                                                                                                       |
| **APP_QUERY_PAGE_SIZE_FIELD**       |               | Name of the query parameter for the page size of the response                                                                                                                                                   |
| **APP_QUERY_PAGE_NUM_FIELD**        |               | Name of the query parameter for the current page number                                                                                                                                                         |
| **APP_QUERY_TIMESTAMP_FIELD**       |               | Name of the query parameter for the timestamp                                                                                                                                                                   |
| **APP_QUERY_PAGE_START**            |               | Value for the first page of the response                                                                                                                                                                        |
| **APP_QUERY_PAGE_SIZE**             |               | Value for the page size of the response                                                                                                                                                                         |
| **APP_MAPPING_FIELD_DISCRIMINATOR** |               | Optional name of the field in the event data payload used to filter created tenants. If provided, only events containing this field with a value specified in **APP_MAPPING_VALUE_DISCRIMINATOR** will be used. |
| **APP_MAPPING_VALUE_DISCRIMINATOR** |               | Optional value of the discriminator field for filtering created tenants. It is used only if **APP_MAPPING_FIELD_DISCRIMINATOR** is provided.                                                                    |
| **APP_TENANT_PROVIDER**             |               | Tenant provider name                                                                                                                                                                                            |
| **APP_METRICS_PUSH_ENDPOINT**       |               | Optional Prometheus Pushgateway endpoint for pushing Tenant Fetcher metrics                                                                                                                                     |
