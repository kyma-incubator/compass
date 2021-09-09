# Compass as a separate component

## Overview

This document describes the technical implementation of the Compass separation from Kyma.

Initially, Compass was planned to become a default component in Kyma, however, Kyma direction has changed recently. As Kyma will become more lightweight, it won't contain the built-in Compass component anymore.

## Reasons

Apart from the decision of not promoting Compass as a default component, the separation reduces the delivery time of Compass changes and makes the Compass development easier.

The new approach enables storing the Compass chart in a single place, without the need to synchronize changes between Compass and Kyma.

## Requirements

- Compass chart located in a single place
- Separate installation and upgrade of Compass chart
- Removal of the Compass component from Kyma Installer
- Migration of the pipelines for internal Compass environments to the new installation approach
  - Nice to have: Zero downtime migration for existing Compass installations

## Implementation

To achieve the separation, Compass is installed separately on top of the Kyma installation. The Compass chart is located in the Compass repository. All Compass-related resources and pipelines will be eventually removed from the Kyma repository.

Until the Runtime Agent is removed, the Kyma Compass GKE Integration pipeline and Compass-related External Solution tests on the Kyma repository must stay as they are.

In the first iteration, Compass is installed with the modified Kyma Operator, installing a single Compass component. In the future, we will investigate different approaches to install Compass, such as:

- Splitting the current Compass chart to multiple charts and using Kyma Installer to install multiple Compass components.
- Installing the Compass chart with pure Helm, without Kyma Installer.

### Steps

The following section lists all steps needed to implement the Compass separation:

- Modify Kyma Installer:

  - Separate overrides (configurable Namespace).
  - Make Installation CR Name and Namespace configurable.

- Move Compass Helm chart from Kyma to Compass.

- Define fixed Kyma release version in the Compass repo:

  - Define the Kyma release version in the `installation/resources/KYMA_RELEASE` file. The file should only contain the version, for example `main-0a8b78da` or `PR-8679`.

- Modify the Kyma Compass GKE Integration pipeline in the Kyma repository:

  - Install full Kyma with Kyma Installer (any overrides are placed in the Kyma repository).
  - Install Compass with Compass Operator (latest artifacts from the Compass repository) with custom overrides placed in the Kyma repository.

- Add the Prow pipeline to build and upload Compass Operator artifacts to a GCS bucket:

  - Base on the development-artifacts job.
  - Run it in the Compass repository.
  - Create Kyma Installer artifacts from the fixed Kyma version and upload them to the same bucket.

- Modify the `run.sh` development script in the Compass repository:

  - Install the lite version of Kyma with CLI from the fixed Kyma version using overrides from the Compass repository.
  - Install Compass with Compass Operator (using the new `run.sh` script).

- Modify the Kyma Compass Integration pipeline in the Compass repository:

  - Use the new `run.sh` script.

- Copy the Kyma Compass GKE Integration pipeline to the Compass repository:

  - Install the lite version of Kyma with Kyma Installer (any overrides are placed in the Compass repository; minimal list of components).
  - Install Compass with Compass Operator (use latest artifacts).

- Move the Prow GKE Provisioner tests pipeline from Kyma to Compass.

- Prepare the periodic Kyma Compass GKE integration job.

- Modify the API Gateway test:

  - Make Gateway configurable for test API rules (currently there is [a hardcoded `kyma-gateway`](https://github.com/kyma-project/kyma/blob/main/tests/integration/api-gateway/gateway-tests/manifests/no_access_strategy.yaml#L11); this component is used by KEB).

- Modify pipelines for internal Compass environments.

### Migration of existing Compass installations

This section covers zero downtime migration from the old to the new approach of the Compass installation.

Migration of the existing Compass installation doesn't require any custom operations.

With the migration to the new installation approach to the pipeline, the following things happen:

1. Kyma installation combo YAML is applied to the existing cluster.
1. Kyma Installer receives the new component list without Compass.
1. Kyma Installer upgrades all Kyma components that are dependent on Compass.

   As Compass is not on the components list, Kyma Installer ignores the existing Compass Helm release.

1. Compass installation combo YAML is applied to the existing cluster.

   **NOTE:** To make the migration as seamless as possible, it is recommended to have the override values identical to the previous Compass installation, as well as the Compass chart YAML files.

1. Compass Installer (Kyma Installer with the Compass component on the list) detects the already installed Compass Helm release and upgrades it.

### Proof of Concept

For the Proof of Concept stage, we decided to use the following temporary solutions:

- Create additional internal pipeline for PoC environment.

- Modify Kyma Developments Artifacts & Kyma Release Artifacts jobs.

  - Upload artifacts for Kyma installation setup needed for Compass (lite Kyma setup for Compass installation on top of it).

### Cleanup

- Remove any temporary resources and pipelines created during the Proof of Concept stage.
- Enable reporting for the pipeline generating Compass installer artifacts.
- Clean up building previous Compass artifacts from Kyma (Kyma with Compass component enabled).
- Remove Compass documentation from the Kyma repository.

Once the Runtime Agent is gone, we need to perform the following steps:

- Remove the Kyma Compass GKE Integration pipeline from the Kyma repository.
- Remove Connectivity Adapter and Compass E2E External Solution tests.
