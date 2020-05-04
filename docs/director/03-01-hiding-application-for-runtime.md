# Hiding Application for Runtime

It is possible to configure Compass to not return Applications labeled with specific labels, in response to query `applicationsForRuntime` used by Runtimes. 

## Configuration

To configure the list of selectors which will be used to filter the list of returned Applications modify **applicationHideSelectors** parameter in the [`values.yaml`](https://github.com/kyma-incubator/compass/blob/master/chart/compass/charts/director/values.yaml) file. You can modify this list by providing appropriate overrides.

**applicationHideSelectors** should be a map of lists of **strings**, where map keys represent label keys and strings represent label values that will be used to filter the returned Applications.

Each label key can have multiple values assigned, if an Application is labeled with that key and any of the specified values, it will not be returned in `applicationsForRuntime` query. Multiple label keys can be configured.

## Example
```yaml
applicationHideSelectors:
  applicationType:
    - "Test Application"
    - "WIP"
  applicationVersion:
    - "alpha"  
```
