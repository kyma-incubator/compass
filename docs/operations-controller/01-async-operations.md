# Asynchronous Operations in Compass

In order to increase the response time for requests in Compass, a new asynchronous mechanism for required external calls is introduced.

## Use Cases

Currently the following use cases exist for enabling asynchronously APIs for CMP and respectively for having out-of-transaction logic executed as part of the API business logic:

* Register application → fetch ORD documents from the application
* Request one time token for pairing from remote token issuer service
* Delete application → Notify the integration system or the LoB application that it has to clean up any artifacts that were created specifically for this application pairing
* Request credentials for a particular usage of the APIs in a application bundle → notify integration application or LoB application that a new set of credentials is needed
* Delete credentials for a particular usage of the APIs in a application bundle →  notify the integration application or LoB application that a set of credentials is no longer needed
* Fetch requests for fetching specs for apiDefinitions and eventDefinitions

## Asynchronous Flow

### Operations

In order to achieve delayed execution of part of the business logic of the API (for example running remote calls separately from the transaction) we will introduce a new intermediary entity that holds information about what more should be done apart from storing/updating/deleting the resource in the database. We will refer to this entity as _operation_. Operations will also be used so that Compass can track the progress of asynchronous requests (_in progress_, _succeeded_, _failed_).

Operations are modeled as [Custom Resource Definitions](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) in Kubernetes, and are created by an entity called Scheduler - more details on the Scheduler can be found in the [next section](#scheduler).

An operation CRD holds the following information:

* **Specification** of the operation - describes what needs to be done
    * **OperationID** - uniquely identifies an operation
    * **OperationType** - Create, Update, Delete
    * **OperationCategory** - semantically categorizes the operation (puts it in one "bucket" with other operations of the same category) - for GraphQL operations this would be the name of the mutation (e.g. registerApplication, unregisterRuntime) 
    * **ResourceID** - the ID of the top level resource targeted by this operation (e.g. `applicationID`)
    * **ResourceType** - the type of the top level resource targeted by this operation (for GraphQL mutations this would be the owning entity (e.g. for `registerApplication` that also contains bundles and APIs, this would be the `Application` type)
    * **CorrelationID** - the correlation ID that was used for uniquely identifying this request. Storing it in the operation allows for whoever processes the operation later on, to also reuse the same correlation ID for outbound calls
    * **WebhookIDs** - a set of webhooks that should be ran as part of this operation
    * **RelatedResources** - an array of pairs of `ResourceType` and `ResourceID` that are related to the top level resource that is targeted by the operation
    * **RequestObject** - JSON containing the referenced object details (e.g. application with its ID, name, template ID, etc.), tenant ID of the object, and also the headers from the initial incoming request, which can be forwarded when external calls are made.

* **Status** - describes how far the operation is with the execution and what is the final result
    * **Webhooks** - contains details about the execution of the webhooks that need to finish in order for the operation to be successful
        * **WebhookID**
        * **RetriesCount** - how many times the webhook was retried so far
        * **WebhookPollURL** - for determining the status of long running/async webhooks
        * **LastPollTimestamp** - the last time polling the status for long running/async webhooks was attempted
        * **State** - _Success_, _Failed_, _In Progress_
    * **Condition** - _Ready_, _Error_ - when the operation is finished one will be `true` and the other - `false`, each of them can contain `message` property with additional details regarding the status (e.g. an error message from a remote call)
    * **Phase** - _Success_, _Failed_, _In Progress_

### Scheduler
The Scheduler is currently a part of Director and is used as a GoLang library that can create and update Operation CRDs when an asynchronous operation is being processed in Director.

The Scheduler is responsible for checking if there is an operation currently in progress, and if it is, it cancels the request - the previous operation should be finalized before a new one is started.

### Operations Controller
The Operations Controller component is a [Kubernetes controller](https://kubernetes.io/docs/concepts/architecture/controller/) that takes care of _Operation CRDs_ created by the Scheduler.

More details on its reconciliation loops can be found in the respective documents for the supported operation types.
