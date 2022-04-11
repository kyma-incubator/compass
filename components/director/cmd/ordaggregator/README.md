# Open Resource Discovery Aggregator

## Overview

The Aggregator application collects, aggregates, and stores the ORD information from multiple ORD providers to the Compass's database.


## Details

The Aggregator basic workflow is as follows:

1. The Aggregator goes through all available applications, stored in the Compass's database.
2. For each application that has a webhook of type `OPEN_RESOURCE_DISCOVERY` it calls the URL that is attached to that webhook.
3. That URL has predefined endpoints, which provide the necessary information to the Aggregator.
4. The Aggregator aggregates and stores the provided information in the Compass's database.

## Configuration

The ORD Aggregator binary allows you to override some configuration parameters. To get a list of the configurable parameters, see [main.go](https://github.com/kyma-incubator/compass/blob/75aff5226d4a105f4f04608416c8fa9a722d3534/components/director/cmd/ordaggregator/main.go#L49).

## Local Development
### Prerequisites
The Aggregator requires access to:
1. A configured PostgreSQL database with the imported Director's database schema.
2. An existing application with a `OPEN_RESOURCE_DISCOVERY` webhook. It can be created manually in the database or can be created via the Director GraphQL API. To run Director locally, see [Director](../director/README.md).
3. An API that can be called to fetch and process ORD documents.

### Run
Since the ORD Aggregator is usually a short-lived process, it is useful to start and debug it directly from your IDE.
Make sure that you provide all required configuration properties as environment variables.
