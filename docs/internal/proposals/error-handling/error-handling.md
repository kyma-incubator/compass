# Error handling in Director GraphQL

Errors can originate from different sources. This document describes our approach of handling them internally and presenting them in the readable form for the users.

## Errors displayed to the user
Errors that are returned to the user can be split into two groups, based on the error handling approach. Errors that originate from external systems, libraries, panics, and directly not depend on the user input are handled as the internal one. Such errors are returned to the user as the `InternalError` type. Errors that directly depend on the user input are classified as separate types which in detail describes the nature of a problem.

## Custom errors and error codes

To differentiate the errors, we introduced our own class of custom errors that include error codes.
This is the list of custom errors that are displayed to the user:

| Error type             | Error code    |                            Description                                                                      |
|------------------------|---------------|-------------------------------------------------------------------------------------------------------------|
| `InternalError`        | `10`          | Specifies any error that cannot be handled in the Director. Such situation is unique and serious.           |
| `NotFound`             | `20`          | Indicates that a given resource cannot be found and further processing is impossible.                       |
| `NotUnique`            | `21`          | Indicates that a given resource is not unique.                                                              |
| `InvalidData`          | `22`          | Indicates that the input data is invalid with the reason described in the error message.                    | 
| `InsufficientScopes`   | `23`          | Indicates that the client doesn't have permission to execute the operation.                                 |
| `ConstraintViolation`  | `24`          | Indicates that this operation can't happen because the referenced resource does not exist.                  |
| `TenantIsRequired`     | `25`          | Indicates that a tenant is not found in the request, which is required to successfully execute the request. |
| `TenantNotFound`       | `26`          | Indicates that the internal tenant is not found in the Director.                                            |

## Error processing flow

![](error-handling.svg)
* Internal errors, External lib errors, PostgreSQL error, directives are sources of errors in the Director.
* PostgreSQL Error Mapper - this component map PostgreSQL errors to custom errors
* Panic Handler - this component recovers all panics and transform them as custom errors
* Error presenter - this component search for custom error in error stack. 
If such error is found, the presenter add `error_code` and `error` metadata to the GraphQL error in section `extensions`
In case of internal errors, the whole error is logged and `internal server error` is sent to client.
In case of errors which are not handled by custom error library, the error is returned without error code and error message is logged.

## PoC
See the [implemented PoC](https://github.com/kyma-incubator/compass/pull/1366) for more details.
