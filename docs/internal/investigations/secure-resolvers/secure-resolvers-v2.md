# Limiting Access to GraphQL Resources 

This document describes a possible solution to the problem of unlimited resource access. The problem that needs to be solved is the current possibility of one application in tenant X to modify the metadata of another application in the same tenant X. This problem is also valid for runtimes. Furthermore, integration systems have global access to all data in Compass. Ideally, an integration system must be able to access only resources that it has created as well as resources to which it has been granted access.

The solution proposal includes limiting the access to operations for `applications`, `application_templates`, `runtimes`, and `integration systems`.

## Terminology

* API Consumer - An API Consumer could be any of the following: application, runtime, integration system, or user.
* Owner (or owning entity) - A top-level entity to which a given resource belongs. For example, an application is the owner of bundles, api definitions, documents, and so on. In Compass, the following owner types are present: `application`, `runtime`, `integration system`, and `application template`.

## Requirements

When a user makes a query or mutation, the secured resolvers concept should not limit the user's access. Put simply, the user consumer type should not be affected by the secured resolvers concept.

By default, an application must not be allowed to modify another application's metadata within the same tenant. This applies not only for the top-level application entity but also for all its related sub-entities such as: bundles, API and event definitions, documents, and so on. For example, application X must not be allowed to insert API definitions in bundle Y when bundle Y belongs to application Z, even if both applications X and Z are in the same tenant.

By default, a runtime must not be allowed to modify another runtime's metadata within the same tenant. Additionally, a runtime must not be allowed to fetch applications that belong to other runtimes.

By default, an integration system must not be allowed to modify another integration system's metadata.

In general, integration systems can manage other systems and their metadata (such as applications, and runtimes). However, an integration system must be allowed to modify only entities that it has created, or entities to which it has been granted access for modifications. For example, if an integration system registers an application, later on, it can also register bundles for this application. However, it cannot modify metadata for applications that other consumers registered, unless it has been explicitly granted access to do so. Another example is when a user registers an application via the UI. To get access to this application, the integration system either must be granted this access by the user via an API call to Director (grant access to the integration system for this application), or some automated procedures must be in place (a concept about stacking credentials via one time tokens is available in the further sections of this document). Although the given examples are related to applications, these requirements are also valid for application sub-entities, application templates, and runtimes managed by integration systems.

Some integration systems (such as UI applications) require unrestricted access to the Director API. They should keep working properly having a global view/access.

An implicit requirement, derived from the requirements above, is that no system must be allowed to list all applications, runtimes, or integration systems.

Out-of-scope: Limit access to the `auths` field of each type to admin users only.

## Current State

The following section analyzes the problem in all current queries and mutations, and groups them in sections providing a conceptual solution.

The problem with securing resolvers is three dimensional. It features all three: mutations/queries; consumer, and owner. In total, there are about 40 mutations/queries that have to be secured, 4 consumer types, and 4 owner types. For specific queries/mutations there is always a single owner but obtaining the ID of that owner is not always straight forward. For example, in `getApplication(applicationID)` the owner ID is passed as input to the query, however, in `updateBundle(bundleID)` some special code must be carried out to find the application ID from the bundle ID.

Why it is important who is the owner and what is the owner ID? Currently, with the help of the ORY integration, the Director API always receives an ID token, regardless of the actual authentication mechanism. Then, the details about the actual caller/consumer can be extracted from the ID token, and then, stored in the Go context. On a high level, it is necessary to ensure that the consumer from the Go context matches the owner of the requested resource. More broadly put (and required especially for integration systems), it is important to verify that the consumer can access the owner's resources. For example, if the owner is an application, the resources that must be accessible are the application details, bundles, and so on.

### Determining the Owner ID

Calculating the owner ID is useful for the cases in which the consumer and owner types match. For example, if a consumer of type application updates an api definition, calculating the owner ID of the api definition and comparing it to the consumer application ID would determine whether this request should be allowed to proceed or not. However, more complicated use cases, where the consumer and the owner types differ, require different approach and solution.

Determining the owner ID during queries and mutations was researched by analyzing each query and mutation. Only the results of this research are provided in this document, and are grouped in sub-sections by the type of the owner:

