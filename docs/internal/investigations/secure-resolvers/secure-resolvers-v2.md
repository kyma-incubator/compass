# Limiting Access to GraphQL Resources 

This document describes a possible solution to the problem of unlimited resource access. The problem that needs to be corrected is the current possibility of one application in tenant X to modify the metadata of another application in the same tenant X. This problem is also valid for runtimes. Furthermore, integration systems have global access to all data in Compass. An optimal solution features an integration system that is limited to access only its own resources (newly created in the system) or resources, to which the system has granted access already.

The solution proposal includes limiting the access to operations for `applications`, `application_templates`, `runtimes`, and `integration systems`.

## Terminology

* API Consumer - An API Consumer could be any of the following: application, runtime, integration system, or user.
* Owner (or owning entity) - A top-level entity to which a given resource belongs. For example, an application is the owner of packages, api definitions, documents, and so on. In this context, the following owner types are available: `application`, `runtime`, `integration system`, and `application template`.

## Requirements

When a user makes a query or mutation, the secured resolvers concept should not limit the user's access. Put simply, the user consumer type should not be affected by the secured resolvers concept.

By default, an application must not be allowed to modify another application's metadata within the same tenant. This applies not only for the top-level application entity but also for all its related sub-entities such as: packages, API and event definitions, documents, and so on. For example, application X must not be allowed to insert API definitions in package Y when package Y belongs to application Z, even if both applications X and Z are in the same tenant.

By default, a runtime must not be allowed to modify another runtime's metadata within the same tenant. Additionally, a runtime must not be allowed to fetch applications that belong to other runtimes.

By default, an integration system must not be allowed to modify another integration system's metadata.

In general, integration systems can manage other systems and their metadata (such as applications, and runtimes). However, an integration system must be allowed to modify only entities that it has created, or entities to which it has been granted access for modifications. For example, if an integration system registers an application, later on, it can  also register packages for this application. However, it cannot modify metadata for applications that other consumers registered, unless it has been granted access explicitly to do so. Another example is when a user registers an application via the UI. To get access to this application, the integration system X either must be granted this access by the user via an API call to Director (grant access to the integration system for this application), or some automated procedures must be in place (a concept about stacking credentials via one time tokens is available in the further sections of this document). Although the given examples are related to applications, these requirements are also valid for application sub-entities, application templates, and runtimes managed by integration systems.

Some integration systems (such as UI applications) require unrestricted access to the Director API. They should keep working properly having a global view/access.

An implicit requirement, derived from the requirements above, is that no system must be allowed to list all applications, runtimes, or integration systems.

Out-of-scope: Limit access to the `auths` field of each type to admin users only.

## Current State

The following section analyzes the problem in all current queries and mutations, and groups them in sections providing a conceptual solution.

The problem with securing resolvers is three dimensional. It features all three: mutations/queries; consumer, and owner. In total, there are about 40 mutations/queries that have to be secured, 4 consumer types, and 4 owner types. For specific queries/mutations there is always a single owner but obtaining the ID of that owner is not always straight forward. For example, in getApplication(applicationID) the owner ID is passed as input to the query, however, in updatePackage(packageID) some special code needs to be executed to find the application ID from the package id.

Why it is important who is the owner and what is the owner ID?

Currently, with the help of the ORY integration, the Director API always recieves an ID token, regardless of the actual authentication mechanism. Then, the details about the actual caller/consumer can be extracted from the ID token, and then, stored in the Go context. Therefore, on a very high level, it is necessary to ensure that the consumer from the Go context matches the owner of the requested resource. More broadly put (and required especially for integration systems), it is important to verify that the consumer can access the owner's resources. For example, if the owner is an application, the resources that must be accessible are the application details, packages, and so on.

### Determining the Owner ID

A determined owner ID is convenient for the use case of matching consumer and owner types. For example, a given application tries to update an api definition and the owner ID of the api definition is determined. In this case the owner is the application to which this api definition belongs. Then, the comparison of the owner ID and the consumer ID controls the update request and allows the operation or not. More complicated use cases, where the consumer and the owner types differ, require different approach and solution.

Determining the owner ID during queries and mutations was researched thoroughly by analysing each query and mutation. Only the results of this research are provided in this document, and are grouped in sub-sections by the type of the owner:

