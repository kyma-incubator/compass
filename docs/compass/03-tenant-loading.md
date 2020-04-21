# Tenant loading

There are two ways in which Compass loads a list of tenants:
- Initial loading of default tenants that occurs after the chart installation
- Manual loading of tenants provided in a ConfigMap

## Initial tenant loading

There are three default tenants predefined for Compass. These tenant are:
- `default`
- `foo`
- `bar`

The list of default tenants is specified in the [**global.tenants**](https://github.com/kyma-incubator/compass/blob/7bc8dae6f37d3477ecd986af2867d2d8a6ab0a56/chart/compass/values.yaml#L1) parameter. You can modify this list by providing appropriate overrides.
 
The `compass-director-tenant-loader-default` job that loads the list of default tenants is executed only once after the installation of the Compass chart. It is enabled by default. To disable it, set **global.tenantConfig.useDefaultTenants** to `false`. 

>**CAUTION:** If you disable the job that loads the list of default tenants, make sure to provide your tenants manually. Otherwise, the list of tenants will be empty, which will make the whole system inoperative.


## Manual tenant loading

You can load external tenants manually at any time by following these steps:

1. Create the `compass-director-external-tenant-config` ConfigMap with an embedded JSON file that contains tenants to add. See the example:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: compass-director-external-tenant-config
  namespace: compass-system
data:
  tenants.json: |-
    [
      {
        "name": "tenant-name-1",
        "id": "tenant-id-1"
      },
      {
        "name": "tenant-name-2",
        "id": "tenant-id-2"
      }
    ]
```

2. Create a job that loads tenants from the provided ConfigMap by using the suspended `compass-director-tenant-loader-external` CronJob as a template: 
    ```sh
    kubectl -n compass-system create job --from=cronjob/compass-director-tenant-loader-external compass-director-tenant-loader-external
    ```
3. Wait for the job to finish adding tenants.
4. Delete the manually created job and ConfigMap from the cluster.

