# GraphQL Partial Updates

## Overview

This documents describes and compares different approaches to designing partial updates in graphql schema. Most of presented approaches have a link to an example they were based on. 

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

#### Description
In this approach every field has to be optional. If the field is defined, we use the passed value to set the new value. If the field is not defined we don't change the current value.

Because we use the `nil` value to determine whether to change current value or not we lose the ability to set the actual value to `nil`. In the example linked below this limitation doesn't exist because in javascript we can distinguish `null` and `undefined` values.

Example: [link](https://medium.com/workflowgen/graphql-mutations-partial-updates-implementation-bff586bda989)

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

#### Description
This approach is very similar to the previous one, with one difference: there has to be only one, unique input object. In the example linked below it is argued that it makes the client implementation easier.

Patching logic is the same as in previous approach.

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

#### Description

In this approach we define commands (`updateApplication` in provided schema) and actions (`ApplicationUpdateActions` in provided schema). Each action requires defining additional input type for it.

This way we are getting rid of restriction on `nil` values from previous examples, because if we don't want to update the current value we just don't use the action, and if we do specify it, the value of field nested inside can be a `nil`.

The drawback of this solution is a lot of boilerplate needed for each mutation.

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

#### Description

This solution works like PUT operation, we always change the value but limit the fields in input object. We need to first fetch the original on client side and modify the fields we want to update.

To delete all subresources of application (e.g. webhooks) we have separate mutations, but we lose the ability to replace them with different ones.

## Conclusion

Personally I'm leaning towards the first presented solution. It's simple yet functional and in our case the drawback of inability of setting the value to `nil` doesn't seem concerning.

I'm not convinced by the "single input object" approach, having at least the `ID` of affected application as a mutation argument seems to me like a clearer option.

The "Commands & Actions" approach seems like a good idea when ability to set the value to `nil` is a requirement, I don't think in our case it's worth adding the additional boilerplate.

The "PUT Approach" requires additional work on client side while limiting the functionality of API, so I don't think that's a good solution.
