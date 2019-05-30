# GraphQL Gateway PoC

## Description

This proof of concept contains Gateway GraphQL server, which communicates via gRPC with Director. Gateway delegates all operations to Director, which contains actual business logic.

From technical point of view, Gateway utilizes connection pool to be easily scaleable. gRPC underneath uses HTTP/2 server-side push, so there is no need to reconnect for every request.

## Run locally

1. Run the following commands:
```bash
GATEWAY_ADDRESS=127.0.0.1:3000 DIRECTOR_ADDRESS=127.0.0.1:4000 go run components/gateway/cmd/main.go
DIRECTOR_ADDRESS=127.0.0.1:4000 go run components/director/cmd/main.go
```
1. Navigate to [http://localhost:3000/](http://localhost:3000/)

## Deploy on Kyma

The PoC can be easily deployed on Kyma. Both Gateway and Director have sidecars injected.

1. Run Kyma on Minikube
1. Run `install.sh` script in the `chart` directory
1. Navigate to [https://compass-gateway.kyma.local/](https://compass-gateway.kyma.local/)

## Test queries

Run the following query in playground:
```graphql
{
  applications {
    id
    name
    tenant
    status {
      timestamp
    }
    annotations
    labels
  }
}
```

Result: Only one resolver will be triggered (1 request).

Next, run this query:

```graphql
{
  applications {
    id
    name
    tenant
    status {
      timestamp
    }
    annotations
    labels
    apis {
      id
      targetURL
    }
  }
}
```

Result: Two resolvers will be triggered (3 requests in total).


## Summary

We have to duplicate the GraphQL schema and write it in gRPC, if we would go with this architecture. There is no point to do another layer of abstraction, which only brings complexity.
