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
1. Get an audit log from the mock service based on the application's name.
1. Compare the audit log with the request for registering the application.

## Asynchronous application operations
### Unpair
The Async Unpair test uses the endpoint that returns "In Progress" until dynamically configured by the test itself to start returning "Completed" - that endpoint simulates the application deleting its internal resources related to the Compass connection. Steps run by the test:

1. Register an Application with webhook of type `UNREGISTER_APPLICATION`.
1. Verify that the application exists.
1. Trigger asynchronous deletion of the application (call `unpairApplication` mutation with `mode: async`).
1. Check that `Operation` resource exists for that application and it's status is `In Progress`.
1. Configure the `External Services Mock` server to return successful deletion of the external resources.
1. Verify that the application still exists.

### Delete
The Async Delete test uses the endpoint that returns "In Progress" until dynamically configured by the test itself to start returning "Completed" - that endpoint simulates the application deleting its internal resources related to the Compass connection. Steps run by the test:

1. Register an Application with webhook of type `UNREGISTER_APPLICATION`.
1. Verify that the application exists.
1. Trigger asynchronous deletion of the application (call `unregisterApplication` mutation with `mode: async`).
1. Check that `Operation` resource exists for that application and it's status is `In Progress`.
1. Configure the `External Services Mock` server to return successful deletion of the external resources.
1. Verify that the application no longer exists.