#### Owner Type: Application

Queries and mutations which owner is an application are grouped in 8 categories, depending on the input provided to the query or mutation. In each category a different piece of source code must be carried out to determine the owner id.

1. Queries and mutations that provide the owner ID as part of the client input (for example: `getApplication`, `updateApplication`, `addPackage`, etc). For these, the owner ID can be determined from the GraphQL arguments list.
2. Queries and mutations that provide a `packageID` as part of the client input (for example: `updatePackage`, `deletePackage`, `addDocumentToPackage`, etc). For these, the owner ID can be determined by fetching the actual package with ID equal to the `packageID` from the GraphQL arguments list and getting its `applicationID`.
3. Queries and mutations that provide the `documentID` as part of the client input (for example: `updateDocument`, `deleteDocument`). To get the owner, fetch the document with ID equal to the `documentID` from the GraphQL arguments list, then get its `packageID`, then fetch the package, and finally, get its `applicationID`, which is the owner ID.
4. Queries and mutations that provide `APIDefinitionID` as part of the client input (for example: `updateAPIDefinition`, `deleteAPIDefinition`). To get the owner, fetch the api definition with ID equal to the `APIDefinitionID` from the GraphQL arguments, then get its `packageID`, then fetch the package, and finally, get its `applicationID`, which is the owner ID.
5. Queries and mutations that provide `EventDefinitionID` as part of the client input (for example: `updateEventDefinition`, `deleteEventDefinition`). To get the owner, fetch the event definition with ID equal to the `EventDefinitionID` from the GraphQL arguments, then get its `packageID`, then fetch the package, and finally, get its `applicationID`, which is the owner ID.
6. Queries and mutations that provide `webhookID` as part of the client input (for example: `updateWebhook`, `deleteWebhook`). To get the owner, fetch the webhook with ID equal to the `webhookID` from the GraphQL arguments, and then get its `applicationID`, which is the owner ID.
7. Queries and mutations that provide `systemAuthID` as part of the client input (for example: `deleteSystemAuthForApplication`). To get the owner, fetch the system auth (or a record from a different table that contains both `applicationID` and `system_auth_id`), and then, get the `application_id`, which is the owner ID.
8. Queries and mutations that provide `packageInstanceAuthID` as part of the client input (for example: `setPackageInstanceAuth`, `deletePackageInstanceAuth`). To get the owner, fetch the package by `packageInstanceAuthID` taken from the GraphQL arguments, and then, get the package `applicationID`, which is the owner ID.

#### Owner Type: Runtime

1. Queries and mutations that provide the owner ID as part of the client input (for example: `getRuntime`, `updateRuntime`, `deleteRuntime`, etc). For these, the owner ID can be determined from the GraphQL arguments list.
2. Cases in which the owner is of a different type (for example: `applicationsForRuntime`). For these, the `runtimeID` is provided as input and it can be used as ownerID to be compared against the consumerID in case the consumer is of type Runtime. However, additional checks based on scenarios are also performed, implented within the resolver code.

Runtimes consuming APIs owned by different types (for example: `query application`, `requestPackageInstanceAuthCreation`) are outlined in a later section of this document.

#### Owner Type: Integration System

1. Queries and mutations that provide the owner ID as part of the client input (for example: `getIntegrationSystem`, `updateIntegrationSystem`, `deleteIntegrationSystem`, etc). For these, the owner ID can be determined from the GraphQL arguments list.

Integration systems consuming APIs owned by different types are outlined in a later section of this document.

#### Owner Type: Application Template

In the context of application templates, by default, only the system which creates them must be allowed to modify them, and no one else, except for Users.

This owner type differs from the rest as there is no consumer of this type. This means that obtaining the `owner_id` does not facilitate the scoping of the access. In this case, if an integration system X created template Y, it is required a different relation, such as an `integration_system_id` column in the `app_templates` table. In fact, this is the use case when the consumer is an integration system for all owner types. The fact that `owner_id` is not sufficient in this case is outlined in more detail in the next section.

#### Summary

