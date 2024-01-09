# ORY

[ORY](https://www.ory.sh/) Open Source OAuth 2.0 & OpenID Connect

## Introduction

This chart bootstraps [Hydra](https://www.ory.sh/docs/hydra/) and [Oathkeeper](https://www.ory.sh/docs/oathkeeper/) components on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Caveats

The current state of `values.yaml` has been configured to work with the `run.sh` script; this script is used for local installation as explained in the [installation document](https://github.com/kyma-incubator/compass/blob/main/docs/compass/04-01-installation.md).

## Oathkeeper chart

The Oathkeeper chart has custom implementations that are marked with `# Custom...`; the most notable being the `pre-install-job` and `cronjob` which handle the creation
and rotation of the Oathkeeper JSON Web Key Set Secret.