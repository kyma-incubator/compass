# ORY

[ORY](https://www.ory.sh/) Open Source OAuth 2.0 & OpenID Connect

> **_NOTE:_** The chart was copied from [Kyma](https://github.com/kyma-project/kyma/tree/main/resources/ory), this means that there are custom implementations

## Introduction

This chart bootstraps [Hydra](https://www.ory.sh/docs/hydra/) and [Oathkeeper](https://www.ory.sh/docs/oathkeeper/) components on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager. 

Furthermore, it includes two other subcharts for storage persistence:

- [gcloud-sqlproxy](https://github.com/rimusz/charts/tree/master/stable/gcloud-sqlproxy) - used to establish connection with databases situated on Google Cloud
- [postgres](https://github.com/bitnami/charts/tree/main/bitnami/postgresql) - used for in-cluster storage persistence

## Caveats

The current state of `values.yaml` has been configured to work with the `run.sh` script; this script is used for local installation as explained in the [installation document](https://github.com/kyma-incubator/compass/blob/main/docs/compass/04-01-installation.md).

### Hydra Secret

The Hydra deployment expects an already existing K8s Secret with the name `existingSecret: "ory-hydra-credentials"`. The K8s Secret is created by the `installation/scripts/install-ory.sh` script, which is invoked by `run.sh`. This approach was chosen as the Hydra secrets need to be rotated in [a specific manner](https://www.ory.sh/docs/hydra/self-hosted/secrets-key-rotation#rotation-of-hmac-token-signing-and-database-and-cookie-encryption-keys); if the Secret was dynamically created with each Helm release it would break the deployment.

### Oathkeeper Secret

The Oathkeeper Secret is created by a CronJob that is part of the Helm release. However, the job needs to be patched and configured properly; this is also handled by `installation/scripts/install-ory.sh`. This interaction requires that the Helm release is installed by the Helm CLI without the `--wait` flag.
