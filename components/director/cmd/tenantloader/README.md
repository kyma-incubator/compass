# Tenant Loader

## Overview

Tenant Loader is an application that loads tenants from a given directory to Compass.

## Usage

To run the application, provide these environment variables:

| Environment variable                                      | Default value                         | Description                                                   |
| ---------------------------------------- | ------------------------------- | ------------------------------------------------------------- |
| **APP_DB_USER**                              | `postgres`                        | Database username                                             |
| **APP_DB_PASSWORD**                          | `pgsql@12345`                     | Database password                                             |
| **APP_DB_HOST**                              | `localhost`                       | Database host                                                 |
| **APP_DB_PORT**                              | `5432`                            | Database port                                                 |
| **APP_DB_NAME**                              | `postgres`                        | Database name                                                 |
| **APP_DB_SSL**                               | `disable`                         | Database SSL mode (disable / enable)                          |

## Details

Tenant Loader basic workflow looks as follows:

1. Tenant Loader reads the `/data/` directory where tenant files reside.
2. Tenant Loader tries to parse each `.json` file from the directory to the tenant structure.
3. Tenant Loader creates tenants with IDs that are not yet in database.


This is the supported format of the `.json` file that can be parsed to the tenant structure:
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
