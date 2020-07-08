# Runtime Agent contract

## Overview

This document describes the contract between the Management Plane and the Runtime Agent. The main responsibilities of the Runtime Agent are:
- [Establishing trusted connection](#establishing-trusted-connection)
- [Renewing trusted connection](#renewing-trusted-connection)
- [Configuring the Runtime](#configuring-the-runtime)

## Establishing trusted connection

Runtime Agent automates the process of establishing trusted connection between the Runtime and the Management Plane. To initially trigger the process the Runtime Provisioner configures the Runtime Agent providing it with the required parameters to establish the trusted connection between the Management Plane and the Runtime. The Runtime Agent configuration in its minimal scope requires setting up the form of the initial secured connection to the Connector. Currently, Runtime Provisioner sets up on the Runtime a resource that holds the one-time token to enable the pairing. Runtime Agent triggered by the creation of the resource performs the pairing process and, in the end, the trusted connection between the Runtime and the Management Plane is established. At this point, further communication between the Runtime Agent and the Management Plane, and Application and the Runtime is possible.

## Renewing trusted connection

Depending on the type of trusted connection, during the runtime lifecycle, it may be required to periodically renew trusted connection.

## Synchronizing Runtime with Director

Runtime Agent is also responsible for synchronizing Applications in the Runtime down from the Director. This means it makes sure that the state in the Runtime matches the state in the Director.