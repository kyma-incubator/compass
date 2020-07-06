# Control Plane resources

This document describes all the resources needed for the Control Plane installation.

## Installation CRs

The following table lists all the Installation custom resource (CR) files needed for the Control Plane installation:

| File                                     | Description                                                                  |
| ---------------------------------------- | ---------------------------------------------------------------------------- |
| `installer-cr-kyma-dependencies.yaml`    | Contains Kyma components that are required for Control Plane installation.    |
| `installer-cr-compass-dependencies.yaml` | Contains Compass components that are required for Control Plane installation. |
| `installer-cr.yaml.tpl`                  | Contains components installed by the Control Plane Installer.                 |

## KYMA_VERSION file

`KYMA_VERSION` is the file that specifies the version of Kyma to use during the installation.

### Possible values

| Value                  | Example value     | Description             |
| ---------------------- | ----------------- | ----------------------- |
| `master`               | `master`          | Latest master version   |
| `master-{COMMIT_HASH}` | `master-34edf09a` | Specific master version |
| `PR-{PR_NUMBER}`       | `PR-1420`         | Specific PR version     |
| `{RELEASE_NUMBER}`     | `1.13.0`          | Release version         |

## COMPASS_VERSION file

`COMPASS_VERSION` is the file that specifies the version of Compass to use during the installation.

### Possible values

| Value                  | Example value     | Description             |
| ---------------------- | ----------------- | ----------------------- |
| `master`               | `master`          | Latest master version   |
| `master-{COMMIT_HASH}` | `master-34edf09a` | Specific master version |
| `PR-{PR_NUMBER}`       | `PR-1420`         | Specific PR version     |
