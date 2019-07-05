# Compass End-To-End tests

Compass end-to-end consists of acceptance tests.
Director folder contains tests of GraphQL API for managing Applications and Runtimes.

## Usage
To run Director tests, execute:
```
env DIRECTOR_GRAPHQL_API={URL} go test -v ./director/...
```