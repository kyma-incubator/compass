# Error Handling

Director API returns custom errors if anything goes wrong while handling the request. The errors can be divided into
two categories: regular HTTP errors and GraphQL errors.

## Regular Errors

These errors are returned when something is wrong with the request eg. missing headers.

- Tenant not found
    - Tenant does not occur in the database
    - In case of Application/Runtime flow the passed tenant does not match the tenant in System Auth


## GraphQL Errors

These errors are returned eg. when there was a problem with connecting to the database, or in case an inter server error occured.GraphQL.

- Tenant is required
    - No tenant was passed but it's required by the operation