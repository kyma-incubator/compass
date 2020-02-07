# Input parameters for Packages in the Runtime

## Overview

On Runtime, Application is represented as Service Class, and every Package within Application is represented as Service Plan. This document describes passing optional input parameters from Kyma Runtime to Application or Integration System during Service Class provisioning.

## Assumptions

- Multiple Service Instances can be created from a given Package within Application.
- During Service Class provisioning, the provided input has to be sent back to Integration System or Application via Compass. The reason is that there is no trusted connection between Integration System and Runtime.
- Passing input parameters is done during requesting credentials for a given Service Instance.

## Solution

The Director GraphQL API is updated to store credentials per Service Instance. Credentials for every Instance across all Runtimes are stored on the Package level.

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
	Populated with `package.defaultAuth` value if `package.defaultAuth` is defined. If not, Compass notifies Application/Integration System about the Auth request.
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