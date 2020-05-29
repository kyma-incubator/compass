# Error handling

Director API returns custom errors when there are problems with handling the request. 

## GraphQL errors

These are the GraphQL errors that may occur when calling the Director API:

| Error type             | Error code    |                            Description                                                                      |
|------------------------|---------------|-------------------------------------------------------------------------------------------------------------|
| `InternalError`        | `10`          | Specifies any error that cannot be handled in the Director. Such situation is unique and serious.           |
| `NotFound`             | `20`          | Indicates that a given resource cannot be found and further processing is impossible.                       |
| `NotUnique`            | `21`          | Indicates that a given resource is not unique.                                                              |
| `InvalidData`          | `22`          | Indicates that the input data is invalid with the reason described in the error message.                    | 
| `InsufficientScopes`   | `23`          | Indicates that the client doesn't have permission to execute the operation.                                 |
| `TenantRequired`       | `24`          | Indicates that a tenant is not found in the request, which is required to successfully handle the request.  |
| `TenantNotFound`       | `25`          | Indicates that the internal tenant is not found in the Director.                                            |
| `Unauthorized`         | `26`          | Indicates that request cannot be authorized.                                                                |            
| `InvalidOperation`     | `27`          | Indicates that the operation is invalid, because of restrictions, eg.: cannot delete label definition, where the label is used.   |

The Graphql Error response have one additional field `extenstion` which contains `error_code` and `error` fields.
See example of response with error:
```json
{
  "errors": [
    {
      "message": "Error message  [key=value; key=value]",
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
