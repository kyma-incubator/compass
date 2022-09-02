# Compass - 1

## Overview

Compass is a multi-tenant system which consists of components that provide a way to register, group, and manage your applications across multiple Kyma runtimes. Compass consists of the following sub-charts:
- `director`
- `connector`
- `gateway`
- `cockpit`
- `ord-service`
- `connectivity-adapter`
- `pairing-adapter`
- `operations-controller`
- `tenant-fetcher`
- `system-broker`
- `postgresql`
- `prometheus-postgres-exporter`
- `external-services-mock` (not recommended to be deployed on production environments)

To learn more about these components, see the [Compass](https://github.com/kyma-incubator/compass/blob/main/README.md) documentation.

The Cockpit and ORD Service components are located in separate GitHub repositories:
- Cockpit: [kyma-incubator/compass-console](https://github.com/kyma-incubator/compass-console)
- ORD Service: [kyma-incubator/ord-service](https://github.com/kyma-incubator/ord-service)

## Details

### Configuration

Compass has a standard Helm chart configuration. You can check all available configurations in the chart, and sub-charts's `values.yaml` files.

The values from those files can be overridden during installation via `ConfigMaps` created in the namespace where the Compass Installer is running. 
> **Note:** The `ConfigMaps` with overrides must be created before the installation is initiated.

**Example**

`chart/compass/values.yaml`:
```yaml
global:
    ingress:
        domainName: compass.com
    director:
        clientIDHeaderKey: client-id
```
`chart/compass/charts/director/values.yaml`:
```yaml
deployment:
  minReplicas: 1
  maxReplicas: 1
```

These values can be overridden with the following override file passed to Compass installation script:
```yaml
global:
    ingress:
        domainName: dev.compass.com
    director:
        clientIDHeaderKey: X-Client-ID
director:
    deployment:
        maxReplicas: 3
```