#### Owner Type: Application

Queries and mutations which owner is an application are grouped in 8 categories, depending on the input provided to the query or mutation. In each category a different piece of code must be carried out to determine the owner ID.

1. Queries and mutations that provide the owner ID as part of the client input (for example: `getApplication`, `updateApplication`, `addBundle`, etc). For these, the owner ID can be determined from the GraphQL arguments list.
2. Queries and mutations that provide a `bundleID` as part of the client input (for example: `updateBundle`, `deleteBundle`, `addDocumentToBundle`, etc). For these, the owner ID can be determined by fetching the actual bundle with ID equal to the `bundleID` from the GraphQL arguments list and getting its `applicationID`.
3. Queries and mutations that provide the `documentID` as part of the client input (for example: `updateDocument`, `deleteDocument`). To get the owner, fetch the document with ID equal to the `documentID` from the GraphQL arguments list, then get its `bundleID`, then fetch the bundle, and finally, get its `applicationID`, which is the owner ID.
4. Queries and mutations that provide `APIDefinitionID` as part of the client input (for example: `updateAPIDefinition`, `deleteAPIDefinition`). To get the owner, fetch the api definition with ID equal to the `APIDefinitionID` from the GraphQL arguments, then get its `bundleID`, then fetch the bundle, and finally, get its `applicationID`, which is the owner ID.
5. Queries and mutations that provide `EventDefinitionID` as part of the client input (for example: `updateEventDefinition`, `deleteEventDefinition`). To get the owner, fetch the event definition with ID equal to the `EventDefinitionID` from the GraphQL arguments, then get its `bundleID`, then fetch the bundle, and finally, get its `applicationID`, which is the owner ID.
6. Queries and mutations that provide `webhookID` as part of the client input (for example: `updateWebhook`, `deleteWebhook`). To get the owner, fetch the webhook with ID equal to the `webhookID` from the GraphQL arguments, and then get its `applicationID`, which is the owner ID.
7. Queries and mutations that provide `systemAuthID` as part of the client input (for example: `deleteSystemAuthForApplication`). To get the owner, fetch the system auth (or a record from a different table that contains both `applicationID` and `systemAuthID`), and then, get the `applicationID`, which is the owner ID.
8. Queries and mutations that provide `bundleInstanceAuthID` as part of the client input (for example: `setBundleInstanceAuth`, `deleteBundleInstanceAuth`). To get the owner, fetch the bundle by `bundleInstanceAuthID` taken from the GraphQL arguments, and then, get the bundle `applicationID`, which is the owner ID.

#### Owner Type: Runtime

1. Queries and mutations that provide the owner ID as part of the client input (for example: `getRuntime`, `updateRuntime`, `deleteRuntime`, etc). For these, the owner ID can be determined from the GraphQL arguments list.
2. Cases in which the owner is of a different type (for example: `applicationsForRuntime`). For these, the `runtimeID` is provided as input and it can be used as ownerID to be compared against the consumerID in case the consumer is of type Runtime. However, additional checks based on scenarios are also performed, implented within the resolver code.

Runtimes consuming APIs owned by different types (for example: `query application`, `requestBundleInstanceAuthCreation`) are outlined in a later section of this document.

#### Owner Type: Integration System

1. Queries and mutations that provide the owner ID as part of the client input (for example: `getIntegrationSystem`, `updateIntegrationSystem`, `deleteIntegrationSystem`, etc). For these, the owner ID can be determined from the GraphQL arguments list.

Integration systems consuming APIs owned by different types are outlined in a later section of this document.

#### Owner Type: Application Template

Application templates, by default, can only be modified by the system which created them and by Users.

This owner type differs from the rest as there is no consumer of this type. This means that obtaining the `owner_id` does not solve scoping of the access on its own. For example, if an integration system or an application creates an application template, it would be necessary to have a relation, such as an `integration_system_id` column or `application_id` column in the `application_templates` table, in order to model the ownership.

#### Summary

