# Hide Applications from Runtimes

It is possible to configure Compass so that it doesn't return Applications labeled with specific labels in response to the `applicationsForRuntime` query used by Runtimes. 

## Configuration

To configure the list of selectors used to filter the list of returned Applications, modify the **applicationHideSelectors** parameter in the [`values.yaml`](https://github.com/kyma-incubator/compass/blob/master/chart/compass/charts/director/values.yaml) file. You can modify this list by providing appropriate overrides.

The **applicationHideSelectors** parameter is a map that contains the lists of **strings**, where map keys represent label keys, and strings represent label values that are used to filter the returned Applications.

Each label key can have multiple values assigned, if an Application is labeled with that key and any of the specified values, it will not be returned in `applicationsForRuntime` query. Multiple label keys can be configured.

See the example of the labels specified in the **applicationHideSelectors** parameter:
```yaml
applicationHideSelectors:
  applicationType:
    - "Test Application"
    - "WIP"
  applicationVersion:
    - "alpha"  
```

In this example, all Applications with the `applicationType: "Test Application"`, `applicationType: "WIP"`, or `applicationVersion: "alpha"` label are not visible for Runtimes.
