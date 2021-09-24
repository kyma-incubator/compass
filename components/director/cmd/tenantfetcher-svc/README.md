# Tenant Fetcher Service

## Overview

This application provides an API for managing tenant subscriptions.

A subscription binds a provider runtime to a tenant. A given runtime is called a provider runtime when it has a specific `APP_SUBSCRIPTION_PROVIDER_LABEL_KEY` label.
Creating a subscription means that the provider runtime with a matching provider label can access tenant resources on behalf of the subscribed tenant. 
If the tenants from the subscription requests do not exist in the Compass DB, they are created automatically at the moment.

### Exposed API endpoints

|                           API                           |                                                      Description                                                    |
|---------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------|
| `PUT <APP_ROOT_API>/<APP_HANDLER_ENDPOINT>`             | You can use this endpoint to create a tenant, if the tenant does not exist. It gracefully handles the already                                                                     existing tenants. If the tenant-fetcher job has fetched the creation event this means that the tenant might exist.                                                               If its parent tenants do not exist, the API endpoint also creates all of them. Finally, it labels the tenant with                                                                 its subdomain.                                                                                                      |
| `PUT <APP_ROOT_API>/<APP_REGIONAL_HANDLER_ENDPOINT>`    | You can use this endpoint to subscribe tenants to runtimes with the given labels. If the tenant does not exist it is                                                             created together with its relative tenants. Regional tenants are labeled with their subdomains and regions.                                                                       Subscribed runtimes are labeled with subscriber tenant IDs.                                                         |
| `DELETE <APP_ROOT_API>/<APP_HANDLER_ENDPOINT>`          | You can use this endpoint to unsubscribe tenants from runtimes with the given labels. Then, the unsubscribing tenant                                                             ID is removed from subscribed runtimes label.                                                                       |
| `DELETE <APP_ROOT_API>/<APP_REGIONAL_HANDLER_ENDPOINT>` | You can use this endpoint to unsubscribe tenants from runtimes with the given labels. Then, the unsubscribing tenant                                                             ID is removed from subscribed runtimes label.                                                                       |
| `GET <APP_ROOT_API>/<APP_DEPENDENCIES_ENDPOINT>`        | You can use this endpoint to return all external applications, which must be informed for the tenant creation before                                                             Compass. That is, if Compass communicates with a multi-tenant application, and they share the same tenants, then                                                                 if a new tenant is created in Compass, the multi-tenant application must also create that tenant if it does not                                                                   exist. Currently, an empty json is returned.                                                                        |

All endpoints expect the same body:

```
{
    "<APP_TENANT_PROVIDER_TENANT_ID_PROPERTY>": "accountTenantID",
    "<APP_TENANT_PROVIDER_CUSTOMER_ID_PROPERTY>": "customerTenantID",
    "<APP_TENANT_PROVIDER_SUBACCOUNT_TENANT_ID_PROPERTY>": "subaccountTenantID",
    "<APP_TENANT_PROVIDER_SUBDOMAIN_PROPERTY>": "my-subdomain",
    "<APP_TENANT_PROVIDER_SUBSCRIPTION_PROVIDER_ID_PROPERTY>": "subscriptionProviderID",

}
```
Note that `<APP_TENANT_PROVIDER_CUSTOMER_ID_PROPERTY>` is optional.

## Installation Prerequisites

Tenant fetcher requires access to a configured PostgreSQL database with imported Director's database schema.

## Configuration

The Tenant Fetcher binary allows you to override some configuration parameters. You can specify the following environment variables:

| Environment variable                                       | Default                                     | Description                                                        |
|------------------------------------------------------------|---------------------------------------------|--------------------------------------------------------------------|
| **APP_DB_USER**                                            | `postgres`                                  | Database username                                                  |
| **APP_DB_PASSWORD**                                        | `pgsql@12345`                               | Database password                                                  |
| **APP_DB_HOST**                                            | `localhost`                                 | Database host name                                                 |
| **APP_DB_PORT**                                            | `5432`                                      | Database port number                                               |
| **APP_DB_NAME**                                            | `postgres`                                  | Database name                                                      |
| **APP_DB_SSL**                                             | `disable`                                   | Database SSL mode (`disable` or `enable`)                          |
| **APP_DB_MAX_OPEN_CONNECTIONS**                            | `2`                                         | The maximum number of open connections to the database             |
| **APP_DB_MAX_IDLE_CONNECTIONS**                            | `1`                                         | The maximum number of connections in the idle connection pool      |
| **APP_LOG_FORMAT**                                         | `kibana`                                    | The format of the logs (`kibana` or `text`)                        |
| **APP_ADDRESS**                                            | `127.0.0.1:8080`                            | The address and port for the service to listen on                  |
| **APP_ROOT_API**                                           | `/tenants`                                  | The root API where the server will listen to. All following APIs                                                                                                                  must be accessed through the root API                              |
| **APP_TENANT_PROVIDER**                                    | `external-provider`                         | Tenant provider name                                               |
| **APP_HANDLER_ENDPOINT**                                   | `/v1/callback/{tenantID}`                   | The endpoint used for tenant management                            |
| **APP_REGIONAL_HANDLER_ENDPOINT**                          | `/v1/regional/{region}/callback/{tenantID}` | The endpoint used for management of regional tenants               |
| **APP_DEPENDENCIES_ENDPOINT**                              | `/v1/dependencies`                          | The endpoint used for declaring external dependencies              |
| **APP_TENANT_PATH_PARAM**                                  | `tenantId`                                  | The path parameter name, which will be used for tenant ID          |
| **APP_REGION_PATH_PARAM**                                  | `region`                                    | The path parameter name, which will be used for region             |
| **APP_TENANT_PROVIDER_TENANT_ID_PROPERTY**                 | `tenantId`                                  | Name of the json field containing the tenant ID                    |
| **APP_TENANT_PROVIDER_CUSTOMER_ID_PROPERTY**               | `customerId`                                | Name of the json field containing the customer ID                  |
| **APP_TENANT_PROVIDER_SUBACCOUNT_TENANT_ID_PROPERTY**      | `subaccountTenantId`                        | Name of the json field containing the subaccount tenant ID         |
| **APP_TENANT_PROVIDER_SUBDOMAIN_PROPERTY**                 | `subdomain`                                 | Name of the json field containing the tenant subdomain             |
| **APP_TENANT_PROVIDER_SUBSCRIPTION_PROVIDER_ID_PROPERTY**  | `subscriptionProviderId`                    | Name of the json field containing the subscription provider ID,                                                                                                                  which must be mapped to a runtime                                  |
| **APP_SUBSCRIPTION_PROVIDER_LABEL_KEY**                    | `subscriptionProviderIdLabel`               | The runtime label key used for mapping subscription provider IDs and                                                                                                              runtimes                                                           |
| **APP_CONSUMER_SUBACCOUNT_IDS_LABEL_KEY**                  | `consumer_subaccount_ids`                   | The runtime label key used for mapping the subaccount IDs, to which                                                                                                              the runtime has access                                             |
| **APP_JWKS_ENDPOINT**                                      | `file://hack/default-jwks.json`             | The path for JWKS                                                  |
| **APP_SUBSCRIPTION_CALLBACK_SCOPE**                        | `Callback`                                  | The JWT scope required for accessing the APIs                      |
| **APP_ALLOW_JWT_SIGNING_NONE**                             | `false`                                     | Enable trust to tokens signed with the `none` algorithm. Must be                                                                                                                  used for test  purposes only.                                      |