It is unclear why the previous implementors decided that the ownerID must be passed in as GraphQL argument in the examples outlined by the section one-s of the previous paragraphs. When the owner type matches the consumer type, and the consumed API response is directly identified by the owner ID (for example, when getApplication(appID) is called by the actual application), it is sufficient for the caller to provide its authentication details, from which the consumer ID is extracted. Forcing the caller to provide the owner ID as a GraphQL argument adds additional complexity to the API logic. That is, the API has to verify that the consumer provided an owner ID that matches its consumer ID. However, it is decided to not change the APIs as part of this design, so such extra check will remain.

If the storage layer is extended with custom queries, it would be possible to calculate the owner ID with no more than 1 DB query in all cases. If not, there are cases when the calculation of ownerID requires 2 or 3 queries. For example, determining the owner id for `updateAPIDefinition` would require to fetch the API definition and also to fetch the bundle. If the consumer is an integration system, it might be needed to fetch the whole application entity as well, which is a third query.

Another approach to determine the owner ID is to store it in each table. That is, each table, which is directly or transitively related to applications must store application_id. However, this adds complexity to the database schema. Additionally, the logic for calculating the owner ID would still be required but it will be implemented during the creation of the resources. Thus, duplicating the owner ID column everywhere does not simplify the overall implementation.

This section analyzed the APIs from the owner perspective. Basically, when the consumer and the owner types match, it is possible to calculate the `ownerID` and compare it to the `consumerID`. If they match, the request is allowed to proceed. The use cases that feature different consumer types are more complex and are outlined in the following sections.

### Analyzing Different Consumer Types

When consumer and owner types match, determination of owner ID is straight forward and is outlined in the previous sections. However, not always the consumer and owner types match, which requires different mechanisms and different approach to the problem. For example, runtimes can consume APIs where the owner type is an application and integration systems can consume APIs where the owner type can vary.

A determined owner ID simplifies solving the problems related to use cases, such as: allowing only application "X" to update its metadata; or allowing only application "X" to update its bundles; or limiting runtime "R" to be the only runtime allowed to fetch its applications; or allowing only integration system "I" to update its own metadata. Anyway, there is more complexity when the consumer is of type integration system. The following section provides analysis of the various use cases of different consumer types (the consumer details are provided in the Go context).

#### Consumer Type: User

No further restrictions apply.

#### Consumer Type: Application

There is no valid use case for an application to consume APIs that belong to other owner types. Therefore, the analysis in the Owners section above is sufficient.

#### Consumer Type: Runtime

Runtimes can be a consumer in a few queries and mutations that belong to owner type application.

* `query application`, `mutation requestBundleInstanceAuthCreation`, `mutation requestBundleInstanceAuthDeletion` - there should be a scenario check, implemented within the resolver code. Currently, such scenario check is missing. Therefore, instead of implementing a scenario check in all these places, it is better to try to fit these in the generic design of secured resolvers.
* `query applicationsForRuntime` - It already performs a scenario check implemented within the code. However, as mentioned already in the document, this query also accepts a `runtime_id` as input parameter and only an "ownerID equals consumerID" check must be added.

#### Consumer Type: Integration System

The `ownerID` check, described above, does not help when the consumer type is integration system and owner type is any of the following: application, runtime, or application template. In this case, it is needed to develop some associations between these entities and the integration system. Instead of calculating the ownerID, it is needed to get the owner entity and get its integration system id, and then, compare it with the `consumerID`.

## Solution Proposal

The purpose of the proposal is to provide a uniform, generic solution that solves all of the above cases. Instead of modeling a design about which consumer (application, integration system, or runtime) can access what owning entity (application, application template, runtime, or integration system) and its sub-entities, it is better to abstract away the consumer dimension and replace it by something that all technical consumers have in common, that is the `system auth`. Therefore, instead of determining what the consumer type is and based on that deciding with what to compare its consumer id with, it is better to check what access has been granted to the system auth related to this consumer. This concept of `system auth accesses` disregards the necessity to check what the consumer type is and allows access checks to be performed against the `system_auth_id` of the consumer, instead of the `consumer_id`. Basically, when a request is received, it can be checked if the consumer system (represented by the `system_auth_id` from the ID token) has been granted access to the owner entity. The owner entity ID needs to be calculated as described previously in this document and then a check if `system_auth_access` exists needs to be performed. `system auth accesses` is stored in a new table. Each row contains information about a `system_auth_id` and the relevant `application_id`, `runtime_id`, `integration_system_id`, or `application_template_id`, to which this `system_auth_id` has access. 

