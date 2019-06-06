# Eventing flow 

## Overview

This document describes two ways of eventing flow we will support with the use of [Management Plane](./terminology.md#management-plane)

## Eventing with Kyma internal Event Bus

![Management Plane Components](assets/internal-event-flow.svg)

1. The event is registered using **Management Plane** API.
2. The **Director** sends an event url to Application.
3. The **Agent** fetches the event data and passes it to Kyma Event Bus.
4. The Application publishes an event.

## Eventing with external Event Bus

![Management Plane Components](assets/external-event-flow.svg)

1. The event is registered using **Management Plane** API with external event bus data.
2. The **Agent** fetches the event and passes information to eventing system.
3. Eventing system subscribes on external event bus.
4. The application publishes an event.