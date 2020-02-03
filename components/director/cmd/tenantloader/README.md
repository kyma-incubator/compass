# Tenant loader

## Overview

The tenant loader is an application that synchronizes tenants from a given directory with Compass.

## How to run 
The user has to provide these environment variables:

| ENV                                      | Default                         | Description                                                   |
| ---------------------------------------- | ------------------------------- | ------------------------------------------------------------- |
| APP_DB_USER                              | postgres                        | Database username                                             |
| APP_DB_PASSWORD                          | pgsql@12345                     | Database password                                             |
| APP_DB_HOST                              | localhost                       | Database host                                                 |
| APP_DB_PORT                              | 5432                            | Database port                                                 |
| APP_DB_NAME                              | postgres                        | Database name                                                 |
| APP_DB_SSL                               | disable                         | Database SSL mode (disable / enable)                          |

## Application flow

1. Read the `/data/` directory where tenant files reside
2. Try to parse each `.json` file from the directory to the tenant structure
3. Synchronize tenants with Compass

## Supported format of .json file with tenants
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
