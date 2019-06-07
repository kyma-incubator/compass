# Management Plane Components

## Overview

This document describes the [Management Plane's](./../terminology.md#Management-Plane) components.

## Components

Management Plane consists of [Compass](./../terminology.md#MP-Compass), [Runtime Provisioners](./../terminology.md#MP-Runtime-Provisioner) and [Cockpit](./../terminology.md#MP-Cockpit) components. [Applications](./../terminology.md#Application) and [Agents](./../terminology.md#Runtime-Agent) can communicate with the [Gateway](./../terminology.md#MP-Gateway) component or [Connector](./../terminology.md#MP-Connector) component. [Administrator](./../terminology.md#Administrator) uses Cockpit to configure Management Plane.

![Management Plane Components](./assets/mp-components.svg)

### Compass

Compass is the Management Plane Core that consists of three components: Connector, Gateway, and [Director](./../terminology.md#MP-Director).

#### Connector

Connector component exposes GraphQL API that can be accessed directly, its responsibility is establishing trust among Applications, Management Plane and [Runtimes](./../terminology.md#Runtime).

#### Gateway

Gateway component serves as the main API Gateway that extracts [Tenant](./../terminology.md#MP-Tenant) from incoming requests and proxies the requests to the Director component.

#### Director

Director component exposes GraphQL API that can be accessed through the Gateway component. It contains all business logic required to handle Applications and Runtimes registration as well as health checks. It also requests Application [Webhook API](./../terminology.md#Application-Webhook-API) for credentials. This component has access to storage.

### Cockpit

Cockpit component calls Management Plane APIs (in particular Compass and Runtime Provisioner APIs). This component is interchangeable.

### Runtime Provisioner

Runtime Provisioner handles the creation, modification, and deletion of Runtimes. This component is interchangeable.
