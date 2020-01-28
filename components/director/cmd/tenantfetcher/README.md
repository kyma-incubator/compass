# Tenant fetcher

This application fetches events containing information about: created, updated, deleted tenants and stores or removes those tenants in the database.

### Prerequisites

Tenant Fetcher requires access to:
1. Configured PostgreSQL database with imported Director's database schema.
2. API that can be called to fetch tenant events, details about implementing Tenant Events API that can be consumed by Tenant Fetcher can be found [here](https://github.com/kyma-incubator/compass/blob/master/docs/compass/03-tenant-fetching.md). 

## Configuration

The Tenant Fetcher binary allows to override some configuration parameters. You can specify following environment variables.

| Environment variable            | Default     | Description                                                                                                                                                                             |
|---------------------------------|-------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| APP_DB_USER                     | postgres    | Database username                                                                                                                                                                       |
| APP_DB_PASSWORD                 | pgsql@12345 | Database password                                                                                                                                                                       |
| APP_DB_HOST                     | localhost   | Database host                                                                                                                                                                           |
| APP_DB_PORT                     | 5432        | Database port                                                                                                                                                                           |
| APP_DB_NAME                     | postgres    | Database name                                                                                                                                                                           |
| APP_DB_SSL                      | disable     | Database SSL mode (disable / enable)                                                                                                                                                    |
| APP_CLIENT_ID                   |             | OAuth 2.0 client id                                                                                                                                                                     |
| APP_CLIENT_SECRET               |             | OAuth 2.0 client secret                                                                                                                                                                 |
| APP_OAUTH_TOKEN_ENDPOINT        |             | Endpoint for fetching OAuth 2.0 access token                                                                                                                                            |
| APP_ENDPOINT_TENANT_CREATED     |             | Tenant Events API endpoint for fetching created tenants                                                                                                                                 |
| APP_ENDPOINT_TENANT_DELETED     |             | Tenant Events API endpoint for fetching deleted tenants                                                                                                                                 |
| APP_ENDPOINT_TENANT_UPDATED     |             | Tenant Events API endpoint for fetching updated tenants                                                                                                                                 |
| APP_MAPPING_FIELD_NAME          | name        | Name of field in event data payload containing tenant name                                                                                                                                      |
| APP_MAPPING_FIELD_ID            | id          | Name of field in event data payload containing tenant id                                                                                                                                        |
| APP_MAPPING_FIELD_DISCRIMINATOR |             | Optional name of field in event data payload used to filter created tenants, if provided only events containing this field with value specified in APP_MAPPING_VALUE_DISCRIMINATOR will be used |
| APP_MAPPING_VALUE_DISCRIMINATOR |             | Optional value of discriminator field used to filter created tenants, used only if APP_MAPPING_FIELD_DISCRIMINATOR is provided                                                                                                                    |
| APP_TENANT_PROVIDER             |             | Tenant provider name                                                                                                                                                                    |