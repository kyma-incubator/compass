# Splitting the Compass project


This document describes the technical implementation of the Compass project split into two separate projects.

Previously, Compass consisted of two main areas: Application Connectivity and Runtime provisioning. The mixed responsibilities of Compass caused some misunderstanding of the project. The goal of splitting the Compass project is to separate the responsibilities and develop these two areas independently.

The Compass project split is the next step after separating Compass from Kyma. To read more about the separation, see the [Compass as a separate component](../separate-compass/separate-compass.md) document.

## Reasons

There are two main reasons for the Compass project split:

- Separation of responsibilities: Application Connectivity components are separated from Runtime-related components.
- Change of the Application Connectivity components ownership.

## Implementation

The Compass project is split into two repositories:

- Compass, which contains the Application Connectivity components and its dependencies:
  - Director
  - Connector
  - Gateway
  - Pairing Adapter
  - External Services Mock
  - Connectivity Adapter
  - Schema Migrator
- Kyma Control Plane (KCP), which contains Runtime-related components and its dependencies:
  - Kyma Environment Broker
  - Runtime Provisioner
  - Metris
  - Kubeconfig Service
  - Schema Migrator

Schema Migrator source code and chart stay in both places to handle database migrations for Application Connectivity and Runtime-related components.

Additionally, the PostgreSQL chart is copied to both repositories as it is a dependency for both groups of components. 

To preserve Git history, the complete Compass history is pushed to the Kyma Control Plane repository. All changes are made on top of the Compass history, including Application Connectivity components removal and code import changes.

### Steps

The following steps have to be executed to complete the Compass project split:

- Create new repository `control-plane` under `kyma-project` GitHub organization

  - Follow the process described in the `kyma-project/community` repository.
  - Configure the repository (add labels, branch protection, etc.).

- Split Compass chart.

  - Split Kyma Control Plane components from the Compass chart to new repository.
  - Remove references between Compass and KCP charts.
  - Make the Compass-KCP communication external.
  - Modify components, if needed.
  - Remove KCP components from the Compass repo and Compass components from the KCP repo.

- Create Kyma Control Plane Installer.

  - Base on the Compass Installer (copy and modify the needed files).
  - Prepare a Prow pipeline to build the KCP installer.
  - Prepare installation CRs for Kyma, Compass, and KCP.
    - Use separate Namespaces.

- Define a fixed Compass release version in the Kyma Control Plane repo.

- Copy the `KYMA_VERSION` file to the Kyma Control Plane repo.

- Modify the `run.sh` script to install Kyma (from the `KYMA_VERSION` file), Compass (from the `COMPASS_VERSION` file), and KCP straight from the repo.
- Prepare the KCP-integration Prow pipeline for the new KCP repo (base on the existing Compass-integration job).
- Prepare the KCP-GKE-integration Prow pipeline for the new KCP repo (base on the existing Compass-GKE-integration job on Compass).

- Duplicate the GKE-Compass-Provisioner tests Prow pipeline and run it on the KCP repo.

- Prepare Prow jobs for building the KCP components.

- Prepare the KCP development artifacts Prowjob.

  - Base on Compass artifacts jobs.

- Prepare the installation CR for KCP.

  - Use separate Namespaces.

- Split the documentation.

  - Discuss how the docs should look on the website (which node to use).

- Configure TestGrid to display the KCP dashboard.

- Modify pipelines for the internal environments:

  - Watch the KCP repo for changes.
  - Install Compass from the fixed `COMPASS_VERSION` from the KCP repo.
  - Modify the file from which we get the Kyma version to install (install Kyma from the fixed `KYMA_VERSION` file from the KCP repo).
  - Install the KCP installer as a new step.

### Migration of the existing Compass installations

Migration of the existing Compass installation doesn't require any custom operations.

Once a pipeline is modified to install Kyma Control Plane components from the `control-plane` repository on top of the existing Compass installation, the transition period begins. In the transition period, there will be duplicated KCP components on the cluster. They will be deployed in two different Namespaces: `compass-system` and `kcp-system`. The transition period ends when the Compass chart is modified and all KCP components are deleted from the Compass chart.

### Cleanup

Once the Compass components are migrated to a separate internal environment:

- Remove the Compass installation step from internal environment pipelines.
- Remove the registration job from Kyma Control Plane.
