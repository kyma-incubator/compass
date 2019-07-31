# GraphQL Partial Updates

## Overview

This document describes and compares different approaches to designing partial updates in graphql schema. Most of the presented approaches have a link to an example they were based on. 

## Partial update mutation designs

### 1. Simple Patch

#### Schema
```graphql
input ApplicationPatchInput {
    name: String
    description: String
    webhooks: [WebhookInput!]
    healthCheckURL: String
    apis: [APIDefinitionInput!]
    eventAPIs: [EventAPIDefinitionInput!]
    documents: [DocumentInput!]
}

type Mutation {
    updateApplication(id: ID!, patch: ApplicationPatchInput!): Application!
}
```

#### Usage

**Set description to empty value**
```graphql
updateApplication(id: "52cc65fe-c94f-4d94-b59a-01c1ba865547", patch: {
    description: ""
}) {
    id
    description
}
```

**Remove all EventAPIs for Application**
```graphql
updateApplication(id: "52cc65fe-c94f-4d94-b59a-01c1ba865547", patch: {
    eventAPIs: []
}) {
    id
    eventAPIs
}
```

#### Description
In this approach, every field has to be optional. If the field is defined, we use the passed value to set the new value. If the field is not defined we don't change the current value.

Because we use the `nil` value to determine whether to change current value or not we lose the ability to set the actual value to `nil`. In the example linked below this limitation doesn't exist because in javascript we can distinguish `null` and `undefined` values.

