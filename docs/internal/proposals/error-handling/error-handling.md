# Error handling in Director graphql

In compass director, we have to somehow filter errors which should be visible to user.
Basic list of errors which should be displayed to the user:
* Internal error
* Invalid Data
* Not Unique
* Not Found
* Tenant Not Found
* Insufficient Scopes

This list can be split into 2 groups due to different handling:
* internal errors
* the rest of errors, which will be presented to the user

In group of internal errors, are errors from:
* external libraries
* panics
* postgreSQL, except errors like NotFound, NotUnique

To deal with the errors we introduced custom errors, which contains error codes.

## Proposed custom errors and theirs error codes (the error codes can change):

| Error type           | Error code  |                            Description                                                      |
|----------------------|-------------|---------------------------------------------------------------------------------------------|
| InternalError       | 1           | Error which cannot be handled in director                                                   |
| NotFound            | 2           | Error indicate that given resource cannot be found and further processing is impossible     |
| NotUnique           | 3           | Error indicates that given resource is not unique                                           |
| TenantNotFound     | 4           | Internal Tenant not found in director                                                       |
| InvalidData          | 5           | The input data is invalid, error description should be delivered in error message           | 
| ConstraintViolation  | 6           | Error which indicate that this operation can't happen because referenced resource not exist |
| Insufficient Scopes  | 7           | Error which indicate insufficient scopes                                                    |

NotFoundError can be triggered in mutation like `addPackage` to not existing application.

ConstraintViolation can be replaced with `NotFound`  error.

## Proposed processing flow:

![](error-handling.svg)

Description of following steps
* PostgreSQL Error Mapper - this component map postgreSQL errors to custom errors
* Error Middleware - this component search for custom error in error stack. If such error is found, the custom `GraphlError` is created with erro message and error code.
In case of internal errors, the whole error is logged and `internal server error` is sent to client.
* Error presenter - is responsible for error presentation to user. Checks if given error is type of custom `GraphlError`.
If yes, the presenter add additional property `error-code` to the metadata to the error grom graphql library.