### Add Additional Details in the ID Token

Currently, the ID token created via data, provided by the tenant mapping handler, contains details about `consumer_type` and `consumer_id`. It will be extended to include also details about `consumer_level` and `system_auth_id`. The level of access of the consumer (`RESTRICTED` or `UNRESTRICTED`) is represented by the `consumer_level`, and `system_auth_id` is the ID of the system auth that the consumer is authenticating with.

Tenant Mapping Handler grants `UNRESRICTED` consumer level access to Users and to special or legacy integration systems, such as, UIs that require global access. To model such systems, an additional column (with values `RESTRICTED` and `UNRESTRICTED`) must be added to the `system_auths` table.

### System Auth Access

System auth access (or system auth restrictions) represents the permissions, granted to a specific set of system credentials. For example, if an integration system uses client credentials, represented by system_auth_id=123; a record in `system_auth_acccess` with system_auth_id=123 and app_id=456 means that the consumer (integration system) can access and modify metadata and sub-entities of an application with id=456.

![System Auths Access Table](./assets/system-auth-access.png)

### limitAccess Directive

It is better to do the checks who is allowed to access what centrally, preferably, before the actual resolver business logic of the query or mutation. Since the relevant data about access permissions is stored in one table, it is possible to design a directive that can perform the checks.

```graphql
directive @limitAccess(ownerProvider: String!, idField: String!) on FIELD_DEFINITION
```

* `ownerProvider` - A key that describes how the ownerID of the owner entity can be obtained.
* `idField` - The name of the GraphQL query or mutation argument that is used to calculate the ownerID.
* The directive struct contains a map of provider functions with keys `ownerProvider` values. These functions contain the logic how to calculate the owner ID for the respective query or mutation where the directive is specified.

#### Process Flow

The proposed directive does the following:

1. Loads consumer information from context. Consumer information contains `consumer_level` and `system_auth_id`.
2. Checks if `consumer_level` is `UNRESTRICTED` and if so, allows access.
3. Gets the GraphQL argument represented by the value of `idField`.
4. Looks into the map of owner providers and finds the owner provider function represented by the `ownerProvider` key, specified in the directive.
5. Executes the provider function by passing the argument obtained from the `idField` and gets the owner entity details.
6. After obtaining the ownerProvider ID in the previous step, it checks if a record in `system_auth_access` exists for this owner ID and the `system_auth_id` that is part of the consumer information.
7. If the check from the previous step is true, the request is allowed. Otherwise, if the check is false, the request is denied.

#### Extending the Database Layer

Steps 5 and 6 in the process flow above can be merged if the `system_auth_access` check is implemented within the provider function by joining the `system_auth_access` table to the underlying database query, carried out by the provider function. To do this, some custom database queries will be needed.

#### Examples

```graphql
type Query {
  application(id: ID!): Application @hasScopes(path: "graphql.query.application") @limitAccess(ownerProvider: "GetApplicationID", idField: "id")
}
```

