# Management Plane Components

## Overview

This document describes the [Management Plane's](/docs/terminology.md#Management-Plane) components and the high-level concepts.

## Components

The Management Plane is an abstract definition and set of exposed functionality on how users can managed different aspects of their application landscape allowing flexible approaches of extending, customizing and integrating their existing application solutions. 

The Management Plane consists of the Management Plane Services (Project [Compass](/docs/terminology.md#MP-Compass)), Manage Plane Integration Services, [Runtime Provisioners](/docs/terminology.md#MP-Runtime-Provisioner) and [Cockpit](/docs/terminology.md#MP-Cockpit) components. The Management Plane Services (Project [Compass](/docs/terminology.md#MP-Compass)) are a set of headless services covering all the generic functionality while optionally leveraging different application specific Management Plane Integration Services to configure and instrument the application to be integrated or extended. All communication, whether it comes from a [Applications](/docs/terminology.md#Application) or other external component is flowing through the [API-Gateway](/docs/terminology.md#MP-Gateway) component. [Administrator](/docs/terminology.md#Administrator) uses Cockpit to configure Management Plane.

![Management Plane Components](./assets/mp-components.svg)

### Compass

Compass is the Management Plane Core that consists of three components: Connector, Gateway, and [Director](/docs/terminology.md#MP-Director).

#### Connector

Connector component exposes GraphQL API that can be accessed directly, its responsibility is establishing trust among Applications, Management Plane and [Runtimes](/docs/terminology.md#Runtime).

#### Gateway

Gateway component serves as the main API Gateway that extracts [Tenant](/docs/terminology.md#MP-Tenant) from incoming requests and proxies the requests to the Director component.

#### Director

Director component exposes GraphQL API that can be accessed through the Gateway component. It contains all business logic required to handle Applications and Runtimes registration as well as health checks. It also requests Application [Webhook API](/docs/terminology.md#Application-Webhook-API) for credentials. This component has access to storage.

### Cockpit

Cockpit component calls Management Plane APIs (in particular Compass and Runtime Provisioner APIs). This component is interchangeable.

### Runtime Provisioner

Runtime Provisioner handles the creation, modification, and deletion of Runtimes. This component is interchangeable.
