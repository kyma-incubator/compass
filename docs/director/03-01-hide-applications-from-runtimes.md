# Hide Applications from Runtimes

It is possible to configure Compass so that it doesn't return Applications labeled with specific labels in response to the `applicationsForRuntime` query used by Runtimes. 

## Configuration

To configure the list of selectors used to filter the list of returned Applications, modify the **applicationHideSelectors** parameter in the [`values.yaml`](https://github.com/kyma-incubator/compass/blob/main/chart/compass/charts/director/values.yaml) file. You can modify this list by providing appropriate overrides.

The **applicationHideSelectors** parameter is a multiline string with a map that contains the lists of **strings**. Map keys represent label keys, and strings represent label values that are used to filter the returned Applications.

You can configure multiple label keys. Each label key can have multiple values assigned. If an Application is labeled with that key and any of the specified values, it is not returned in the `applicationsForRuntime` query.

See the example of labels specified in the **applicationHideSelectors** parameter:
```yaml
applicationHideSelectors: |-
  applicationType:
    - "Test Application"
    - "WIP"
  applicationVersion:
    - "alpha"
```

In this example, all Applications with the `applicationType: "Test Application"`, `applicationType: "WIP"`, or `applicationVersion: "alpha"` label are not visible for Runtimes.
