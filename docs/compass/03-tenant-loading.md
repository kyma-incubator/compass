# Initial tenant loading

## Overview
This document describes how Compass handles initial tenant loading.

## Job

In Compass the user can define his own tenants. There is a special `compass-director-tenant-loader` job provided for that operation. User has to define a ConfigMap which has a JSON file containing tenants as a data.
There is also `compass-director-default-tenant-loader` job which loads tenants defined in the `global.tenants` field from [compass/values.yaml](../../chart/compass/values.yaml) file with `tenantProvider` value set to `dummy`.

These jobs load the file with tenants from the ConfigMap, parse its content and add it to Compass storage as external tenants with mapping to internal(technical) tenants.

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

Also when loading external tenants, the `global.tenantConfig.tenantProvider` value should be overridden with tenant provider name.
