# Open Resource Discovery Aggregator

## Overview

The Aggregator application collects, aggregates, and stores the ORD information from multiple ORD providers to the Compass's database.


## Details

The Aggregator basic workflow is as follows:

1. The Aggregator goes through all available Applications, stored in the Compass's database.
2. For each Application that has a Webhook of type `OPEN_RESOURCE_DISCOVERY` it calls the URL that is attached to that Webhook.
3. That URL has predefined endpoints, which provide the necessary information to the Aggregator.
4. The Aggregator aggregates and stores the provided information in the Compass's database.

## Configuration

The ORD Aggregator binary allows overriding of some configuration parameters. Up-to-date list of the configurable parameters can be found [here](https://github.com/kyma-incubator/compass/blob/75aff5226d4a105f4f04608416c8fa9a722d3534/components/director/cmd/ordaggregator/main.go#L49).

## Local Development
### Prerequisites
The Aggregator requires access to:
1. Configured PostgreSQL database with the imported Director's database schema.
2. Pre-existing application with a `OPEN_RESOURCE_DISCOVERY` webhook - it can be created manually in the DB, but can also be created via the Director GraphQL API. Read [this](../director/README.md) document to learn how to run Director locally.
3. API that can be called to fetch and process ORD Documents.

### Run
Since the ORD Aggregator is usually a short-lived process, it is useful to start and debug it directly from your IDE.
Make sure to provide all required configuration properties as environment variables.
