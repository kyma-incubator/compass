# Initial tenant loading

## Overview
This document describes how Compass handles initial tenant loading.

## Job

In Compass, the user can define their own tenants. There is a special `compass-director-tenant-loader` job provided for that operation. To use that job, the user has to follow these steps:
  - change the `useExternalTenants` value, which can be found [here](../../chart/compass/values.yaml), to `true`.
 - create a ConfigMap which has a JSON file containing tenants as a data.

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

Furthermore, the job loads default tenants also defined in the `values.yaml` file. To prevent loading default tenants, the user has to override `global.tenantConfig.useDefaultTenants` with a `false` value.

To make the job mount a file with tenants from the ConfigMap, parse its content and add it to Compass storage as external tenants with mapping to internal(technical) tenants.