# Error handling in Director graphql

In Compass Director, we have to filter errors which should be visible to user and add information which can be readable by machine.

## Error types
Basic list of errors which should be displayed to the user:
* InternalError
* NotFound
* NotUnique
* InvalidData
* InsufficientScopes
* ConstraintViolation
* TenantIsRequired
* TenantNotFound

This list can be split into 2 groups due to different handling:
* internal errors
* the rest of errors, which will be presented to the user

In group of internal errors, are errors from:
* external libraries
* panics
* postgreSQL, except errors like NotFound, NotUnique

To deal with the errors we introduced custom errors, which contains error codes.

## Custom errors and theirs error codes:

| Error type           | Error code  |                            Description                                                            |
|----------------------|-------------|---------------------------------------------------------------------------------------------------|
| InternalError        | 10          | Error which cannot be handled in director                                                         |
| NotFound             | 20          | Error indicate that given resource cannot be found and further processing is impossible           |
| NotUnique            | 21          | Error indicates that given resource is not unique                                                 |
| InvalidData          | 22          | The input data is invalid, error description should be delivered in error message                 | 
| InsufficientScopes   | 23          | Error which indicate that the client doesn't have sufficient permissions to execute the operation |
| ConstraintViolation  | 24          | Error which indicate that this operation can't happen because referenced resource not exist       |
| TenantIsRequired     | 25          | Tenant not found in request and is required to successful execute the request                     |
| TenantNotFound       | 26          | Internal Tenant not found in director                                                             |

NotFoundError can be triggered in mutation like `addPackage` to not existing application.

## Error processing flow:

![](error-handling.svg)

Description of following steps
* PostgreSQL Error Mapper - this component map postgreSQL errors to custom errors
* Error presenter - this component search for custom error in error stack. 
If such error is found, the presenter add `error_code` and `error` metadata to the GraphQL error in section `extensions`
In case of internal errors, the whole error is logged and `internal server error` is sent to client.
In case of errors which are not handled by custom error library, the error is returned without error code and error message is logged.

## Proof of concept
[Here](https://github.com/kyma-incubator/compass/pull/1366) is implemented PoC.
