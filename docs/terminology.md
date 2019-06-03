# Terminology

## Overview
This document contains all terminology used across compass documentation.

## Management Plane 
Abbr.: *MP*\
\
The multi-tenant system which allows to:
- configure Applications
- configure Runtimes
- group Applications
- assign Applications or Runtimes to the group

### MP Connector
Abbr.: *Connector*\
\
Component that establish trust among Applications, Management Plane and Runtimes. In first iteration we support only client certificates.

### MP Gateway
Abbr.: *Gateway*\
\
Component that.. TBD

### MP Director
Abbr.: *Director*\
\
Component that.. TBD

### MP Healtchecker
Abbr.: *Healtchecker*\
\
Component that.. TBD

### MP Tenant
Abbr.: *Tenant*\
\
TBD

## Application
Existing system registered to *MP* with its *API and Event Definitions*.

### Application API Definiton
Abbr.: *API Definiton*

### Application Event Definiton
Abbr.: *Event Definiton*

### Application Webhook API
Abbr.: *Webhook API*

### Application Documentation

## Runtime
Any system that can configure itself according to the configuration provided by the *Management Plane*. Takes care about a customer workload.

### Runtime Agent
Abbr.: *Agent*  

Component responsible:
- to fetch configuration from *MP* to *Runtime*.
- for reporting health checks

## Administrator

User who:
- configures *Applications* and *Runtimes* in the *Management Plane*. 
- groups *Applications* and *Runtimes*.
