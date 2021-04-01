# Open Resource Discovery Aggregator

## Overview

The Aggregator application collects, aggregates and stores the ORD information from multiple ORD providers to the Compass' database.

## Prerequisites

The Aggregator requires access to:
1. Configured PostgreSQL database with the imported Director's database schema.
2. Running Director component
3. API that can be called to fetch and process ORD Documents.

## Configuration

To run the application, provide these environment variables:

| Environment variable       | Default value                                                 | Description                                |
| -------------------------- | ------------------------------------------------------------- | ------------------------------------------ |
| **APP_DB_USER**            | `postgres`                                                    | Database username                          |
| **APP_DB_PASSWORD**        | `pgsql@12345`                                                 | Database password                          |
| **APP_DB_HOST**            | `localhost`                                                   | Database host                              |
| **APP_DB_PORT**            | `5432`                                                        | Database port                              |
| **APP_DB_NAME**            | `postgres`                                                    | Database name                              |
| **APP_DB_SSL**             | `disable`                                                     | Parameter that activates database SSL mode |
| **APP_CONFIGURATION_FILE** | Absolute path to `components/director/hack/config-local.yaml` | The path to the configuration file         |

## Details

The Aggregator basic workflow is as follows:

1. The Aggregator loops through all available Applications stored in the Compass' database.
2. For each Application that has a Webhook of type `OPEN_RESOURCE_DISCOVERY` it calls the URL that is attached to that Webhook.
3. That URL has predefined endpoints which serve the necessary information for the Aggregator.
4. The Aggregator aggregates and stores the served information in the Compass' database.