* Value of `idField` is "id" because this is the name of the GraphQL argument.
* In the directive source code `ownerProvider` map key "GetApplicationID" matches to a function that executes a DB query which calculates the `owner_id` (in this case the ID is part of the input) and checks whether the relevant system_auth_access exists.

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
		bundleTable,
		applicationRefField,
		eventDefinitionsTable,
		bundleRefField,
		systemAuthRestrictionsTable,
		applicationRefField,
		systemAuthRefField)

	return r.exists(ctx, stmt, tenantID, id, authID)
}
```

Note how the query simultaneously calculates the owner ID (application id) and checks the necessary `system_auth_access`, and then, simply returns a boolean result whether access has been granted or not.

The actual provider function is pretty simple:

```go
"GetApplicationIDByAPIDefinitionID": appRepo.ExistsApplicationByAPIDefinitionIDAndAuthID,
```

Finally, the following section shows a reference implementation for the directive logic:

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
3. Out-of-scope: It is possible to also model scenario labeling as another input for `system_auth_access` creation. This would allow us to not necessarily implement scenario checks within some specific directives or within the actual resolvers but rather use the system access check that is part of the `limitAccess` directive. For reference, in the implementation, prior to this design, the scenario check is implemented within the `applicationsForRuntime` resolver and the implementation is very suboptimal.

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

New scope is introduced so that it is possible to forbid systems from listing all applications, runtimes, or integration systems. This new scope will be assigned to administrator users only, as follows:

* integration_system:list - for admin user
* runtime:list - for admin user
* application:list - for admin user

A possible improvement would be to grant list access to systems, too, but filter the response based on `system_auth_access` so that a system can list only what it is authorized to access.

### Benefits

1. There are foreign key constraints for everything related to the `system_auth_access` table and no custom exists checks are needed. Also, cascade deletions can be leveraged and no custom implementation in deletion flows in needed. 
2. Logic in grant access API is relatively simple.
3. Amount of data in `system_auth_access` is relatively small. 

### Drawbacks

1. Since the consumer dimension is abstracted away, by using its system_auth/credentials, when a consumer gets a second set of credentials, by default, this second set of credentials will not be authorized to access what the first set can access. To grant the second set of credentials the same access as the first set, the records from `system_auth_access` of the first `system_auth_id` must be copied to the second `system_auth_id`. This can be mitigated by allowing a single set of credentials per system (which is natural and correct). Also, this can be considered as a security improvement in the sense that when somebody manages to request a second set of credentials for an existing system, they will not be able to tamper with the data, which the system can actually access.
2. While a design with a single directive is achieved, its logic still contains a degree of complexity, due to the concept of provider functions and the required provider key input parameter. 

## Variations to the `system_auth_acesss` Table

### Modeling Consumer to Resource Access 

One possible variation is to rename the table to `resource_access` and store information in the following columns: `consumer_id`, `consumer_type`, `resource_id`, and `resource_type`. The first two columns (`consumer_id` and `consumer_type`) specify what is the consumer (application, integration system, runtime) that can access the resource. The second two columns (`resource_id` and `resource_type`) specify which is the actual resource that can be accessed (it can be in any other table in the database).

Since in this approach information is stored about who can access every single resource, it is possible to model access filters (criteria) and add them in the Go context. These access filters can be used to extend the select database query with a join and a where clause. This way it is not needed to do DB calls in the directive. The directive just adds access filters to the Go context.

#### Benefits:

1. The directive logic is simplified. It does not need a map of providers but can just execute a common query for all cases, as follows:
```sql
SELECT 1 
FROM resource_access 
WHERE consumer_id='consumer-id-from-ctx' 
  AND consumer_type='consumer-type-from-ctx' 
  AND resource_id='input-from-graphql-args'