Example: [link](https://medium.com/workflowgen/graphql-mutations-partial-updates-implementation-bff586bda989)

##### Handling nil values if we would need them (@kfurgol)
If such a case occurred, we came up with several ideas:
First, we thought about providing a wrapper type which would look like that:
```graphql
input nilableInteger {
	value: Int
	isNil: Boolean
}
```
and the size would be of type `nilableInteger`.
In the case when we don't know the value, we would set `isNil` value to true. Otherwise, we would set `isNil` to false and `value` to 0.

Another idea would be that we stick to strings in our api regardless of what value would suggest. So the `size` value would be a string.

Also, we investigated more the `Commands & Actions` approach, where such a problem wouldn't occur, but we would have to specify setter inputs for pretty much all of inputs.

There is also a case to consider with default values. For example, for now, we have `VersionInput` type which has two boolean values with default values. Considering a case when we want to update just value of the version type, and actual booleans differ from the default values, they would've been overridden. One of our propositions to handle that case is to create another input type for the `Version`.

We have to remember that if we want to use `PATCH` updates, we'll have to specify additional inputs for each existing `PUT` input.

### 2. Simple Patch (single input)

#### Schema
```graphql
input ApplicationUpdateInput {
    id: ID!,
    patch: ApplicationPatchInput!
}

input ApplicationPatchInput {
    name: String
    description: String
    webhooks: [WebhookInput!]
    healthCheckURL: String
    apis: [APIDefinitionInput!]
    eventAPIs: [EventAPIDefinitionInput!]
    documents: [DocumentInput!]
}

type Mutation {
    updateApplication(in: ApplicationUpdateInput!): Application!
}
```

#### Usage

**Set description to empty value**
```graphql
updateApplication(in: {
    id: "52cc65fe-c94f-4d94-b59a-01c1ba865547"
    patch: {
        description: ""        
    }
}) {
    id
    description
}
```

**Remove all EventAPIs for Application**
```graphql
updateApplication(in: {
    id: "52cc65fe-c94f-4d94-b59a-01c1ba865547"
    patch: {
        eventAPIs: []        
    }
}) {
    id
    eventAPIs
}
```

#### Description
This approach is very similar to the previous one, with one difference: there has to be only one, unique input object. In the example linked below, it is argued that it makes the client implementation easier.

Patching logic is the same as in the previous approach.

Example: [link](https://blog.apollographql.com/designing-graphql-mutations-e09de826ed97)

### 3. Commands & Actions

#### Schema
```graphql
input ApplicationUpdateActions {
    setName: SetApplicationName
    setDescription: SetApplicationDescription
    setWebhooks: SetApplicationWebhooks
    setHealthCheckURL: SetApplicationHealthCheckURL
    setAPIs: SetApplicationAPIs
    setEventAPIs: SetApplicationEventAPIs
    setDocuments: SetApplicationDocuments
}

input SetApplicationName {
    name: String
}

input SetApplicationDescription {
    description: String
}

input SetApplicationWebhooks {
    webhooks: [WebhookInput!]
}

input SetApplicationHealthCheckURL {
    healthCheckURL: String
}

input SetApplicationAPIs {
    apis: [APIDefinitionInput!]
}

input SetApplicationEventAPIs {
    eventAPIs: [EventAPIDefinitionInput!]
}

input SetApplicationDocuments {
    documents: [DocumentInput!]
}

type Mutation {
    updateApplication(id: ID!, actions: ApplicationUpdateActions!): Application!
}
```

#### Usage

**Set description to empty value**
```graphql
updateApplication(id: "52cc65fe-c94f-4d94-b59a-01c1ba865547", actions: {
    setDescription: {
        description: nil # Can be either nil or empty string, depending on implementation
    }
}) {
    id
    description
}
```

**Remove all EventAPIs for Application**
```graphql
updateApplication(id: "52cc65fe-c94f-4d94-b59a-01c1ba865547", actions: {
    setEventAPIs: {
        eventAPIs: nil # Can be either nil or empty array, depending on implementation 
    }
}) {
    id
    eventAPIs
}
```

#### Description

In this approach, we define commands (`updateApplication` in the provided schema) and actions (`ApplicationUpdateActions` in the provided schema). Each action requires defining additional input type for it.

This way we are getting rid of restriction on `nil` values from previous examples, because if we don't want to update the current value we just don't use the action, and if we do specify it, the value of field nested inside can be a `nil`.

The drawback of this solution is a lot of boilerplate needed for each mutation.

Alternatively, we could limit the required action input types introduced to one per field type. For example instead of `setName` and `setHealthCheckURL` we could use just `setString`. That would still require us to have separate actions for required and optional fields (e.g. `setOptionalString` and `setRequiredString`).
That way we could significantly reduce the needed boilerplate.

Example: [link](https://techblog.commercetools.com/modeling-graphql-mutations-52d4369f73b1)

### 4. PUT Approach with additional mutations

#### Schema
```graphql
input ApplicationUpdateInput {
    name: String
    description: String
    healthCheckURL: String
}

type Mutation {
    updateApplication(id: ID!, in: ApplicationUpdateInput!): Application!
    deleteAllApplicationWebhooks(id: ID!): Application!
    deleteAllApplicationAPIs(id: ID!): Application!
    deleteAllApplicationEventAPIs(id: ID!): Application!
    deleteAllApplicationDocuments(id: ID!): Application!
}
```

#### Usage

**Set description to empty value**
```graphql
updateApplication(id: "52cc65fe-c94f-4d94-b59a-01c1ba865547", in: {
    name: "Application"
    description: nil # Can be either nil or empty string, depending on implementation
    healthCheckURL: "https://health.check/"
}) {
    id
    description
}
```

**Remove all EventAPIs for Application**
```graphql
deleteAllApplicationEventAPIs(id: "52cc65fe-c94f-4d94-b59a-01c1ba865547") {
    id
    eventAPIs
}
```

#### Description

This solution works like PUT operation, we always change the value but limit the fields in the input object. We need to first fetch the original on the client side and modify the fields we want to update.

To delete all subresources of application (e.g. webhooks) we have separate mutations, but we lose the ability to replace them with different ones.

## Conclusion

After taking everything into account I think the best solution would be to create additional "PATCH" update mutations for cases that require them and leave current "PUT" updates. In such case I'd use "Commands and Actions" approach with either unique setters or type based ones.

If we were to replace all our current "PUT" updates with "PATCH" updates, then I'd be leaning towards the first presented solution. It's simple yet functional and in our case, the drawback of the inability of setting the value to `nil` doesn't seem concerning.

I'm not convinced by the "single input object" approach, having at least the `ID` of affected application as a mutation argument seems to me like a clearer option.

The "Commands & Actions" approach seems like a good idea when the ability to set the value to `nil` is a requirement, I don't think in our case it's worth adding the additional boilerplate.

The "PUT Approach" requires additional work on the client side while limiting the functionality of API, so I don't think that's a good solution.

## Appendix
We had to make some additional choices related to resource versioning and performing updates on the SQL database.

### Resource versioning

#### Overview

Each resource that can be updated (Runtime, Application, API, Document, etc.) should have additional field `version` (name TBD) that would be updated each time its update mutation is executed, and the client would need that version each time it presents user some data that can be updated.

If the client sends an update request with the version that doesn't match the version currently stored on the server, that update will be rejected (because someone else probably modified the resource since the user received his data).

#### Versioning and SQL db tables

Let's imagine a situation when a user sends `updateApplication` request that would set the Application's Document array to a different one. On the database level that wouldn't result in any changes to `applications` table, only to `documents`.
We still need to update the Application's `version` field to avoid the situation when, for example, one user sets the new value to the Document array and other user sets that value to an empty array. 

### SQL Update

We considered two ways to perform updates on the SQL database.

In the first approach, we would start with fetching the current version of the resource, then we would patch the fields that were sent in the request, increase the `version` field and save the updated resource in the database.

The second approach would require implementing a SQL query builder that would dynamically build update queries modifying only the specified fields and increasing resource `version` field.

After discussions in the team, we decided to implement the first approach.