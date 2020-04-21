# Tenant loading

There are two ways in which Compass loads a list of tenants:
- Initial loading of default tenants that occurs after the chart installation
- Manual loading of tenants provided in a ConfigMap


## Load default tenants

In Compass, there is a `compass-director-tenant-loader-default` job that allows you to load default tenants specified in the [**global.tenants** parameter](../../chart/compass/values.yaml).

This job is executed once, after the installation of the Compass chart.
 
The job is enabled by default. To disable it, set **global.tenantConfig.useDefaultTenants** to `false`.

See the example of the specified **global.tenants** parameter:
```yaml
global:
  tenants:
    - name: default
      id: 3e64ebae-38b5-46a0-b1ed-9ccee153a0ae
    - name: foo
      id: 1eba80dd-8ff6-54ee-be4d-77944d17b10b
    - name: bar
      id: af9f84a9-1d3a-4d9f-ae0c-94f883b33b6e
``` 


## Load external tenants manually

You can load external tenants manually at any time by following these steps:
1. Create the `compass-director-external-tenant-config` ConfigMap with an embedded JSON file that contains tenants to add.
2. Create a job that loads tenants from the provided ConfigMap by using the suspended `compass-director-tenant-loader-external` CronJob as a template: 
    ```sh
    kubectl -n compass-system create job --from=cronjob/compass-director-tenant-loader-external compass-director-tenant-loader-external
    ```
3. Wait for the job to finish adding tenants.
4. Delete the manually created job and ConfigMap from the cluster.

See the example of a ConfigMap:
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
