# Initial tenant loading

## Overview
This document describes how Compass handles initial tenant loading.

## Job

In Compass the user can define his own tenants. There is a special `compass-director-tenant-loader` job provided for that operation. To use that job, the user has to follow these steps:
 - change the `useExternalTenants` value to `true`, 
 - create a ConfigMap which has a JSON file containing tenants as a data (example below),
 
or 
 - reuse [existing ConfigMap](../../chart/compass/charts/director/templates/configmap-external-tenant-config.yaml) by setting `useExternalTenants` and `useExternalTenantsConfigMapFromChart` values to true. The user should change the data to what is appropriate for his use-case. 

All mentioned values can be found from [here](../../chart/compass/values.yaml).

Furthermore, the job loads default tenants also defined in `values.yaml` file pointed above. To prevent loading default tenants the user has to override `global.tenantConfig.useDefaultTenants` with `false` value.   

That job mount file with tenants from the ConfigMap, parse its content and add it to Compass storage as external tenants with mapping to internal(technical) tenants.

Example ConfigMap:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: compass-director-tenant-config
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
