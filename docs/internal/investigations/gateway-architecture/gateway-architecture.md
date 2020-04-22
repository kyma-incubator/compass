# Gateway architecture

## Overview

We discussed three different approaches to exposing our API to users.

1. Maintaining one handwritten GraphQL API on the gateway, that would call all other microservices (it could use for example gRPC)
2. Stitching multiple GraphQL schemas written for internal services into a single one and proxying the traffic to those internal GraphQL servers
3. Having separate HTTP endpoints proxying traffic to specific internal servers (exposing many separate GraphQL APIs)

## Possible solutions

### 1. Schema on Gateway delegating to other services via gRPC

The first option is to use a single GraphQL server on the gateway with manually written schema and resolvers, that would delegate the operations to internal services. We could use for example gRPC to communicate with them (a PoC of this can be found [here](https://github.com/kyma-incubator/compass/pull/21/)).

@aszecowka ran some tests, here are the results:

```
even if we reuse connections, with many requests, provided solution seems to be not the most efficient:
I queried for:

100 apps, each has 10 APIs: response time was ~0.04s
100 apps without querying about APIs: 0.01s
1000 apps, each has 10 APIs: response time ~ 0.5s
1000 apps, without querying about APIs: 0.04s
This was tested on the localhost.
```

Pros

- full GraphQL support (introspection + subscriptions)
- avoiding Node.js
- single endpoint exposed on the gateway, that would be stored by clients
- we can hide our internal APIs behind a "facade" that would allow us to make changes without breaking compatibility

Cons

- need to maintain almost identical protobuf schema to relay requests to internal services
- each resolver sends a separate call to proxied API (although it's worth noting that thanks to gRPC connection pool active connection can be maintained making the cost of remote calls not that high)

### 2. Multiple schemas stitched into one

#### 2.1. Apollo Server

The simplest way to stitch multiple GraphQL schemas would be to use the Node.js [Apollo Server](https://www.apollographql.com/), that already has the functionality to merge schemas and proxy traffic to specific internal graphql servers. Merging works by introspection of existing schemas, and it takes care of things such as type conflicts. It is even possible to combine and modify types and fields from different schemas. Apollo supports queries, mutations, and subscriptions. Examples of code needed to run Apollo Server with stitched schema can be found [here](https://www.contentful.com/blog/2019/01/30/combining-apis-graphql-schema-stitching-part-2/).

Pros

- easy to configure
- hosts fully functioning GraphQL server, so that API user can't tell he's using an abstraction over few separate APIs
- out of the box support for proxying queries, mutations, and subscriptions
- type conflict detection and resolution
- combining and editing types of merged schemas
- single endpoint exposed on the gateway, that would be stored by clients
- a single call to proxied API

Cons

- adding Node.js to our technology stack
- we expose 1:1 schemas from internal APIs
- introduces dependencies on internal APIs that have to expose their schemas before stitching, in case one of them is modified at runtime of system, the gateway needs to update its stitched schema

#### 2.2. Custom HTTP proxy in Go

Another idea was to implement an HTTP proxy that would first create a mapping of queries and remote servers they should access. Then it would forward received queries and mutations to remote servers using the mapped relations.

That would, however, require us to give up on features such as introspection because they would become hidden behind the proxy without hosting the merged schema on a real GraphQL server.

Handling subscriptions would be problematic as well because that would require proxying WebSocket traffic.

Pros

- avoiding Node.js
- single endpoint exposed on the gateway, that would be stored by clients
- a single call to proxied API

Cons

- lack of some GraphQL features (introspection)
- implementing subscriptions proxying which would be problematic because that would require proxying WebSocket traffic
- we expose 1:1 schemas from internal APIs
- introduces dependencies on internal APIs that have to expose their schemas before stitching, in case one of them is modified at runtime of system, the gateway needs to update its stitched schema

#### 2.3. Custom stitching implementation in Go

In this approach, we would have to either write our custom solution that would stitch remote schemas in a similar manner to how it's done in Apollo library or contribute to [gqlgen](https://github.com/99designs/gqlgen/issues/5). Either way, it seems like a lot of work to support all edge cases and the gqlgen issue that's open for over a year and still not implemented seems to be proof of that.

Pros

- avoiding Node.js
- hosts fully functioning GraphQL server, so that API user can't tell he's using an abstraction over few separate APIs
- support for proxying queries, mutations, and subscriptions
- type conflict detection and resolution
- combining and editing types of merged schemas
- single endpoint exposed on the gateway, that would be stored by clients
- a single call to proxied API

Cons

- seems like a huge amount of work
- we expose 1:1 schemas from internal APIs
- introduces dependencies on internal APIs that have to expose their schemas before stitching, in case one of them is modified at runtime of system, the gateway needs to update its stitched schema

### 3. Separate HTTP endpoints

This approach requires us to write a reverse proxy server that would expose a separate endpoint for each internal API we want to expose. This way we wouldn't be limited to GraphQL APIs, we could, for example, have one endpoint with REST API and a different one with GraphQL API.

Pros

- full GraphQL support (per endpoint)
- avoiding Node.js
- support for proxying queries, mutations, and subscriptions
- a single call to proxied API

Cons

- we expose 1:1 schemas from internal APIs
- implementing subscriptions proxying which would be problematic because that would require proxying WebSocket traffic
- multiple endpoints (that clients have to store)
- we expose our internal architecture

## Summary

Solution | Introspection & subscriptions support<br>(must have) | Single call to proxied API | Good performance<br>(Go) | Single endpoint | Not exposing internal APIs 1:1 | Low maintenance effort | Relative amount of work
:-:|:-:|:-:|:-:|:-:|:-:|:-:|:-:
Schema on Gateway delegating to other services via gRPC | ✓ | ✗ | ✓ | ✓ | ✓ | ✗ | medium
Apollo Server | ✓ | ✓ | ✗ | ✓ | ✗ | ✓ | very small
Custom HTTP proxy in Go | ✗ | ✓ | ✓ | ✓ | ✗ | ✓ | small*
Custom stitching implementation in Go | ✓ | ✓ | ✓ | ✓ | ✗ | ✓ | big
Separate HTTP endpoints | ✓<br>(per endpoint) | ✓ | ✓ | ✗ | ✗ | ✓ | small

\* Unless we decide to support subscriptions and introspection, then I believe the amount of work would be big.

## Conclusion

Taking into account all the pros and cons of mentioned solutions I believe that two options would work best in our case: **Apollo Server** and **Separate HTTP endpoints**. While Apollo Server is very handy to use and offers a lot of out of the box functionality it would introduce unnecessary change to our technology stack and performance hit. Because of that, I am personally leaning towards the more straightforward solution - **Separate HTTP endpoints** that seems to meet all our key requirements and doesn't seem to require a lot of work to implement.

## Decision

The decision made by the team is that we will implement third option (separate HTTP endpoints) but for now we will expose only one endpoint (`/graphql`) that will point at our `Director` service, hosting our current API. This way in the future we'll still be able to smoothly transition to a more complex solution.

The first solution was rejected due to high maintenance effort and performance issues. We decided to reject Apollo Server because we didn't want to introduce Node.js to our technology stack. The team was unsure how difficult implementing custom stitching solution in Go would be, so we rejected that option for now as well.