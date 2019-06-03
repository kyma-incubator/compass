# Management Plane Components

## Overview

This document describes the Management Plane's components.

## Components

Management Plane consists of three separate components. Applications, Agents, and UIs can communicate with the Gateway component or Connector component.

Connector component exposes GraphQL API that can be accessed directly, its responsibility is the pairing of Applications and Runtimes.

Gateway component serves as the main API Gateway that extracts token from incoming requests and proxies the requests to the Director component.

Director component exposes GraphQL API that can be accessed through the Gateway component. It contains all business logic required to handle registered Applications and Runtimes requests. This component has access to storage.

![Management Plane Components](./assets/mp-components.svg)
