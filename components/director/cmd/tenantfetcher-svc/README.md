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

## Development

### Prerequisites

Tenant Fetcher requires access to:
1. Configured PostgreSQL database with the imported Director's database schema.

### Run

You can start the Tenant Fetcher deployment via the designated `run.sh` script in Director:
```bash
./run.sh --tenant-fetcher
```

That will bring up a PostgreSQL database with Director's schema.

The up-to-date properties supported by the `run.sh` script can be found in the [Director component local documentation](../../README.md#local-development).
