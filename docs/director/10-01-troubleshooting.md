# Troubleshooting

## Application eventing configuration returns empty defaultURL

It is possible for the Director to return empty `defaultURL` property for the Application eventing configuration in the following circumstances:
- Application does not have a Scenario assigned.
- Application has a Scenario assigned, but the Scenario does not have any Runtimes assigned.
- Default Runtime determined by the assignment is not fully provisioned or connected to Compass.
- Runtime does not have the label for event service URL assigned or the label value is empty.

Whenever the Application eventing configuration returns the empty `defaultURL`, check the following:

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

2. Verify the default Runtime assigned for Application eventing configuration.

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
