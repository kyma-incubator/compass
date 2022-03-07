# Scopes Synchronizer Job

## Overview

The Scopes Synchronizer Job is responsible to synchronize the scopes of OAuth clients in hydra, created for a given consumer type, in case that consumer is granted more scopes, or some scopes are removed.

## Details

The Scopes Synchronizer basic workflow is as follows:

1. List all clients available in ORY Hydra
2. List all system auths with OAuths (should be 1:1 mapping with above clients)
3. Based on the system auth entry consumer type (application, runtime, integration system) update the client in Hydra with the required scopes taken from the [Director scopes configuration](../../../../chart/compass/charts/director/config.yaml).

## Configuration

The Scopes Synchronizer binary allows overriding of some configuration parameters. Up-to-date list of the configurable parameters can be found [here](https://github.com/kyma-incubator/compass/blob/8a8ecb2fcf3a38f8f6392f5669b98c1a10342363/components/director/cmd/scopessynchronizer/main.go#L27).

## Local Development
### Prerequisites
The Scopes Synchronizer requires access to:
1. Configured PostgreSQL database with the imported Director's database schema.
1. Up and running ORY Hydra.

### Run
There is no easy way to run this component locally, as you need ORY Hydra, so the recommended approach is to start a local Minikube installation and deploy Compass there. The Scopes Synchronizer job is a post-install job that runs at the very end of the Helm installation.
