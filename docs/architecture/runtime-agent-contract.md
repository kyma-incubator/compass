# Runtime Agent contract

## Overview

This document describes contract between [Management Plane](../terminology.md#management-plane) and [Runtime Agent](../terminology.md#runtime-agent). Main responsibilities of the Runtime Agent are:
- [Establishing trusted connection](#establishing-trusted-connection)
- [Renewing trusted connection](#renewing-trusted-connection)
- [Configuring the runtime](#configuring-the-runtime)

## Establishing trusted connection

Runtime Agent automates the process of [establishing trusted connection](./establishing-trusted-connection.md#establishing-trusted-connection) between [Runtime](../terminology.md#runtime) and Managemt Plane. To initialy trigger the process the [Runtime Provisioner](../terminology.md#mp-runtime-provisioner) creates the instance of CustomResource that holds the one-time token to enable the pairing. Runtime Agent triggered by the creation of CustomResource performs the pairing process and in the end the trusted connection between the Runtime and the Management Plane is established. At this point further communication between Runtime Agent and Management Plane, and [Application](../terminology.md#application) and Runtime is possible.

## Renewing trusted connection

Depending on the type of trusted connection, during the runtime lifecycle, it may be required to periodically [renew trusted connection](./establishing-trusted-connection.md#client-certificate-flow---certificate-renewal).

## Configuring the runtime

Runtime Agent periodically requests for the configuration of its Runtime from the Management Plane. Changes in the configuration for the Runtime are applied by the Runtime Agent on the Runtime. To fetch the Runtime configuration the Runtime Agent calls `applicationsForRuntime(runtimeId: ID!)` query offered by [Director](../terminology.md#mp-director). The response for the query contains list of Applications assigned for the Runtime.

Runtime Agent reports back to the Director the Runtime specific LabelDefinitions that represent runtime configuration together with their values.

Runtime specific LabelDefinitions:

- Events Gateway URL
- Runtime Console URL