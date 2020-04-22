# Runtime Agent contract

## Overview

This document describes the contract between the Management Plane and the Runtime Agent. The main responsibilities of the Runtime Agent are:
- [Establishing trusted connection](#establishing-trusted-connection)
- [Renewing trusted connection](#renewing-trusted-connection)
- [Configuring the Runtime](#configuring-the-runtime)

## Establishing trusted connection

Runtime Agent automates the process of [establishing trusted connection](../archive/establishing-trusted-connection.md#establishing-trusted-connection) between the Runtime and the Management Plane. To initially trigger the process the Runtime Provisioner configures the Runtime Agent providing it with the required parameters to establish the trusted connection between the Management Plane and the Runtime. The Runtime Agent configuration in its minimal scope requires setting up the form of the initial secured connection to the Connector. Currently, Runtime Provisioner sets up on the Runtime a resource that holds the one-time token to enable the pairing. Runtime Agent triggered by the creation of the resource performs the pairing process and, in the end, the trusted connection between the Runtime and the Management Plane is established. At this point, further communication between the Runtime Agent and the Management Plane, and Application and the Runtime is possible.

## Renewing trusted connection

Depending on the type of trusted connection, during the runtime lifecycle, it may be required to periodically [renew trusted connection](../archive/establishing-trusted-connection.md#client-certificate-flow---certificate-renewal).

## Configuring the Runtime

> **NOTE:** To represent API and Event Definitions of the connected Applications on Runtime, Open Service Broker API usage is recommended.

In Kyma Runtime, during Runtime configuration, Application's Packages are integrated into the Service Catalog using [Application](https://kyma-project.io/docs/components/application-connector#custom-resource-application) custom resources and [Application Broker](https://kyma-project.io/docs/components/application-connector#architecture-application-broker). By default, a single Application is represented as a ServiceClass, and a single Package is represented as a ServicePlan in the Service Catalog. To learn more about Packages, read [this](../compass/03-packages-api.md) document.

Runtime Agent periodically requests for the configuration of its Runtime from the Management Plane. Changes in the configuration for the Runtime are applied by the Runtime Agent on the Runtime.

To fetch the Runtime configuration, Runtime Agent calls `applicationsForRuntime(runtimeId: ID!, first: Int = 100, after: PageCursor)` query offered by the Director. The response for the query contains a page with the list of Applications assigned for the Runtime and info about the next page. Each Application will contain only credentials that are valid for the Runtime that called the query. Each Runtime Agent can fetch the configurations for Runtimes that belong to its tenant, there is no validation if the Runtime Agent is fetching the configuration for the Runtime on which it runs.

Runtime Agent reports back to the Director the Runtime-specific LabelDefinitions that represent Runtime configuration together with their values.

Runtime specific LabelDefinitions:

- Events Gateway URL
- Runtime Console URL