It is unclear why the design requires the ownerID to be passed the way it is described in all first use cases of the previous sub-sections. When the owner matches the consumer, and the consumed API is directly identified by the owner ID (for example, when getApplication(appID) is called by the actual application); it is sufficient for the caller to provide its authentication details, from which the consumer ID is extracted. Forcing the caller to provide the owner ID as a GraphQL argument adds additional complexity to the API logic. That is, the API has to verify that the consumer provided an owner ID that matches its consumer ID. However, it is decided to not change the APIs as part of this design, so such extra check will remain and it could be optimized as part of a future, more generic solution.

If the storage layer is extended with custom queries, it must be ensured that no more than 1 DB query is required for the calculation of the ownerID. If the currently provided queries remain in use, there are cases when the calculation of ownerID requires 2 or 3 queries. For example, `updateAPIDefinition` requires to fetch the API definition and to fetch the package. If the consumer is an integration system, it might be needed to fetch the whole application entity, which is a third query.

Another approach to determine the owner ID is to store it in each table. That is, each table, which is directly or transitively related to applications must store application_id. However, this adds complexity to the database schema. Additionally, the logic for calculating the owner ID would still be required but it will be implemented during the creation of the resources. Thus, duplicating the owner ID column everywhere does not simplify the overall implementation.

This section analysed the APIs from the owner perspective. Basically, when the consumer and the owner types match, it is possible to calculate the `ownerID` and compare it to the `consumerID`. When they match, the request is allowed to proceed. The use cases that feature different consumer types are more complex and are outlined in the following sections.

### Analyzing Different Consumer Types

When consumer and owner types match, determination of owner ID is straight forward and is outlined in the previous sections. However, not always the consumer and owner types match, which requires different mechanisms and different approach to the problem. For example, runtimes can consume APIs where the owner type is an application and the integration systems can consume APIs where the owner type can vary.

A determined owner ID simplifies solving the problems related to use cases, such as: allowing only application "X" to update its metadata; or allowing only application "X" to update its packages; or limiting runtime "R" to be the only runtime allowed to fetch its applications; or allowing only integration system "I" to update its own metadata. Anyway, there is more complexity when the consumer is of type integration system. The following section provides a quick analysis of the various use cases of diffrent consumer types (the consumer details are provided in the Go context).

#### Consumer Type: User

No further restrictions apply.

#### Consumer Type: Application

There is no valid use case for an application to consume APIs that belong to other owner types. Therefore, the analysis in the Owners section above is sufficient.

#### Consumer Type: Runtime

Runtimes consume a few queries and mutations that belong to owner type application.

* `query application`, `mutation requestPackageInstanceAuthCreation`, `mutation requestPackageInstanceAuthDeletion` - there should be a scenario check, implemented within the resolver code. Currently, such scenario check is missing. Therefore, instead of implementing a scenario check in all these places, it is better to try to fit these in the generic design of secured resolvers.
* `query applicationsForRuntime` - It already performs a scenario check implemented within the code. However, as  mentioned already in the document, this query also accepts a `runtime_id` as input parameter and only an "ownerID equals consumerID" check must be added.

#### Consumer Type: Integration System

The `ownerID` check, described above, does not help when the consumer type is integration system and owner type is any of the following: application, runtime, or application template. In this case, it is needed to develop some associations between these entities and the integration system. Instead of calculating the ownerID, it is needed to get the owner entity and get its integration system id, and then, compare it with the `consumerID`.

## Solution Proposal

The purpose of the proposal is to provide a uniform, generic solution that solves all of the above cases. Instead of modeling a design about which consumer (application, integration system, or runtime) can access what owning entity (application, application template, runtime, or integration system) and its sub-entities; it is better to introduce an abstraction of the consumer and replace it by something that all technical consumers have in common, that is the system auth. Therefore, instead of determining what is the consumer type and based on it proceeding with the comparison of its comsumer id; it is better to check what access has been granted to the system auth in relation to this consumer. This abstraction of `system auth accesses` disregards the consumer type and performs access checks against the `system_auth_id` of the consumer, instead of the `consumer_id`. Basically, when a request is received, it can be checked if the consumer system (represented by the `system_auth_id` from the ID token) has been granted access to the owner entity. The owner entity ID needs to be calculated as described previously in this document. There are several inputs for granting access to consumers for an owner entity. Most of these are automated as part of other flows. `System auth accesses` is stored in a new table. Each row contains information about a `system_auth_id` and the relevant `application_id`, `runtime_id`, `integration_system_id`, or `application_template_id`, to which this `system_auth_id` has access.

