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

It is possible for the Director to return empty `defaultURL` property for the Application eventing configuration in certain circumstances:
- the Application does not have a scenario assigned
- the Application have scenario assigned, but the scenario does not have any Runtimes assigned
- the default Runtime determined by the assignment is not fully provisioned or connected to the Compass
- the Runtime does not have the label for event service URL assigned or the label value is empty

1. Verify the Application.

The following GraphQL snippet queries the application by its ID. The response contains application `id`, `name`, `labels` collection, and `defaultURL` for eventing. Please ensure that the Application belongs to at least one scenario which also has runtime assigned.

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

The following GraphQL snippet queries the runtimes using the label filter to return the runtimes with the label `{APPLICATION_ID}_defaultEventing`. There should be only one runtime returned. The eventing configuration for the runtime should return `defaultURL` property with the URL pointing to the runtime. If the `defaultURL` property is empty, ensure that the runtime is fully provisioned and connected to the Compass.

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

3. Verify that the Application scenarios have Runtimes assigned.

The following GraphQL snippet queries the runtimes using the label filter for a given scenario. The response contains the list of runtimes returning `id`, `name`, and `defaultURL` for eventing.

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