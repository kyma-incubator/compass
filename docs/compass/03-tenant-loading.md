# Initial tenant loading

## Overview
This document describes how Compass handles initial tenant loading.

## Job

In Compass the user can define his own tenants. There is a special `compass-director-external-tenant-loader` job provided for that operation. To use that job, the user has to follow these steps:
 - change the `useExternalTenants` value to `true`, 
 - define a ConfigMap which has a JSON file containing tenants as a data (example below),
 
or 
 - reuse [existing ConfigMap](../../chart/compass/charts/director/templates/configmap-external-tenant-config.yaml) by setting `useExternalTenants` and `useExternalTenantsConfigMapFromChart` values to true. The user should change the data to what is appropriate for his use-case 

There is also `compass-director-default-tenant-loader` job which loads tenants defined in the `global.tenants` field with `tenantProvider` value set to `dummy`.

All mentioned values can be found from [here](../../chart/compass/values.yaml).

These jobs load the file with tenants from the ConfigMap, parse its content and add it to Compass storage as external tenants with mapping to internal(technical) tenants.

Example ConfigMap:
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

Also when loading external tenants, the `global.tenantConfig.tenantProvider` value should be overridden with tenant provider name.