### Add Additional Details in the ID Token

Currently, the ID token created via data, provided by the tenant mapping handler, contains details about `consumer_type` and `consumer_id`. It will be extended to include also details about `consumer_level` and `system_auth_id`. The level of access of the consumer (`RESTRICTED` or `UNRESTRICTED`) is represented by the `consumer_level`, and `system_auth_id` is the ID of the system auth that the consumer is authenticating with.

Tenant Mapping Handler grants `UNRESRICTED` consumer level access to Users and to special or legacy integration systems, such as, UIs that require global access. To model such systems, an additional column (with values `RESTRICTED` and `UNRESTRICTED`) must be added to the `system_auths` table.

### System Auth Access

System auth access (or system auth restrictions) represents the permissons, granted to a specific set of system credentials. For example, if integration system "X" uses client credentials, represented by system_auth_id=123; a record in `system_auth_acccess` with system_auth_id=123 and app_id=456 means that the consumer (integration system) can access and modify metadata and sub-entities of an application with id=456.

![System Auths Access Table](./assets/system-auth-access.png)

### limitAccess Directive

It is better to do the checks about who is allowed to access what centrally. Preferably, before the actual resolver business logic of the query or mutation. Since the relevant data about access permissions is stored in one table, it is possible to design a directive that can perform the checks.

```graphql
directive @limitAccess(ownerProvider: String!, idField: String!) on FIELD_DEFINITION
```

* `ownerProvider` - A key that describes how the ownerID of the owner entity can be obtained.
* `idField` - The name of the GraphQL query or mutation argument that is used to calculate the ownerID.
* The directive structure contains a map of provider functions with keys `ownerProvider` values. These functions contain the logic how to calculate the owner ID for the respective query or mutation where the directive is specified.

#### Proces Flow

The proposed directive does the following:

1. Loads consumer information from context. Consumer information contains `consumer_level` and `system_auth_id`.
2. Checks if `consumer_level` is `UNRESTRICTED` and allows access when it is true.
3. Gets the GraphQL argument represented by the value of `idField`.
4. Looks into the map of owner providers and finds the owner provider function represented by the `ownerProvider` key, specified in the directive.
5. Executes the provider function by passing the argument obtained from the `idField` and gets the owner entity details.
6. After obtaining the ownerProvider ID in the previous step, checks if a record in `system_auth_access` exists for this owner ID and the `system_auth_id` that is part of the consumer information.
7. If the check is true, the request is allowed. Otherwise, if the check is false, the request is stopped.

#### Extending the Database Layer

Steps 5 and 6 in the process flow above can be merged if the `system_auth_access` check is implemented within the provider function by joining the `system_auth_access` table to the underlying database query, executed by the provider function. To do this, some custom database queries will be needed.

#### Examples

```graphql
type Query {
    application(id: ID!): Application @hasScopes(path: "graphql.query.application") @limitAccess(ownerProvider: "GetApplicationID", idField: "id")
}
```

* Value of `idField` is "id" because this is the name of the GraphQL argument.
* In the directive source code `ownerProvider`, "GetApplicationID" matches to a function that executes a DB query which calculates the `owner_id` (in this case the ID is part of the input) and checks whether the relevant system_auth_access exists.

The actual provider function that the directive executes looks like the following:

```go
providers: map[string]func(ctx context.Context, tenantID string, id string, authID string) (bool, error){
			"GetApplicationID": func(ctx context.Context, tenantID string, id string, authID string) (bool, error) {
				consumerInfo, err := LoadFromContext(ctx)
				if err != nil {
					return false, errors.New("error missing consumer info")
				}

				// performance optimization, db call below would work too
				if consumerInfo.ConsumerType == Application {
					return consumerInfo.ConsumerID == id, nil
				}

				return appRepo.ExistsApplicationByIDAndAuthID(ctx, tenantID, id, authID)
            },
  // Other providers go here
}
```

