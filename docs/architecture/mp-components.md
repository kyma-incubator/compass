# Management Plane Components

## Overview

This document describes the [Management Plane's](./../terminology.md#Management-Plane) components and the high-level concepts.

## Components

The Management Plane is an abstract definition and set of exposed functionality on how users can managed different aspects of their application landscape allowing flexible approaches of extending, customizing and integrating their existing application solutions.

The Management Plane consists of the Management Plane Services (Project [Compass](./../terminology.md#MP-Compass)), Manage Plane Integration Services, [Runtime Provisioners](./../terminology.md#MP-Runtime-Provisioner) and [Cockpit](./../terminology.md#MP-Cockpit) components. The Management Plane Services (Project [Compass](./../terminology.md#MP-Compass)) are a set of headless services covering all the generic functionality while optionally leveraging different application specific Management Plane Integration Services to configure and instrument the application to be integrated or extended. All communication, whether it comes from a [Applications](./../terminology.md#Application) or other external component is flowing through the [API-Gateway](./../terminology.md#MP-Gateway) component. [Administrator](./../terminology.md#Administrator) uses Cockpit to configure Management Plane.


![Management Plane Components](./assets/mp-components.svg)

### Management Plane Services Compass

Compass is the project name of the Management Plane Services that consists of three components: Connector, Gateway, Registry and [Director](./../terminology.md#MP-Director).

### Management Plane Integration

Define the abstract protocol an application has to support to be orchestrated through the manage plane. The protocol might be also provided by surrogate components which implement the protocol flow and instrument and configure the application.

There are different level of Integration:

- **basic** - Application registration is done via static Application and API/Events Metadata. Mainly used for simple use-case scenarios and doesn't support all management plane features
- **application** - Management Plane Integration is build into the application
- **proxy** - A proxy component colocated to the application is providing the Management Plane Integration and is controlling the application. The proxy can be highly application specific.
- **service** - A central service is providing the Management Plane Integration for a class of application managing multiple instances of these applications. Multiple service can be integrated to support different type of applications.

### Connector

Connector component exposes GraphQL API that can be accessed directly, its responsibility is establishing trust among Applications, Management Plane and [Runtimes](./../terminology.md#Runtime).

### API-Gateway

API-Gateway component serves as the main Gateway that extracts [Tenant](./../terminology.md#MP-Tenant) from incoming requests and proxies the requests to the Director component.

### Director

Director component exposes GraphQL API that can be accessed through the Gateway component. It contains all business logic required to handle Applications and Runtimes registration as well as health checks. It also requests Application [Webhook API](./../terminology.md#Application-Webhook-API) for credentials. This component has access to storage.

### Cockpit

Cockpit component calls Management Plane APIs (in particular Compass and Runtime Provisioner APIs). This component is interchangeable.

### Registry

The registry component serves as the persistent storage interface.


### Runtime Provisioner

Runtime Provisioner handles the creation, modification, and deletion of Runtimes. This component is interchangeable.
