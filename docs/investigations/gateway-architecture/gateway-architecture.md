# Gateway architecture

## Overview

We discussed three different approaches to exposing our API to users.

1. Maintaining one handwritten GraphQL API on gateway, that would call all other microservices (it could use for example gRPC)
2. Stitching multiple GraphQL schemas written for internal services into a single one and proxying the traffic to those internal GraphQL servers
3. Having separate HTTP endpoints proxying traffic to specific internal servers (exposing many separate GraphQL APIs)

## Possible solutions

### 1. Handwritten GraphQL Gateway

First option is to use single GraphQL server on gateway with manually written schema and resolvers, that would delegate the operations to internal services. We could use for example gRPC to communicate with them (a PoC of this can be found [here](https://github.com/kyma-incubator/compass/pull/21/)).

Pros

- full GraphQL support (introspection + subscriptions)
- avoiding Node.js (better performance)
- single endpoint
- facade over internal APIs

Cons

- need to maintain almost identical protobuf schema to relay requests to internal services
- each resolver sends separate call to proxied API

### 2. Multiple schemas stitched into one

#### 2.1. Apollo Server

Simplest way to stitch multiple GraphQL schemas would be to use the Node.js [Apollo Server](https://www.apollographql.com/), that already has the functionality to merge schemas and proxy traffic to specific internal graphql servers. Merging works by introspection of existing schemas, and it takes care of things such as type conflicts. It is even possible to combine and modify types and fields from different schemas. Apollo supports queries, mutations and subscriptions.

Pros

- easy to configure
- actually hosts fully functioning GraphQL server, so that API user can't tell he's using a abstraction over few separate APIs
- out of the box support for proxying queries, mutations and subscriptions
- type conflict detection and resolution
- combining and editing types of merged schemas
- single endpoint
- single call to proxied API

Cons

- adding Node.js to our technology stack (worse performance)
- no facade over internal APIs

#### 2.2. Custom HTTP proxy in Go

Another idea was to implement a HTTP proxy that would first create a mapping of queries and remote servers they should access. Then it would forward received queries and mutations to remote servers using the mapped relations.

That would, however, require us to give up on features such as introspection because they would become hidden behind the proxy without hosting the merged schema on a real GraphQL server.

Handling subscriptions would be problematic as well, because that would require proxying websocket traffic.

Pros

- avoiding Node.js (better performance)
- single endpoint
- single call to proxied API

Cons

- lack of some GraphQL features (introspection)
- implementing subscriptions proxying
- no facade over internal APIs

#### 2.3. Custom stitching implementation in Go

In this approach we would have to either write our custom solution that would stitch remote schemas in similar manner to how it's done in Apollo library or contribute to [gqlgen](https://github.com/99designs/gqlgen/issues/5). Either way it seems like a lot of work to support all edge cases and the gqlgen issue that's open for over a year and still not implemented seems to be a proof of that.

Pros

- avoiding Node.js (better performance)
- actually hosts fully functioning GraphQL server, so that API user can't tell he's using a abstraction over few separate APIs
- support for proxying queries, mutations and subscriptions
- type conflict detection and resolution
- combining and editing types of merged schemas
- single endpoint
- single call to proxied API

Cons

- seems like a huge amount of work
- no facade over internal APIs

### 3. Separate HTTP endpoints

This approach requires us to write reverse proxy server that would expose a separate endpoint for each internal API we want to expose. This way we wouldn't be limited to GraphQL APIs, we could for example have one endpoint with REST API and different one with GraphQL API.

Pros

- full GraphQL support (per endpoint)
- avoiding Node.js (better performance)
- support for proxying queries, mutations and subscriptions
- single call to proxied API

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

\* Unless we decide to fake GraphQL behaviour to support all it's features, then I believe the amount of work would be big.