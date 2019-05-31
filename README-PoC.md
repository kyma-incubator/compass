# GraphQL Gateway PoC

## Description

This proof of concept contains Gateway GraphQL server, which communicates via gRPC with Director. Gateway delegates all operations to Director, which contains actual business logic.

From technical point of view, Gateway utilizes connection pool to be easily scalable. gRPC underneath uses HTTP/2 server-side push, so there is no need to reconnect for every request.

Not only queries and mutations, but also subscriptions are supported with gRPC using [gRPC server-side streaming](https://grpc.io/docs/guides/concepts/).

## Run locally

1. Run the following commands:
```bash
GATEWAY_ADDRESS=127.0.0.1:3000 DIRECTOR_ADDRESS=127.0.0.1:4000 go run components/gateway/cmd/main.go
DIRECTOR_ADDRESS=127.0.0.1:4000 go run components/director/cmd/main.go
```
1. Navigate to [http://localhost:3000/](http://localhost:3000/)

## Deploy on Kyma

The PoC can be easily deployed on Kyma. Both Gateway and Director have Istio sidecars injected.

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

Result: Resolver `applications` is triggered once (1 request) and returns two applications.

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

Result: Resolver `applications` is triggered once (1 request) and returns two applications. For each applications there is `apis` resolver triggered, which results in two calls. In this case, for two applications, we have 3 calls in total.


## gRPC Client generation

In this PoC, the Protocol Buffers schema for Director is located in the [`components/director/protobuf/director.proto`](`./components/director/protobuf/director.proto`) file.

From this schema we can generate Go client and server. As Gateway and Director are separate components, the code is generated twice in two separate locations, for Director and Gateway at the same time:

```go
//go:generate protoc --go_out=plugins=grpc:. director.proto
//go:generate protoc --go_out=plugins=grpc:../../gateway/protobuf/ director.proto
```

This approach have the following benefits:
- we can do changes in Gateway and Director in the same pull request
- there is no way to break compatibility between Director client and server when modifying just Director component: the generated code will be also updated for Gateway, and the CI build will validate the changes

The downsize is that we keep duplicates of generated files, but that shouldn't be an issue for us.

## Performance tests

Author: @aszecowka

Even if we reuse gRPC connections, with many requests, provided solution seems to not be the most efficient one. The following table represents results of querying applications depending on the items number.

| Test Case | Response Time |
| --------- | ------------- |
| 100 apps, each has 10 APIs | ~0.04s |
| 100 apps without querying APIs | 0.01s |
| 1000 apps, each has 10 APIs | ~ 0.5s |
| 1000 apps, without querying APIs | 0.04s |

The tests have been done on localhost machine.

## Summary

We have to duplicate the GraphQL schema and write it in gRPC, if we would go with this architecture. There is no point to do another layer of abstraction, which only brings complexity.
