# Error Handling

Director API returns custom errors when there are problems with handling the request. 

## GraphQL Errors

These are the GraphQL errors that may occur when calling the Director API:

| Error type             | Error code    |                            Description                                                                                                            |
|------------------------|---------------|---------------------------------------------------------------------------------------------------------------------------------------------------|
| `InternalError`        | `10`          | Specifies any error that cannot be handled in the Director. Such situation is unique and serious.                                                 |
| `NotFound`             | `20`          | Indicates that the given resource cannot be found and further processing is impossible.                                                           |
| `NotUnique`            | `21`          | Indicates that the given resource is not unique.                                                                                                  |
| `InvalidData`          | `22`          | Indicates that the input data is invalid with the reason described in the error message.                                                          | 
| `InsufficientScopes`   | `23`          | Indicates that the client doesn't have permission to execute the operation.                                                                       |
| `TenantRequired`       | `24`          | Indicates that a tenant is not found in the request, which is required to successfully handle the request.                                        |
| `TenantNotFound`       | `25`          | Indicates that the internal tenant is not found in the Director.                                                                                  |
| `Unauthorized`         | `26`          | Indicates that the request cannot be authorized.                                                                                                  |            
| `InvalidOperation`     | `27`          | Indicates that the operation is invalid because of certain restrictions, e.g.: you cannot delete a given label definition when the label is used. |

The GraphQL Error response has one additional field `extensions` which contains the `error_code` and `error` fields.
See an example of the response with an error:
```json
{
  "errors": [
    {
      "message": "Error message [key=value; key=value]",
      "path": [
        "path"
      ],
      "extensions": {
        "error": "ErrorName",
        "error_code": 0
      }
    }
  ],
  "data": null
}
```
