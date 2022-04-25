# Scopes Synchronizer Job

## Overview

The role of the Scopes Synchronizer Job is to synchronize the scopes of OAuth clients in Hydra, which are created for a given consumer type. The synchronization is required when there is a change in the consumer scopes, such as, granting additional scopes or removing given scopes.

## Details

The basic workflow of Scopes Synchronizer is as follows:

1. List all clients available in ORY Hydra
2. List all system auths with OAuths (should be 1:1 mapping with above clients)
3. Based on the system auth entry consumer type (application, runtime, integration system) update the client in Hydra with the required scopes taken from the [Director scopes configuration](../../../../chart/compass/charts/director/config.yaml).

## Configuration

The Scopes Synchronizer binary allows you to override some configuration parameters. To get a list of the configurable parameters, see [main.go](https://github.com/kyma-incubator/compass/blob/8a8ecb2fcf3a38f8f6392f5669b98c1a10342363/components/director/cmd/scopessynchronizer/main.go#L27).

## Local Development
### Prerequisites
The Scopes Synchronizer requires access to:
1. Configured PostgreSQL database with the imported Director's database schema.
1. Up and running ORY Hydra.

### Run
This component requires ORY Hydra and for this reason its local run is cumbersome. It is recommended to start a local Minikube installation and deploy Compass on it. The Scopes Synchronizer job is a post-install job that runs at the end of the Helm installation.
