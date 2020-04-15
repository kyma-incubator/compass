# Tenant loading

## Overview
This document describes how Compass handles tenant loading. 

## Loading default tenants

In Compass, there is a `compass-director-tenant-loader-default` job that allows user to load default tenants specified in `global.tenants` value, which can be found [here](../../chart/compass/values.yaml).

This job is executed once, after the installation of compass chart.
 
It is enabled by default. To disable it, user has to set `global.tenantConfig.useDefaultTenants` value to `false`, [here](../../chart/compass/values.yaml).

Example `global.tenants` value:
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


## Loading external tenants from json

External tenants can be manually loaded at any time by following these steps:
1. Create a ConfigMap `compass-director-external-tenant-config` with embedded JSON file containing tenants to add
2. Create a Job that will load tenants from provided config map by using suspended `compass-director-tenant-loader-external` CronJob as template: 
`kubectl -n compass-system create job --from=cronjob/compass-director-tenant-loader-external compass-director-tenant-loader-external`
3. Wait for the job to finish adding tenants
4. Delete manually created job and config map from cluster

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
