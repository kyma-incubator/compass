# Gateway architecture

## Overview

We discussed three different approaches to exposing our API to users.

1. Maintaining one handwritten GraphQL API on the gateway, that would call all other microservices (it could use for example gRPC)
2. Stitching multiple GraphQL schemas written for internal services into a single one and proxying the traffic to those internal GraphQL servers
3. Having separate HTTP endpoints proxying traffic to specific internal servers (exposing many separate GraphQL APIs)

## Possible solutions

### 1. Handwritten GraphQL Gateway

The first option is to use a single GraphQL server on the gateway with manually written schema and resolvers, that would delegate the operations to internal services. We could use for example gRPC to communicate with them (a PoC of this can be found [here](https://github.com/kyma-incubator/compass/pull/21/)).

Pros

- full GraphQL support (introspection + subscriptions)
- avoiding Node.js (better performance)
- single endpoint
- facade over internal APIs

Cons

- need to maintain almost identical protobuf schema to relay requests to internal services
- each resolver sends a separate call to proxied API (although it's worth noting that thanks to gRPC connection pool active connection can be maintained making the cost of remote calls not that high)

### 2. Multiple schemas stitched into one

#### 2.1. Apollo Server

The simplest way to stitch multiple GraphQL schemas would be to use the Node.js [Apollo Server](https://www.apollographql.com/), that already has the functionality to merge schemas and proxy traffic to specific internal graphql servers. Merging works by introspection of existing schemas, and it takes care of things such as type conflicts. It is even possible to combine and modify types and fields from different schemas. Apollo supports queries, mutations, and subscriptions.

Pros

- easy to configure
- hosts fully functioning GraphQL server, so that API user can't tell he's using an abstraction over few separate APIs
- out of the box support for proxying queries, mutations, and subscriptions
- type conflict detection and resolution
- combining and editing types of merged schemas
- single endpoint
- a single call to proxied API

Cons

- adding Node.js to our technology stack (worse performance)
- no facade over internal APIs

#### 2.2. Custom HTTP proxy in Go

Another idea was to implement an HTTP proxy that would first create a mapping of queries and remote servers they should access. Then it would forward received queries and mutations to remote servers using the mapped relations.

That would, however, require us to give up on features such as introspection because they would become hidden behind the proxy without hosting the merged schema on a real GraphQL server.

Handling subscriptions would be problematic as well because that would require proxying WebSocket traffic.

Pros

- avoiding Node.js (better performance)
- single endpoint
- a single call to proxied API

Cons

- lack of some GraphQL features (introspection)
- implementing subscriptions proxying
- no facade over internal APIs

#### 2.3. Custom stitching implementation in Go

In this approach, we would have to either write our custom solution that would stitch remote schemas in a similar manner to how it's done in Apollo library or contribute to [gqlgen](https://github.com/99designs/gqlgen/issues/5). Either way, it seems like a lot of work to support all edge cases and the gqlgen issue that's open for over a year and still not implemented seems to be proof of that.

Pros

- avoiding Node.js (better performance)
- hosts fully functioning GraphQL server, so that API user can't tell he's using an abstraction over few separate APIs
- support for proxying queries, mutations, and subscriptions
- type conflict detection and resolution
- combining and editing types of merged schemas
- single endpoint
- a single call to proxied API

Cons

- seems like a huge amount of work
- no facade over internal APIs

### 3. Separate HTTP endpoints

This approach requires us to write a reverse proxy server that would expose a separate endpoint for each internal API we want to expose. This way we wouldn't be limited to GraphQL APIs, we could, for example, have one endpoint with REST API and a different one with GraphQL API.

Pros

- full GraphQL support (per endpoint)
- avoiding Node.js (better performance)
- support for proxying queries, mutations, and subscriptions
- a single call to proxied API

Cons

- no facade over internal APIs
- implementing subscriptions proxying
- multiple endpoints

## Summary

Solution | Introspection & subscriptions support | Single call to proxied API | Good performance<br>(Go) | Single endpoint | Facade over internal components | No need to maintain almost identical protobuf schema | Relative amount of work
:-:|:-:|:-:|:-:|:-:|:-:|:-:|:-:
Handwritten GraphQL Gateway | ✓ | ✗ | ✓ | ✓ | ✓ | ✗ | medium
Apollo Server | ✓ | ✓ | ✗ | ✓ | ✗ | ✓ | very small
Custom HTTP proxy in Go | ✗ | ✓ | ✓ | ✓ | ✗ | ✓ | small*
Custom stitching implementation in Go | ✓ | ✓ | ✓ | ✓ | ✗ | ✓ | big
Separate HTTP endpoints | ✓<br>(per endpoint) | ✓ | ✓ | ✗ | ✗ | ✓ | small

\* Unless we decide to fake GraphQL behavior to support all its features, then I believe the amount of work would be big.

## Conclusion

Taking into account all the pros and cons of mentioned solutions I believe that two options would work best in our case: **Apollo Server** and **Separate HTTP endpoints**. While Apollo Server is very handy to use and offers a lot of out of the box functionality it would introduce unnecessary change to our technology stack and performance hit. Because of that, I am personally leaning towards the more straightforward solution - **Separate HTTP endpoints** that seems to meet all our key requirements and doesn't seem to require a lot of work to implement.