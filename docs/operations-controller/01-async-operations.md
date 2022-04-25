# Asynchronous Operations in Compass

To increase the response time for requests in Compass, a new asynchronous mechanism for required external calls is introduced.

## Use
The operation below outlines the use case that has enabled asynchronous API in CMP. This asynchronous API contains additional logic that is run separately of the main transaction. The additional API logic for the asynchronous functionality is part of the API business logic.

Currently, asynchronous operation is supported when deleting an application (system), and separately notifying the integration system or the LoB application that it must clean up the artifacts that were created specifically for this application pairing.

## Asynchronous Flow

### Operations

To achieve a delayed run of a given part of the API business logic (for example, running remote calls separately from the main transaction), it is introduced an intermediary entity that holds some additional information. This information includes all additional tasks that must be performed, except for storing, updating, or deleting resources in the database. This intermediary entity is referred to as _operation_. Operations are also used so that Compass can track the progress of asynchronous requests. The statuses that are used are: _in progress_, _succeeded_, and _failed_.

Operations are modeled as [Custom Resource Definitions](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) (CRD) in Kubernetes, and are created by an entity called Scheduler. For more information, see the section [Scheduler](#scheduler) that follows.

An Operation CRD holds the following information:

* **Specification** of the operation - Describes what needs to be done.
    * **OperationID** - A unique identifier of an operation.
    * **OperationType** - Operation type can be one of the following: create, update, or delete.
    * **OperationCategory** - Semantic category of the operation. It is used to group operations of the same category. For example, GraphQL operations are grouped by the name of the mutation (registerApplication or unregisterRuntime). 
    * **ResourceID** - The ID of the top-level resource that is targeted by the operation (for example, `applicationID`).
    * **ResourceType** - The type of the top-level resource that is targeted by the operation. For example, for GraphQL mutations it is the owning entity; for  `registerApplication` that also contains bundles and APIs it is the `Application` type.
    * **CorrelationID** - A correlation ID that is used to identify the request uniquely. When it is stored in the operation, it allows any follow-on processing of the operation to also reuse the same correlation ID for outbound calls.
    * **WebhookIDs** - A set of webhooks that are run as part of the operation.
    * **RelatedResources** - An array of `ResourceType` and `ResourceID` pairs, related to the top-level resource, which is targeted by the operation.
    * **RequestObject** - A JSON that contains the referenced object details (for example, an application, its ID, name, template ID, etc.), tenant ID of the object, and the headers from the initial incoming request, which can be forwarded when external calls are made.

* **Status** - Used for the operation progress and the final result of it.
    * **Webhooks** - Contains details about the execution of the webhooks that need to finish to ensure a successful operation. 
        * **WebhookID** - A unique webhook identifier.
        * **RetriesCount** - Number of webhook retries so far.
        * **WebhookPollURL** - Determines the status of long running or asynchronous webhooks.
        * **LastPollTimestamp** - A timestamp of the last polling attempt on the status of long running or asynchronous webhooks.
        * **State** - A state of the webhook, such as, _In Progress_, _Success_, or _Failed_.
    * **Condition** - Valid values are: _Ready_ and _Error_. When the operation finishes, its result is `true` or `false`. The result can contain a property `message` that has additional details regarding the status (for example, an error message from the remote call).
    * **Phase** - Operation execution phase. Valid values are: _In Progress_, _Success_, or _Failed_.

### Scheduler
The Scheduler is part of the Director and is used as a GoLang library that can create and update Operation CRDs when an asynchronous operation is processed in Director.

The Scheduler is responsible for checking the progress of the operations. If there is an operation in progress, it cancels the new requests as the previous operation must be finalized before a new one is started.

### Operations Controller
The Operations Controller component is a [Kubernetes controller](https://kubernetes.io/docs/concepts/architecture/controller/) that takes care of _Operation CRDs_ created by the Scheduler.

For more information about its reconciliation loops, see the corresponding documents of the supported operation types.
