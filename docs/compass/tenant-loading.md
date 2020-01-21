# Initial tenant loading

## Overview
This document describes how Compass handles initial tenant loading.

## Job

In Compass the user can define his own tenants. There is a special `compass-tenant-loader` job provided for that operation. User has to define a JSON file which looks like that
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
The JSON file has to be added to `compass-tenant-config` ConfigMap defined in `chart/compass/charts/director/templates/configmap-tenant-config.yaml` where you have to put your file into the data object.

Example ConfigMap:
```helmyaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: compass-tenant-config
  namespace: {{ .Release.Namespace }}
  labels:
    app: {{ .Chart.Name }}
    release: {{ .Release.Name }}
data:
    {{- tpl ((.Files.Glob "tenants.json").AsConfig) . | nindent 2 }}
```

**The file has to be named `tenants.json`.**

The job will load the file from the ConfigMap, parse its content and add it to Compass storage as external tenants with mapping to internal(technical) tenants.

Also when loading custom tenants, the `global.tenantProvider` value should be overridden.
 
If a user doesn't want to provide custom tenants, by default the job will load tenants defined in `chart/compass/values.yaml` with `tenantProvider` value set to `compass`.

## Custom tenants JSON file
There might be a use-case, when user wants to provide tenants JSON file with another format, e.g.
```json
[
  {
    "my-tenant-name": "tenant-name-1",
    "my-tenant-id": "tenant-id-1"
  },
  {
    "my-tenant-name": "tenant-name-2",
    "my-tenant-id": "tenant-id-2"
  }
]
```

There is a mechanism which allows mapping these custom JSON keys to fulfilling the contract of the tenant loader Job.
To do so, the `name` and `id` of `global.tenantKeyMapping` value has to be overridden with actual JSON key names.

To match the example above, the values would look like that:
```yaml
global:
  tenantKeyMapping:
    name: "my-tenant-name"
    ID: "my-tenant-id"
```
