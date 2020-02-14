# Request credentials for Packages

In Kyma Runtime, a single Application is represented as a ServiceClass, and a single Package of a given Application is represented as a ServicePlan in the Service Catalog. You can create many instances of a Package in a single Runtime. For example, in the Kyma Runtime, an instance of a Package is represented by a ServiceInstance that can be created for every Namespace.

When provisioning a new ServiceInstance from a Package, a Runtime requests API credentials. API credentials for every ServiceInstance are defined on the Package level and stored in the Director component. Multiple APIs under the same Package share the same credentials.


## Requesting API credentials flow

This diagram illustrates the detailed flow of requesting API credentials. The Application provides Webhook API where Compass requests new credentials for the given Package.

![API Credentials Flow](./assets/api-credentials-flow.svg)

> **NOTE:** There is an option that the Application does not support Webhook API. That means the Application must monitor registered Packages and set Package credentials when a new instance of a Package is created.

Assume we have Application which is already registered in Compass. Application has one Package which contains a single API Definition.

1. In the Kyma Runtime, the user creates a ServiceInstance from the Application's Package.
1. Runtime Agent calls the Director to request credentials for a given ServiceInstance.
1. Director asynchronously notifies the Application about the credentials request using Webhook API.
1. Application sets credentials for a given ServiceInstance.
1. Runtime Agent fetches configuration with credentials for a given ServiceInstance.

When the user deletes a ServiceInstance, Runtime Agent requests the Director to delete credentials. The Director sets credentials status to `UNUSED`, notifies the Application, and waits for credentials deletion.

### Example flow of requesting credentials

1. User connects the `foo` Application with the single `bar` Package which contains few API and Event Definitions. The Package has `instanceAuthRequestInputSchema` defined.
1. User selects the `foo` ServiceClass and the `bar` ServicePlan.
1. User provides the required input defined by `instanceAuthRequestInputSchema` and provisions the selected ServicePlan. The ServiceInstance is in the `PROVISIONING` state.
1. Runtime Agent calls the Director with `requestPackageInstanceAuthCreation` and passes the user's input.
1. Director validates the user's input against `instanceAuthRequestInputSchema`. If the user's input is valid, a new `PackageInstanceAuth` is created within the `foo` Package.
   
   a. If `defaultInstanceAuth` for the `foo` Package is defined, the newly created `PackageInstanceAuth` is filled with credentials from the `defaultInstanceAuth` value. The status is set to `SUCCEEDED`.
   
   b. If `defaultInstanceAuth` for the `foo` Package is not defined, the `PackageInstanceAuth` waits in the `PENDING` state until the Application does `setPackageInstanceAuth`. Then, the status is set to `SUCCEEDED`.
   
1. After the Runtime fetches valid credentials for the ServiceInstance, the status of the ServiceInstance is set to `READY`.

### GraphQL schema

The following snippet describes GraphQL API for Packages credential requests:

```graphql
type Package {
  id: ID!
  # (...)

  """
  Optional JSON schema for validating user's input when provisioning a ServiceClass.
  """
  instanceAuthRequestInputSchema: JSONSchema
  instanceAuth(id: ID!): PackageInstanceAuth
  instanceAuths: [PackageInstanceAuth!]!
  """
  When defined, all requests via `requestPackageInstanceAuthCreation` mutation fallback to defaultInstanceAuth.
  """
  defaultInstanceAuth: Auth
}

type PackageInstanceAuth {
  id: ID!
  """
  Context of PackageInstanceAuth - such as Runtime ID, namespace, etc.
  """
  context: Any

  """
  User input
  """
  inputParams: Any
  
  """
  It may be empty if status is PENDING.
  Populated with `package.defaultInstanceAuth` value if `package.defaultAuth` is defined. If not, Compass notifies Application/Integration System about the Auth request.
  """
  auth: Auth
  status: PackageInstanceAuthStatus
}

type PackageInstanceAuthStatus {
  condition: PackageInstanceAuthStatusCondition!
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

enum PackageInstanceAuthStatusCondition {
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

input PackageInstanceAuthSetInput {
	"""
	If not provided, the status has to be set. If provided, the status condition  must be "SUCCEEDED".
	"""
	auth: AuthInput
	"""
	Optional if the auth is provided.
	If the status condition is "FAILED", auth must be empty.
	"""
	status: PackageInstanceAuthStatusInput
}

input PackageInstanceAuthStatusInput {
	condition: PackageInstanceAuthSetStatusConditionInput! = SUCCEEDED
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

input PackageInstanceAuthRequestInput {
	"""
	Context of PackageInstanceAuth - such as Runtime ID, namespace, etc.
	"""
	context: Any
	"""
	JSON validated against package.instanceAuthRequestInputSchema
	"""
	inputParams: Any
}

type Mutation {
  """
  When PackageInstanceAuth is not in the pending state, the operation returns an error.

  When used without error, the status of pending auth is set to success.
  """
  setPackageInstanceAuth(packageID: ID!, authID: ID!, in: PackageInstanceAuthSetInput!): PackageInstanceAuth!
  deletePackageInstanceAuth(packageID: ID!, authID: ID!): PackageInstanceAuth!
  requestPackageInstanceAuthCreation(packageID: ID!, in: PackageInstanceAuthRequestInput!): PackageInstanceAuth!
  requestPackageInstanceAuthDeletion(packageID: ID!, authID: ID!): PackageInstanceAuth!
}
```

## Passing additional input parameters

You can pass additional input parameters when provisioning a new ServiceInstance from a Package. Input parameters are validated against input JSON schema provided in the **instanceAuthRequestInputSchema** field by the Application or Integration System. The parameters, as well as the input JSON schema, are completely optional. As there is no trusted connection between an Integration System and Runtime, additional input parameters have to be passed to the Application or Integration System through the Director.