```

2. In addition, the concern in the above approach, about creating a second set of credentials (for a system) that does not have the same access rights as the first set, is solved. This solution does not require additional implementation as `system_auths` are not used.

#### Drawbacks 

1. The table will become very big. It will contain as much rows as the sum of all rows, of all other tables multiplied by the number of consumers.
2. Inserting in the table will happen in each create, register, and add resolver. Additionally, in some resolvers, such as `RegisterApplication`, which create additional related resources as part of the resolver logic (`createRelatedResources` method), there will also be inserts to the `consumer_access` table. Basically, each `Repository.Create` method will be modified to insert in the `consumer_access` table, too.
3. The new API `grant_consumer_access` will be hard to implement. For example, if access is granted to integration system for application and all its subentities, custom logic would be needed in order to verify that the application exists and in order to find all the subentities' IDs so that inserts to `consumer_access` with relevant `resource_id` values can be done. This custom logic will be different for runtimes and integration systems, and as a result, this API has the potential to become too complex.
4. Because the ID columns can represent multiple different entities, there cannot be foreign key constraints to verify that the consumer and the resource actually exist. It is also not possible to allow cascade deletion when the consumer or the resource are deleted. This means that during the deletion of a resource, custom logic is needed in order to find and delete all of its `related` records in the `consumer_access` table. Basically, each `Repo.Delete` method also needs to check and delete relevant related `consumer_access` records. 
5. In this approach, it is an open question how to model UNRESTRICTED access for global integration systems such as UI, KEB, and provisioner. 

### Modeling Consumer to Owner Access

Having access data stored for every single resource seems excessive. In this approach, the stored access information is limited to the owner resources only.
Here, both models from the previous sections meet in the middle and the result is a table `consumer_access` with columns `consumer_id`, `consumer_type`, `owner_id`, and `owner_type`.

#### Benefits

1. This approach still solves the drawback from the first approach, in which a second set of credentials is not be authorized to access what the first set can access without additional implementation.
2. The actual table contains less rows than the previous two approaches. 
3. The grant access API is simpler to implement compared to the one in the second approach. However, it is a bit more complex than the one in the first approach, because it is needed to check if the consumer and owner exist (due to lack of FK constraints).

#### Drawbacks

1. Similarly to the first approach, this approach still has a more complex directive with provider functions. 
2. Similarly to the second approach, usage of foreign keys is not possible. Therefore, custom validation for exists in grant access API and custom delete logic during the deletion of an owner resource is still needed. In this case, it is not needed to modify the deletion of all resources but only for the 4 owner types.
3. It is an open question how to model UNRESTRICTED access for global integration systems such as UI, KEB, and provisioner. Perhaps an empty owner is a solution, however, this is typically not a good practice.

## Additional Research Approaches

This section outlines several additional approaches that were taken into account during the proof of concept (PoC) phase. If they are considered in the future, additional, more thorough proof of concept can be planned and the designs can be discussed in a broader round.

* Usage of helper functions and custom implementation in each resolver.
* `owner_id` and `integration_system_id` are copied to each table. This way, the owner ID does not need to be calculated in each mutation. Field criteria approach can now be used because the `owner_id` is present in each table. The actual business logic query will have a specific where clause based on the field criteria, so that it can decide whether access is granted or not. Basically, a field filter/criteria will be appended by some middleware, and this field criteria will specify values for `owner_id` and `integration_system_id`. This filter/criteria is eventually included as part of the where clause of the DB query.
* Introducing a generic middleware that switches over all mutations and queries, and executes the relevant custom logic.
* Require an owner ID in each mutation on top of the already required input, so that the owner ID does not need to be calculated separately on each mutation.
* Add an ownerReference resolver on each input that returns details about the owning entity. Via a special GraphQL field interceptor/middleware add an additional field to the input so that the resolver is called. In the same or in a different middleware, add logic to compare the owner and the consumer.
* Add support for labels to all entities, and then, label each one with the system auth IDs that are allowed to access it. Add system auth ID in the ID token. Add label criteria filters to all operations (CRUD). Based on the system auth from the token, append a label filter criteria that finds the system auth ID from the token in the system auth IDs label values of the target entity. All DB calls, in the scope of the request, will have this label filter appended in the where clause.

## Improving Integration Systems (separate document eventually)

The following section proposes an idea for simplifying the set up of integration systems to work with Central Management Plane (CMP). It also provides an idea for automated granting of access to integration systems for applications created by users.

Currently, each integration system should bring its own one time token (OTT) service, which is represented by a pairing adapter in CMP. During deployment of the pairing adapter, OTT template mapping for request and response is specified as env variable. During deployment of Director, the association between the integration system and the pairing adapter is specified as a Kubernetes config map. To load the pairing adapter to the integration system mapping, the Director must be restarted. Then, the Director can request OTTs from the external OTT service via the pairing adapter. Currently, each integration system always needs a separate pairing adapter and therefore, the full set up and registration of an integration system in CMP always requires a restart of the Director and also a new pairing adapter installation for the integration system, or at least the existing pairing adapter must be restarted and reconfigured (provided that the existing pairing adapter implementation is refactored to support multiple external OTT services).

The fact that we always need to restart Director and pairing adapter when introducing a new integration system, and the fact that we can associate only one OTT Service with one integration system, or that we need a pairing adapter deployment per OTT service, are all very limiting.

It must be possible to dynamically add OTT services configuration without restarting anything or having to install new components.

It should not be needed to spin up a new pairing adapter for each external OTT service. Either a multitenant, dynamically configurable, pairing adapter must be implemented, or the pairing adapter must be removed as a component altogether.

If one integration system provides multiple application types (application templates) it might want to associate different OTT issuer URLs with each (because the OTT service for each is different, or because the OTT service is the same but serves tokens for the different applications on different URLs). Instead of associating OTT URL with an integration system, it might be better to include the OTT issuer details in the application template. This would allow the integration system or user, that is creating the template, to also provide OTT issuer metadata there. For applications created without application templates, the connector should be used as the default OTT service. If OTT service metadata was not specified in the application template, it means that the integration system wants to use the default connector as an OTT service.

The following is a *desired* flow for setting up and using integration systems with CMP.

A better separation between Director and connector responsibilities is still required before the flow is implemented.
It is yet unclear if removing the pairing adapter is the best choice. It might make sense to detail out and consider another flow with one dynamically configurable multitenant pairing adapter. 

**Note**: Steps from 1 to 9 can be replaced by a simpler alternative: The user directly requests client credentials for the integration system and puts them in the integration system's environment.

1. A user registers an integration system.
2. The user requests OTT for integration system pairing.
3. The Director stores system auths record with empty credentials and forwards it to the Connector.
4. The Connector returns OTT back to the Director, and then, the Director returns it to the user.
5. The user sets OTT in the integration system.
6. The integration system provides OTT with query parameter ?oath or ?cert.
7. The Connector verifies the OTT and returns the client credentials or CSR details.
8. In the case of client credentials, the system auth record is updated with these credentials.
9. In the case of CSR, the integration system does one more call to the Connector to exchange CSR for the actual cert.
10. Using its new credentials, the integration system registers application templates with the application OTT issuer URL and body/response mapping if it provides external OTT service.
11. The user creates an application from the template.
12. The user requests OTT for the application from the UI.
13. The Director gets the application template for this application, then, resolves the req body template, and then, calls the OTT issuer URL. If an application template does not exist for the application or an external OTT service is not specified in the existing template, it calls the Connector.
14. The Director processes the response and extracts the token.
15. The Director stores system auth with application ID and OTT as credentials.
16. The LoB administrator puts the token in the application.
17. The application calls the integration system with the token. If it is a Connector token, it should contain a URL pointing to the integration system as well, and not only to the connector.
18. If needed, the integration system verifies the token with its Token Issuer Service. If it is a Connector token, the next step can also serve as a verification and nothing needs to be done in this step.
19. The integration system establishes trust with the LoB application.
20. The integration system calls the Director with token and integration system credentials for credentials stacking.
21. The Director verifies that it is a known token. Then, it grants system auth access to the integration system credentials for the application for which this token was issued. As a result, the system auths record that was created as part of the token issuing is now merged to the integration system credentials and can be deleted. Additionally, the Director may also return the application details (for example ID, etc) so that the integration system would know what it has been granted access for. Alternatively, this can be skipped if the application_id was encoded in the OTT.
22. The integration system can now register bundles, APIs and events for the application, and can set a webhook for credentials requests (bundle instance auth requests).

Some benefits of the proposed approach:

* Integration systems can be fully configured to work with CMP without redeployments and without the need to install a new pairing adapter for each integration system. 
* Stacking credentials enables an automated way to grant integration systems access to resources. This is an alternative to the manual way, in which a user triggers the `GrantSystemAccess mutation`. Often, there is a user who connects the application (Account Admin) and a user who manages the actual LoB application and integration system (LoB Admin). It is not expected that the Account Admin knows that an integration system will be used for the application pairing; and it is even less likely that the Account Admin knows the `id` or `name` of the integration system in order to grant it access to the application.
* Integration systems can reuse the OTT service that is part of the CMP connector, if they do not have any specific needs for providing a custom OTT Service.