```go
func (r *pgRepository) ExistsApplicationByIDAndAuthID(ctx context.Context, tenantID, id, authID string) (bool, error) {
	stmt := fmt.Sprintf(`SELECT 1
FROM %s AS a 
JOIN %s AS s on s.%s=a.id 
WHERE a.tenant_id=$1 AND a.id=$2 AND s.%s=$3`,
		applicationTable,
		systemAuthRestrictionsTable,
		applicationRefField,
		systemAuthRefField)

	return r.exists(ctx, stmt, tenantID, id, authID)
}
```

The following example is more complex:

```graphql
updateEventDefinition(id: ID!, in: EventDefinitionInput! @validate): EventDefinition! @hasScopes(path: "graphql.mutation.updateEventDefinition") @limitAccess(ownerProvider: "GetApplicationIDByEventDefinitionID", idField: "id")
```
* `idField` matches the name of the GraphQL argument for the event definition id
* `GetApplicationIDByEventDefinitionID` key matches a function that executes the following database query:

```go
func (r *pgRepository) ExistsApplicationByEventDefinitionIDAndAuthID(ctx context.Context, tenantID, id, authID string) (bool, error) {
	stmt := fmt.Sprintf(`SELECT 1 
FROM %s AS a 
JOIN %s AS p on a.id=p.%s 
JOIN %s AS d on p.id=d.%s
JOIN %s AS s on s.%s=a.id
WHERE a.tenant_id=$1 AND d.id=$2 AND s.%s=$3`,
		applicationTable,
		packageTable,
		applicationRefField,
		eventDefinitionsTable,
		packageRefField,
		systemAuthRestrictionsTable,
		applicationRefField,
		systemAuthRefField)

	return r.exists(ctx, stmt, tenantID, id, authID)
}
```

Note how the query simultaneously calculates the owner ID (application id) and checks the necessary `system_auth_access`, and then, simply returns a boolean result whether access has been granted or not.

The actual provider function is pretty simple as no external performance optimizations are needed:

```go
"GetApplicationIDByAPIDefinitionID": appRepo.ExistsApplicationByAPIDefinitionIDAndAuthID,
```

Finally, the following example shows a reference implementation for the directive logic:

```go
func (d *limitAccessDirective) LimitAccess(ctx context.Context, _ interface{}, next graphql.Resolver, key, idField string) (interface{}, error) {
	ctx = persistence.SaveToContext(ctx, d.db)

	consumerInfo, err := LoadFromContext(ctx)
	if err != nil {
		return nil, errors.New("error missing consumer info")
	}

	if consumerInfo.ConsumerLevel == Unrestricted {
		return next(ctx)
	}

	if consumerInfo.SystemAuthID == "" {
		return nil, errors.Errorf("system auth id not found in consumer context for consumer type %s", consumerInfo.ConsumerType)
	}

	fieldContext := graphql.GetFieldContext(ctx)
	inputID := fieldContext.Args[idField].(string)

	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	providerFunc, found := d.providers[key]
	if !found {
		return nil, fmt.Errorf("owner provider not found for key %s", key)
	}

	exists, err := providerFunc(ctx, tenantID, inputID, consumerInfo.SystemAuthID)
	if err != nil {
		return nil, errors.Wrapf(err, "could not provide owning entity")
	}

	if !exists {
		return nil, apperrors.NewInvalidOperationError(fmt.Sprintf("consumer of type %s with id %s is not allowed to access the requested resource",
			consumerInfo.ConsumerType, consumerInfo.ConsumerID))
	}

	return next(ctx)
}
```

### Inputs for Creating System Auth Access

It is important to mention when the actual records in `system_auth_restrictions` are created. There are several existing flows that are extended to automatically insert `system_auth_restrictions`.

1. During the creation of applications, runtimes, integration systems, and application templates, when the consumer type is NOT User, a record has to be inserted so that later on the consumer system can access what it has created.
2. During issuing system credentials for applications, runtimes, and integration systems, a record has to be inserted so that later on the relevant system can access its own metadata.
3. Out-of-scope: It is possible to also model scenario labeling as another input for `system_auth_access` creation. This would allow us to not necessarily implement scenario checks within some specific directives but rather use the system access check that is part of the `limitAccess` directive.

