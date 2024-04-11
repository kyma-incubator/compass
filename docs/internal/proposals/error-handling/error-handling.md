# Error Handling in Director GraphQL

Errors can originate from different sources. This document describes our approach to handling them internally and presenting them in a readable form to the users.

## Errors Displayed to the User

Errors that are returned to the user can be split into two groups, based on the error handling approach. Errors that originate from external systems, libraries, panics, and directly not dependent on the user input are handled as the internal ones. Such errors are returned to the user as the `InternalError` type. Errors that directly depend on the user input are classified as separate types.

## Custom Errors and Error Codes

To differentiate the errors, we introduced our own class of custom errors that include error codes.
This is the list of custom errors that are displayed to the user:

| Error type             | Error code    |                            Description                                                                      |
|------------------------|---------------|-------------------------------------------------------------------------------------------------------------|
| `InternalError`        | `10`          | Specifies any error that cannot be handled in the Director. Such situation is unique and serious.           |
| `NotFound`             | `20`          | Indicates that a given resource cannot be found and further processing is impossible.                       |
| `NotUnique`            | `21`          | Indicates that a given resource is not unique.                                                              |
| `InvalidData`          | `22`          | Indicates that the input data is invalid with the reason described in the error message.                    | 
| `InsufficientScopes`   | `23`          | Indicates that the client doesn't have permission to execute the operation.                                 |
| `ConstraintViolation`  | `24`          | Indicates that the operation cannot be executed because the referenced resource does not exist.                  |
| `TenantIsRequired`     | `25`          | Indicates that a tenant is not found in the request, which is required to successfully handle the request. |
| `TenantNotFound`       | `26`          | Indicates that the internal tenant is not found in the Director.                                            |

## Error Processing Flow

![](error-handling.svg)
1. Internal errors, external lib errors, PostgreSQL errors, and directives are sources of errors in the Director.
2. PostgreSQL Error Mapper maps PostgreSQL errors to custom errors.
3. Panic Handler recovers all panics and transforms them into custom errors.
3. Error Presenter searches for custom errors in the error stack. 
If such an error is found, the Error Presenter adds `error_code` and `error` metadata to the GraphQL error in the **extensions** section.
In case of internal errors, the error is logged and the `internal server error` is sent to the client.
In case of errors that are not handled by custom error library, the error is returned without any error code and an error message is logged.

## PoC

See the [implemented PoC](https://github.com/kyma-incubator/compass/pull/1366) for more details.
