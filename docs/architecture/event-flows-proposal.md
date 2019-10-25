# Eventing flow proposal

## Overview

This document is a proposal for eventing flows we want to support with use of the Management Plane.

## Eventing with Kyma internal Event Bus

![Management Plane Components](assets/internal-event-flow.svg)

1. The event is registered using **Management Plane** API.
2. The **Director** sends an event url to **EventGateway**.
3. The **Agent** fetches the event data and exposes it inside **Runtime**.  
4. The **Application** publishes an event.

>**NOTE** The Event Gateway will support multiple protocols like HTTP or MQTT which can be specified inside EventSpec data.
It will also support many authentication (e.g. OpenID Connect) and authorization (e.g. OAuth2) methods.
## Eventing with external Event Bus

![Management Plane Components](assets/external-event-flow.svg)

1. The event is registered using **Management Plane** API with external event bus data.
2. The **Agent** fetches the event and exposes it inside **Runtime**.
3. **Eventing system** subscribes on external event bus.
4. The application publishes an event.

## GraphQL Schema  

### Types
```graphql

scalar EventAPISubscriptionAttributes # -> map[string]string

type EventAPIDefinition {
    id: ID!
    """group allows you to find the same API but in different version"""
    group: String
    spec: EventAPISpec!
    auth: Auth
    externalSubscriptions: [EventAPISubscription!]!
    version: Version
}

type EventAPISpec {
    data: CLOB
    type: EventAPISpecType!
    format: SpecFormat
    fetchRequest: FetchRequest
}

type EventAPISubscription {
    url: String!
    auth: Auth
    topic: String!
    attributes: EventAPISubscriptionAttributes
}


enum EventAPISpecType {
    ASYNC_API
}

```

### Input Types

```graphql

input EventAPIDefinitionInput {
    spec: EventAPISpecInput!
    group: String
    auth: AuthInput
    externalSubscriptions: [EventAPISubscriptionInput!]
    version: VersionInput
}

input EventAPISpecInput {
    data: CLOB
    eventSpecType: EventAPISpecType!
    fetchRequest: FetchRequestInput
}

input EventAPISubscriptionInput {
    url: String!
    auth: AuthInput
    topic: String!
    attributes: EventAPISubscriptionAttributes
}
```

The `EventSpecType` type is the same as in `Types` section.