Also, two new GraphQL mutations are introduced to allow administrator users to grant and revoke access for systems. The new mutations are protected with a new scope `system_access:write` that is given to administrator users.

```graphql
input SystemAuthAccessInput {
	to: SystemAuthAccessToInput!
	for: SystemAuthAccessForInput!
}

input SystemAuthAccessForInput {
	applicationID: String
	applicationTemplateID: String
	runtimeID: String
	integrationSystemID: String
}

input SystemAuthAccessToInput {
	applicationID: String
	runtimeID: String
	integrationSystemID: String
}

type SystemAuthAccessFor {
	applicationID: String
	applicationTemplateID: String
	runtimeID: String
	integrationSystemID: String
}

type SystemAuthAccessTo {
	applicationID: String
	runtimeID: String
	integrationSystemID: String
}

type SystemAuthAccess {
	to: SystemAuthAccessTo!
	for: SystemAuthAccessFor!
}

grantSystemAccess(in: SystemAuthAccessInput! @validate): SystemAuthAccess @hasScopes(path: "graphql.mutation.grantSystemAccess")

revokeSystemAccess(in: SystemAuthAccessInput! @validate): SystemAuthAccess @hasScopes(path: "graphql.mutation.revokeSystemAccess")
```

A relevant use case for the API would be to have a user who registers an application, and then, the user can grant access to an integration system to manage the application details further.

### Additional List Scope

New scope is introduced to avoid systems to list all applications, runtimes, or integration systems. This new scope will be assigned to administrator users only, as follows:

* integration_system:list - for admin user
* runtime:list - for admin user
* application:list - for admin user

A possible improvement would be to grant list access to systems, too, but filter the response based on `system_auth_access` so that a system can list only what it is authorized to access.

### Benefits

1. There are foreign key constraints for everything related to the `system_auth_access` table and no custom exists checks are needed. Also, cascade deletions can be leveraged and no custom implementation in deletion flows in needed.  
2. Logic in grant access API is relatively simple.
3. Amount of data in `system_auth_access` is relatively small. 

### Drawbacks

1. Since it is introduced an abstraction of the the consumer dimension, by using its system_auth/credentials, when a consumer gets a second set of credentials, by default, this second set of credentials is not authorized to access what the first set can access. To grant the second set of credentials the same access as the first set, the records from `system_auth_access` of the first `system_auth_id` must be copied manually to the second `system_auth_id`. This can be mitigated by allowing a single set of credentials per system (which is natural and correct). Also, it can be considered as a security improvement in the case when somebody manages to request a second set of credentials for an existing system, they will not be able to tamper with the data, which the system can actually access.
2. While a design with a single directive is achieved, its logic still contains a degree of complexity, due to the concept of provider functions and the required provider key input parameter. 

## Variations to the `system_auth_acesss` Table

### Modeling Consumer to Resource Access 

One possible variation is to rename the table to `resource_access` and store information in the following columns: `consumer_id`, `consumer_type`, `resource_id`, and `resource_type`. The first two columns (`consumer_id` and `consumer_type`) specify what is the consumer (application, integration system, runtime) that can access the resource. The second two columns (`resource_id` and `resource_type`) specify which is the actual resource that can be accessed (also it can be in any other table in the database).

This approach suggests storing information about who can access every single resource. To control this information, access filters can be modeled and added in the Go context. These access filters can be used to extend the select database query with a join and a where clause. This way it is not needed to do DB calls in the directive. The directive  just adds access filters to the go context.

#### Benefits:

1. The directive logic is simplified. It does not need a map of providers but can just execute a common query for all cases, as follows:
```sql
SELECT 1 
FROM resource_access 
WHERE consumer_id='consumer-id-from-ctx' 
    AND consumer_type='consumer-type-from-ctx' 
    AND resource_id='input-from-graphql-args'
```

2. In addition, the concern in the above approach, about creating a second set of credentials (for a system) that does not have the same access rights, is solved. This solution does not require additional implementation as it does not relate to system_auths.

#### Drawbacks 

