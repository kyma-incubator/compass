# Asynchronous Unregister and Unpair of Applications

**_Unregister_ vs _Unpair_**

The unregister operation results in deletion of the application from the Director database, along with all of its related resources, including external ones. The unpair operation keeps the application, its APIs and events, but clears the issued credentials, and externally created resources.

From _Operations Controller_ point of view, the operations are the same, since the same steps need to be done by it - the externally created resources have to be deleted.

When _unregister_ is mentioned below, one can also think of _unpair_.

## Flow

1. Director receives unregister/unpair request
1. A GraphQL directive for asynchronous operations is triggered before the request is processed further
1. Scheduler (built-in Director) checks if there is concurrent operation running. Returns 
1. Error response is returned if there is concurrent operation in progress
1. If no operation is is progress, the Scheduler creates a new Operation CRD
1. The GQL flow is continued, and an appropriate response with poll URL is returned to the client
1. The Operations Controller processes the new operation:
    1. Retrieves the application from Director along with its webhooks (application and application template owned)
    1. Initiates the delete operation on the external system side
    1. Gets the _Poll URL_ from the response and checks the status of the operation there
    1. When the operation is finished in the external system (successfully or not) the Operations Controller notifies Director that the operation is finalized with the given status
1. When Director receives the finalization notification it
    * deletes the application and all related resources in case of unregister
    * marks the application as ready - that way follow up asynchronous operations will be possible, in case of unpair

![](./assets/async-application-operation.png)

More detailed diagram of the reconciliation loop:

![](./assets/async-application-operation-loop.png)
