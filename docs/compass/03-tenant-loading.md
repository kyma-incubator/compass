# Initial tenant loading

## Overview
This document describes how Compass handles initial tenant loading.

## Job

In Compass the user can define his own tenants. There is a special `compass-director-tenant-loader` job provided for that operation. User has to define a ConfigMap which has a JSON file containing tenants as a data. 

The ConfigMap name has to be `compass-director-tenant-config` and the file inside has to be named `tenants.json`. Content of JSON file has to be in the following format:
```json
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

Additionally the user has to set a `global.tenantConfig.externalTenants` value to true. 

The job will load the file from the ConfigMap, parse its content and add it to Compass storage as external tenants with mapping to internal(technical) tenants.

Also when loading custom tenants, the `global.tenantConfig.tenantProvider` value should be overridden with tenant provider name.

If a user doesn't want to provide custom tenants, by default the job will load tenants defined in `chart/compass/values.yaml` with `tenantProvider` value set to `compass`.
