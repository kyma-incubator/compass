# Credential requests for Packages

## Overview

On Runtime, Application is represented as Service Class, and every Package within Application is represented as Service Plan. This document describes credential requests for APIs during the Service Plan provisioning. It also mentions passing optional input parameters from Kyma Runtime to Application or Integration System during the provisioning process.

## Assumptions

- Multiple Service Instances can be created from a given Package within Application.
- During Service Class provisioning, the provided input has to be sent back to Integration System or Application via Compass. The reason is that there is no trusted connection between Integration System and Runtime.
- Passing input parameters is done during requesting credentials for a given Service Instance.

## Details

The Director GraphQL API is updated to store credentials per Service Instance. Credentials for every Instance across all Runtimes are stored on the Package level.

### API Credentials Flow

This diagram illustrates the API credentials flow in details. The Application provides Webhook API where Management Plane requests for providing new credentials for given Package.

>**NOTE:** There is an option that Application does not support Webhook API. That means Application needs to monitor registered API Definitions and set API credentials when new Runtime assigned. The Administrator can exchange credentials for registered APIs at any time too.

![Application Webhook](./assets/api-credentials-flow.svg)

Assume we have Application which is already registered into Management Plane. No Runtimes are assigned yet. Application has one Package which contains single API Definition.

1. The Administrator requests new Runtime with Application via Cockpit.
2. The Cockpit requests configuration for Runtime and the Director asks Application for new credentials.
3. The Cockpit requests Runtime with configuration for Runtime Agent and Runtime Provisioner creates Runtime.
4. The Application sets Package credentials for the particular Service Instance of a given Runtime.
5. The Runtime Agent enables Runtime to call Application APIs.


### GraphQL Schema

```graphql

type PackageDefinition {
	id: ID!
    # (...)

    """
    Optional JSON schema for validation user input while provisioning Service Class.
    """
	authRequestJSONSchema: JSONSchema
	auth(id: ID!): APIInstanceAuth
	auths: [APIInstanceAuth!]!
	"""
	When defined, all Auth requests via `requestAPIInstanceAuthForPackage` mutation fallback to defaultAuth.
	"""
	defaultAuth: Auth
}

type APIInstanceAuth {
	id: ID!
	"""
	Context of APIInstanceAuth - such as Runtime ID, namespace, etc.
	"""
	context: Any
	"""
	It may be empty if status is PENDING.
	Populated with `package.defaultAuth` value if `package.defa	ultAuth` is defined. If not, Compass notifies Application/Integration System about the Auth request.
	"""
	auth: Auth
	status: APIInstanceAuthStatus
}

type APIInstanceAuthStatus {
	condition: APIInstanceAuthStatusCondition!
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

enum APIInstanceAuthStatusCondition {
	"""
	When creating or deleting new one
	"""
	PENDING
	SUCCEEDED
	FAILED
}

type Mutation {
    """
	When APIInstanceAuth is not in pending state, the operation returns error. If the APIInstanceAuth is in "Pending" status, it is set to success.
	"""
	setAPIInstanceAuthForPackage(packageID: ID!, authID: ID!, in: AuthInput! @validate): APIInstanceAuth! 
	deleteAPIInstanceAuthForPackage(packageID: ID!, authID: ID!): APIInstanceAuth!
	requestAPIInstanceAuthForPackage(packageID: ID!, in: APIInstanceAuthRequestInput!): APIInstanceAuth! 
}
```

Application or Integration System can set optional `authRequestJSONSchema` field with a JSON schema with parameters needed during Service Class provisioning. The values provided by User are validated against the JSON schema.

### Example request credentials flow
1. User connects Application `foo` with single Package `bar` which contain few API and Event Definitions. The Package has `authRequestJSONSchema` defined.
1. User selects Service Class `foo` and Service Plan `bar`.
1. User provides required input (defined by `authRequestJSONSchema`) and provisions selected Service Plan.
1. Runtime Agent calls Director with `requestAPIInstanceAuthForPackage`, passing user input.
1. Director validates user input against `authRequestJSONSchema`. When the user input is valid, a new `APIInstanceAuth` within Package `foo` is created.
    a. If `defaultAuth` for Package `foo` is defined, the newly created `APIInstanceAuth` is filled with credentials from `defaultAuth` value. The status is set to `SUCCEEDED`.
    b. If `defaultAuth` for Package `foo` is not defined, the `APIInstanceAuth` waits in `PENDING` state until Application does `setAPIInstanceAuthForPackage`. Then the status is set to `SUCCEEDED`.
1. After fetching valid credentials for Service Instance by Runtime, Service Instance is set to `READY` state.