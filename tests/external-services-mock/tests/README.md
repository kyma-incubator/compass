#  External Services Mock integration tests

External Services Mock tests check the contract between Compass and external services. 

The tests cover the following scenarios:
- Audit logging
- API specification fetching
- Asynchronous deletion of applications
- Asynchronous unpair of applications

## Audit log test scenario

The audit log test performs the following operations:
1. Register an application through the Compass Gateway.
2. Get an audit log from the mock service based on the application's name
3. Compare the audit log with the request for registering the application.

## API specification scenario

The API specification test uses the endpoint that returns a random API specification on every call. It performs the following operations:
1. Register an API Definition with a fetch request.
2. Fetch the API specification.
3. Refetch the API specification and check if it is different from the previous one.
4. Get the API Definition and check if the API specification is equal to the new one.
