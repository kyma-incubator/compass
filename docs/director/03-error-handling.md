# Error handling

Director API returns custom errors if anything goes wrong while handling the request. The errors can be divided into
two categories: 
- Regular HTTP errors 
- GraphQL errors

## Regular errors

Regular errors are related to the request itself. They may concern issues such as missing headers. These are the regular errors that may occur when calling the Director API:

**Tenant not found**
This error occurs either when there is no such tenant in the database, or when the passed tenant does not match the tenant in the [System Auth](https://github.com/kyma-incubator/compass/blob/732486482e1d71384be4d705b1ed260365f74c1c/docs/compass/03-01-security.md#system_auths-table). 


## GraphQL errors

GraphQL errors are related to the request execution. They may concern issues with connecting to the database, or inter-server issues. These are the GraphQL errors that may occur when calling the Director API:

**Tenant is required**
This error occurs when a tenant, that is required by the operation, wasn't passed.
