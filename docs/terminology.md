# Terminology

## Overview
This document contains all terminology used across compass documentation.

## Management Plane 
Abbr.: *MP*\
\
Includes *Compass*, *Runtime Provisioners* and *Cockpit* components.

The multi-tenant system which allows to:
- create Applications
- create Runtimes
- manage Applications and Runtimes

### MP Compass
Abbr.: *Compass*\
\
Includes *Connector*, *Gateway*, *Director* and *Healtchecker* components.

The multi-tenant system which allows to:
- configure Applications
- configure Runtimes
- assign Applications or Runtimes to the group

### MP Connector
Abbr.: *Connector*\
\
Connector component establishes trust among Applications, Management Plane and Runtimes. In first iteration we support only client certificates.

### MP Gateway
Abbr.: *Gateway*\
\
Gateway component serves as the main API Gateway that extracts *Tenant* from incoming requests and proxies the requests to the Director component.

### MP Director
Abbr.: *Director*\
\
Director component is mainly responsible for *Applications* and *Runtimes* registration. In addition, requests *Appliction Webhook API* for credentials and exposes health information about *Runtimes*.

### MP Runtime Provisioner
Abbr.: *Provisioner*\
\
Runtime Provisioner system manages *Runtimes*.

### MP Cockpit
Abbr.: *Cockpit*\
\
Cockpit component calls *Management Plane* APIs (in particular *Compass* and *Runtime Provisioner* APIs).

### MP Tenant
Abbr.: *Tenant*\
\
Represents customer tenant.

## Application
Existing system registered to *MP* with its *API and Event Definitions*.

### Application API Definiton
Abbr.: *API Definiton*

### Application Event Definiton
Abbr.: *Event Definiton*

### Application Webhook API
Abbr.: *Webhook API*

### Application Documentation

## Integration System
Any system that works in context of multiple tenants, managing multiple Applications or Runtimes.

## Runtime
Any system that can configure itself according to the configuration provided by the *Management Plane*. Takes care about a customer workload.

### Runtime Agent
Abbr.: *Agent*  

This component is responsible:
- to fetch configuration from *MP* to *Runtime*.
- for reporting health checks

## Administrator

The User who:
- configures *Applications* and *Runtimes* in the *Management Plane*. 
- groups *Applications* and *Runtimes*.
