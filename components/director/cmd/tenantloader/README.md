# Tenant loader

## Overview

It is an application that is supposed to synchronize given tenants from given file with Compass.

## How to run
The user has to provide two environment variables:

- `TENANTS_SRC` - which is a path to JSON file which holds tenant data
- `TENANT_PROVIDER` - which specifies the provider of tenants

## Application flow

1. Read the file from source provided in `TENANTS_SRC` environment variable.
2. Try to parse JSON file to tenant structure
3. Synchronise tenants with Compass (not implemented yet)

## Supported format of JSON file with tenants
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
