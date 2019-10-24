# Runtime Agent contract

## Overview

This document describes contract between Management Plane and Runtime Agent. Main responsibilities of the Runtime Agent are:
- [Establishing trusted connection](#establishing-trusted-connection)
- [Renewing trusted connection](#renewing-trusted-connection)
- [Configuring the runtime](#configuring-the-runtime)

## Establishing trusted connection

Runtime Agent automates the process of [establishing trusted connection](./establishing-trusted-connection.md#establishing-trusted-connection) between Runtime and Management Plane. To initially trigger the process the Runtime Provisioner configures the Runtime Agent providing it the required parameters to establish the trusted connection between Management Plane and the Runtime. The Runtime Agent configuration in its minimal scope requires setting up the form of the initial secured connection to the Connector. Currently, Runtime Provisioner sets up on the Runtime a resource that holds the one-time token to enable the pairing. Runtime Agent triggered by the creation of the resource performs the pairing process and in the end, the trusted connection between the Runtime and the Management Plane is established. At this point further communication between Runtime Agent and Management Plane, and Application and Runtime is possible.

## Renewing trusted connection

Depending on the type of trusted connection, during the runtime lifecycle, it may be required to periodically [renew trusted connection](./establishing-trusted-connection.md#client-certificate-flow---certificate-renewal).

## Configuring the runtime

Runtime Agent periodically requests for the configuration of its Runtime from the Management Plane. Changes in the configuration for the Runtime are applied by the Runtime Agent on the Runtime. To fetch the Runtime configuration the Runtime Agent calls `applicationsForRuntime(runtimeId: ID!, first: Int = 100, after: PageCursor)` query offered by Director. The response for the query contains a page with list of Applications assigned for the Runtime and info about next page. Each Application will contain only credentials that are valid for the runtime that called the query. Each Runtime Agent can fetch the configurations for Runtimes that belong to its tenant, there is no validation if the Runtime Agent is fetching the configuration for the runtime on which it runs.

Runtime Agent reports back to the Director the Runtime specific LabelDefinitions that represent runtime configuration together with their values.

Runtime specific LabelDefinitions:

- Events Gateway URL
- Runtime Console URL
