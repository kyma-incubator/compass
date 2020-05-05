# Default eventing for application


Director, as the main component which manages the information for Applications and Runtimes, is also responsible for providing eventing configurations. The following document describes the process of determining which Runtime should be configured to process the Applications' events.

>**NOTE:** Currently, only one Runtime can receive events from the Application.

An application can obtain the eventing configuration from the Director's API using the `application(id: ID!): Application`. The `Application` type offers the **eventingConfiguration** property that holds the **defaultURL** property with the URL to the eventing backend.
If the value of the **defaultURL** parameter is an empty string, it means that there is no Runtime assigned that can process the Application events.

## Determining default Runtime for the Application events

When querying **eventingConfiguration** for the `Application` type, the Director component tries to determine the default Runtime that will process the events.

In the first step, it searches for a Runtime that belongs to one of the Application's scenarios and has the `{APPLICATION_ID}_defaultEventing = true` label assigned. If there is no such Runtime, the Director looks for the oldest Runtime that is assigned to the Application's scenarios. After finding the oldest Runtime, it is labeled with the `{APPLICATION_ID}_defaultEventing = true` label and next time, it will be treated as the default one. Having the default Runtime determined, the Director component fetches its eventing URL and returns as **defaultURL** for the **eventingConfiguration**.

When the Director is unable to find the default Runtime, the **defaultURL** for the **eventingConfiguration** is an empty string.

## Changing the default Runtime assigned for the Application eventing

The Director API offers a mutation that allows for assigning the default Runtime for the Application eventing. The `setDefaultEventingForApplication(appID: String!, runtimeID: String!): ApplicationEventingConfiguration!` mutation verifies whether the given Runtime belongs to the Application's scenarios and labels it with the `{APPLICATION_ID}_defaultEventing = true` label. If the Application had previously assigned default Runtime, the label from that Runtime is removed.

## Deleting the default Runtime assigned for the Application eventing

The Directors API offers a mutation that allows for deleting the assignment of the default Runtime for the Application eventing. The `deleteDefaultEventingForApplication(appID: String!): ApplicationEventingConfiguration!` mutation deletes the `{APPLICATION_ID}/defaultEventing = true` label from the labeled Runtime.

>**NOTE:** After deleting the `{APPLICATION_ID}_defaultEventing = true` label and querying for the **eventingConfiguration** for the Application, the Director API determines the default Runtime for the Application once again.

## Troubleshooting

### Application eventingConfiguration returns empty defaultURL

It is possible for the Director to return empty `defaultURL` property for the Application eventing configuration in the following circumstances:
- Application does not have a Scenario assigned.
- Application has a Scenario assigned, but the Scenario does not have any Runtimes assigned.
- Default Runtime determined by the assignment is not fully provisioned or connected to Compass.
- Runtime does not have the label for event service URL assigned or the label value is empty.
Whenever the Application eventingConfiguration returns the empty `defaultURL`, check the following:
1. Verify the Application.

The following GraphQL snippet queries the Application by its ID. The response contains Application `id`, `name`, `labels` collection, and `defaultURL` for eventing. Ensure that the Application belongs to at least one Scenario which also has a Runtime assigned.

```graphql
query {
  application(id: "{APPLICATION_ID}") {
    id
    name
    labels
    eventingConfiguration {
      defaultURL
    }
  }
}
```

2. Verify the default Runtime assigned for Application eventing configuration

The following GraphQL snippet queries the Runtimes using the label filter to return the Runtimes with the `{APPLICATION_ID}_defaultEventing` label. There should be only one Runtime returned. The eventing configuration for the Runtime should return the `defaultURL` property with the URL pointing to the Runtime. If the `defaultURL` property is empty, ensure that the Runtime is fully provisioned and connected to Compass.

```graphql
query {
  runtimes(filter: { key: "{APPLICATION_ID}_defaultEventing"}) {
    data {
      id
      name
      eventingConfiguration {
        defaultURL
      }
    }
  }
}
```

3. Verify that the Application Scenarios have Runtimes assigned.

The following GraphQL snippet queries the Runtimes using the label filter for a given Scenario. The response contains the list of Runtimes returning `id`, `name`, and `defaultURL` for eventing.

```graphql
query {
  runtimes(filter: { key: "scenarios", query: "$[*] ? (@ == \"{SCENARIO}\")"}) {
    data {
      id
      name
      eventingConfiguration {
        defaultURL
      }
    }
  }
}
```