1. The table will become very big. It will contain as much rows as the sum of all rows, of all other tables multiplied by the number of consumers.
2. Inserting in the table will happen in each create, register, and add resolver. Additionally, in some resolvers, such as RegisterApplication, which create additional related resources as part of the resolver logic (`createRelatedResources` method), there will be inserts to the `consumer_access` table. Basically, each Repository.Create method will be modified to insert in the `consumer_access` table too.
3. The new API `grant_consumer_access` will be hard to implement as if we are granting access to integration system for application and all its sub-entities. To insert relevant records to `consumer_access` with `resource_id` values, custom logic will be needed to verify that the app exists and find all sub-entity IDs. This custom logic will be different for runtimes and integration systems, and as a result, this API has the potential to become too complex.
4. Because the ID columns can represent multiple different entities, there cannot be foreign key constraints to verify that the consumer and resource actually exist, and to allow cascade deletion when the consumer or resource are deleted respectively. This means that during the deletion of a resource it is needed a custom logic to find and delete all of its `related` records in the `consumer_access` table. Basically, each Repo.Delete method also needs to check and delete relevant related `consumer_access` records. 
5. It is an open question how to model UNRESTRICTED access for global integration systems such as UI, KEB, and provisioner. 

### Modeling Consumer to Owner Access

Having access data stored for every single resource is unnecessary. In this approach, the stored access information is limited only to the owners.
Here, both models from the previous sections cross and the result is a table `consumer_access` with columns `consumer_id`, `consumer_type`, `owner_id`, and `owner_type`.

#### Benefits

1. This resolves the drawback of having multiple credentials for a system in the first approach.
2. The actual table contains less rows than the previous two approaches. 
3. The grant access API is simpler to implement compared to the one in the second approach. However, it is a bit more complex than the one in the first approach, because it is needed to check if the consumer and owner exist (due to lack of FK constraints).

#### Drawbacks

1. Similarly to the first approach, it still has a more complex directive with provider functions. 
2. Similarly to the second approach, usage of foreign keys is not possible. Therefore, custom validation for exists in grant access API and custom delete logic during the deletion of an owner resource is still needed. In this case, it is not needed to modifiy the deletion of all resources but only for the 4 owner types.
3. It is an open question how to model UNRESTRICTED access for global integration systems such as UI, KEB, and provisioner. Perhaps an empty owner is a solution, however, this is neither typical, nor approved practice.

## Additional Research Approaches

This section outlines several additional approaches that were taken into account during the proof of concept (PoC) phase. If they are considered in the future, additional, more thorough proof of concept can be planned and the designs can de discussed in a broader round.

* Usage of helper functions and custom implementaion in each resolver.
* Copying owner_id and integration_system_id to each table. This way, the owner ID does not need to be calculated in each mutation. Usage of a field criteria approach in which the actual business logic query will have a specific where clause, so that it can decide whether access is granted or not. Basically, a field filter/criteria will be appended by some middleware, and this field criteria will specify values for owner_id and integration_system_id. This filter/criteria is eventually included as part of the where clause of the DB query.
* Introducing a generic middleware that swithces over all mutations and queries, and executes the relevant custom logic.
* Require an owner ID in each mutation on top of the already required input, so that the owner ID does not need to be calculated separately on each mutation.
* Add an ownerReference resolver on each input that returns details about the owning entity. Via a special GraphQL field interceptor/middleware add an additional field to the input so that the resolver is called. In the same or in a different middleware, add logic to compare the owner and the consumer.
* Add support for labels to all entities, and then, label each one with the system auth IDs that are allowed to access it. Add system auth ID in an ID token. Add label criteria filters to all operations (CRUD). Based on system auth from the token, append a label filter criteria that finds the system auth ID from the token in the system auth IDs label values of the target entity. All DB calls, in the scope of the request, will have this label filter appended in the where clause.

## Improving Integration Systems (separate document eventually)

The following section proposes an idea for simplifying the set up of integration systems to work with Central Management Plane (CMP). It also provides an idea for automated granting of access to integration systems for applications created by users.

Currently, each integration system should bring its own one time token (OTT) service, which is represented by a pairing adapter in CMP. During deployment of the pairing adapter OTT template mapping for request and response it is specified as env variable. During deployment of Director association between integration system and pairing adapter it is specified as a configuration map. To add the pairing adapter to integration system mapping, the Director must be restarted. Then, the Director can request OTTs from the external OTT service via the pairing adapter. Currently, the integration system always needs a pairing adapter, therefore, the full set up and registration of an integration system in CMP always requires a restart of the Director. It is also possible that a new pairing adapter must be installed, or at least the existing pairing adapter must be restarted (provided that existing pairing adapter is refactored to support multiple external OTT services).

