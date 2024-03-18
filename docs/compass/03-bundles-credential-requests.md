# Request Credentials for Bundles

In Kyma Runtime, a single Application is represented as a ServiceClass, and a single Bundle of a given Application is represented as a ServicePlan in the Service Catalog. You can create many instances of a Bundle in a single Runtime. For example, in the Kyma Runtime, an instance of a Bundle is represented by a ServiceInstance that can be created for every Namespace.

When provisioning a new ServiceInstance from a Bundle, a Runtime requests API credentials. API credentials for every ServiceInstance are defined on the Bundle level and stored in the Director component. Multiple APIs under the same Bundle share the same credentials.


## Requesting API Credentials Flow

This diagram illustrates the detailed flow of requesting API credentials. The Application provides Webhook API where Compass requests new credentials for the given Bundle.

![API Credentials Flow](./assets/api-credentials-flow.svg)

> **NOTE:** There is an option that the Application does not support Webhook API. That means the Application must monitor registered Bundles and set Bundle credentials when a new instance of a Bundle is created.

Assume we have Application which is already registered in Compass. Application has one Bundle which contains a single API Definition.

1. In the Kyma Runtime, the user creates a ServiceInstance from the Application's Bundle.
1. Runtime Agent calls the Director to request credentials for a given ServiceInstance.
1. Director asynchronously notifies the Application about the credentials request using Webhook API.
1. Application sets credentials for a given ServiceInstance.
1. Runtime Agent fetches configuration with credentials for a given ServiceInstance.

When the user deletes a ServiceInstance, Runtime Agent requests the Director to delete credentials. The Director sets credentials status to `UNUSED`, notifies the Application, and waits for credentials deletion.

### Example Flow of Requesting Credentials

1. User connects the `foo` Application with the single `bar` Bundle which contains few API and Event Definitions. The Bundle has `instanceAuthRequestInputSchema` defined.
1. User selects the `foo` ServiceClass and the `bar` ServicePlan.
1. User provides the required input defined by `instanceAuthRequestInputSchema` and provisions the selected ServicePlan. The ServiceInstance is in the `PROVISIONING` state.
1. Runtime Agent calls the Director with `requestBundleInstanceAuthCreation` and passes the user's input.
1. Director validates the user's input against `instanceAuthRequestInputSchema`. If the user's input is valid, a new `BundleInstanceAuth` is created within the `foo` Bundle.
   
   a. If `defaultInstanceAuth` for the `foo` Bundle is defined, the newly created `BundleInstanceAuth` is filled with credentials from the `defaultInstanceAuth` value. The status is set to `SUCCEEDED`.
   
   b. If `defaultInstanceAuth` for the `foo` Bundle is not defined, the `BundleInstanceAuth` waits in the `PENDING` state until the Application does `setBundleInstanceAuth`. Then, the status is set to `SUCCEEDED`.
   
1. After the Runtime fetches valid credentials for the ServiceInstance, the status of the ServiceInstance is set to `READY`.

### GraphQL Schema

The following snippet describes GraphQL API for Bundles credential requests:

```graphql
type Bundle {
  id: ID!
  # (...)

  """
  Optional JSON schema for validating user's input when provisioning a ServiceClass.
  """
  instanceAuthRequestInputSchema: JSONSchema
  instanceAuth(id: ID!): BundleInstanceAuth
  instanceAuths: [BundleInstanceAuth!]!
  """
  When defined, all requests via `requestBundleInstanceAuthCreation` mutation fallback to defaultInstanceAuth.
  """
  defaultInstanceAuth: Auth
}

type BundleInstanceAuth {
  id: ID!
  """
  Context of BundleInstanceAuth - such as Runtime ID, namespace, etc.
  """
  context: Any

  """
  User input
  """
  inputParams: Any
  
  """
  It may be empty if status is PENDING.
  Populated with `bundle.defaultInstanceAuth` value if `bundle.defaultAuth` is defined. If not, Compass notifies Application/Integration System about the Auth request.
  """
  auth: Auth
  status: BundleInstanceAuthStatus
}

type BundleInstanceAuthStatus {
  condition: BundleInstanceAuthStatusCondition!
  timestamp: Timestamp!
  message: String!
  """
  Possible reasons:
  - PendingNotification
  - NotificationSent
  - CredentialsProvided
  - CredentialsNotProvided
  - PendingDeletion
  """
  reason: String!
}

enum BundleInstanceAuthStatusCondition {
  """
  When creating, before Application sets the credentials
  """
  PENDING
  SUCCEEDED
  FAILED
  """
  When Runtime requests deletion and Application has to revoke the credentials
  """
  UNUSED
}

input BundleInstanceAuthSetInput {
	"""
	If not provided, the status has to be set. If provided, the status condition  must be "SUCCEEDED".
	"""
	auth: AuthInput
	"""
	Optional if the auth is provided.
	If the status condition is "FAILED", auth must be empty.
	"""
	status: BundleInstanceAuthStatusInput
}

input BundleInstanceAuthStatusInput {
	condition: BundleInstanceAuthSetStatusConditionInput! = SUCCEEDED
	"""
	Required, if condition is "FAILED". If empty for SUCCEEDED status, default message is set.
	"""
	message: String
	"""
	Required, if condition is "FAILED". If empty for SUCCEEDED status, "CredentialsProvided" reason is set.
	
	Example reasons:
	- PendingNotification
	- NotificationSent
	- CredentialsProvided
	- CredentialsNotProvided
	- PendingDeletion
	"""
	reason: String
}

input BundleInstanceAuthRequestInput {
	"""
	Context of BundleInstanceAuth - such as Runtime ID, namespace, etc.
	"""
	context: Any
	"""
	JSON validated against bundle.instanceAuthRequestInputSchema
	"""
	inputParams: Any
}

type Mutation {
  """
  When BundleInstanceAuth is not in the pending state, the operation returns an error.

  When used without error, the status of pending auth is set to success.
  """
  setBundleInstanceAuth(authID: ID!, in: BundleInstanceAuthSetInput!): BundleInstanceAuth!
  deleteBundleInstanceAuth(authID: ID!): BundleInstanceAuth!
  requestBundleInstanceAuthCreation(bundleID: ID!, in: BundleInstanceAuthRequestInput!): BundleInstanceAuth!
  requestBundleInstanceAuthDeletion(authID: ID!): BundleInstanceAuth!
}
```

## Passing Additional Input Parameters

You can pass additional input parameters when provisioning a new ServiceInstance from a Bundle. Input parameters are validated against input JSON schema provided in the **instanceAuthRequestInputSchema** field by the Application or Integration System. The parameters, as well as the input JSON schema, are completely optional. As there is no trusted connection between an Integration System and Runtime, additional input parameters have to be passed to the Application or Integration System through the Director.

## Registering Bundle API Without Credentials

In order to register API that does not require credentials, you must set the **defaultInstanceAuth.credential** property to `null`.