The fact that we always need to restart Director and pairing adapter when introducing a new integration system, and the fact that we can associate only one OTT Service with one integration system, or that we need a pairing adapter deployment per OTT service, are all very limiting.

It must be possible to dynamically add OTT services configuration without restarting anything or having to install new components.

It should not be needed to spin up a new pairing adapter for each external OTT service. Any multitenant, dynamically configurable, pairing adapter must run or it must be removed as a component altogether.

If one integration system provides multiple application types (application templates) it might want to associate different token issuer URLs with each (because the OTT service for each is different, or because the OTT service is the same but serves tokens for the different applications on different URLs). Instead of associating token URL with an integration system, it might be better to include the OTT issuer details in the application template. This would allow the integration system or user, who is creating the template, to also provide OTT issuer metadata there. For applications created without templates, the connector should be used as the default OTT service. If OTT service metadata was not specified in the template, it means that the integration system wants to use the default connector OTT service.

The following is a *desired* flow for setting up and using integration systems with CMP.

A better separation between Director and connector responsibilities is still required.
It is yet unclear is it correct that the flow removes the pairing adapter component. It could be considered a flow with one dynamically configurable multitenant pairing adapter.

Note that the steps from 1 to 9 can be replaced by a simpler alternative when the user directly requests client credentials for the integration system.

1. A user registers an integration system.
2. The user requests OTT for integratiom system pairing.
3. The Director stores system auths record with empty credentials and forwards it to the connector.
4. The connector returns OTT back to the Director, and then, the Director returns it to the user.
5. The user sets OTT in the integration system.
6. The integration system provides OTT with query parameter ?oath or ?cert.
7. The connector verifies the OTT and returns the client credentials or CSR details.
8. In the case of client credentials, the system auth record is updated with these credentials.
9. In the case of CSR, the integration system does one more call to the connector to exchange CSR for the actual cert.
10. Using its new credentials, the integration system registers application templates with the application OTT issuer URL and body/response mapping if it provides external OTT service.
11. The user creates an application from the template.
12. The user requests OTT for the application from the UI.
13. The Director gets the application template for this application, then, resolves the req body template, and then, calls the OTT issuer URL. If an application template does not exist for the application or an external OTT service is not specified in the existing template, it calls connector.
14. The Director processes the response and extracts the token.
15. The Director stores system auth with application ID and OTT as credentials.
16. The LoB administrator puts the token in the application.
17. The application calls the integration system with the token. If it is a connector token, it connects the URL to the integration system as well, and not only to the connector.
18. If needed, the integration system verifies the token with its Token Issuer Service. If it is a connector token, the following step can also serve as a verification.
19. The integration system establishes trust with the LoB application.
20. The integration system calls the Director with OTT and integration system credendtialss for credentials stacking.
21. The Director verifies that it is a known token. Then, it grants system auth access to the integration system credentials for the app, to which this token was issued. As a result, the merged system auths record that was created as part of the token issuing is deleted. Additionally, the Director can return the application details (for example ID, etc) so that the integration system knows what access it has been granted. Alternatively, this can be skipped if the application_id was encoded in the OTT.
22. The integration system can now register packages, APIs and events for the app, and can set a webhook for credentials requests (package instance auth requests).

Some benefits:

* Integration systems can be fully configured to work with CMP without redeployments.
* Stacked credentials enable an automated way to grant access to integration systems, which can access resources. This is an alternative to the manual way, in which a user triggers the `GrantSystemAccess mutation`. Often, there is a user who connects the application (Account Admin) and a user who manages the actual LoB application and integration system (LoB Admin). It is not expected that the Account Admin knows that an integration system will be used for the application pairing; and it is even less likely that the Account Admin knows the `id` or `name` of the integration system when they grant access in it to the application.
* Integration systems can reuse the OTT service that is part of the CMP connector, if they do not have any specific needs for providing a custom OTT issuer.
