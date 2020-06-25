# Splitting Compass project

## Overview

This document describes the technical implementation of the Compass project split into two separate projects.

Previously, Compass consisted of two main areas: Application Connectivity and Runtime provisioning. Mixed concerns of Compass made confusion around understanding of the project. The goal of the splitting Compass project is to separate the concerns and develop these two areas independently.

Compass project split is a next step after making Compass separated from Kyma. To read more about the separation, see the [Compass as a separate component](../separate-compass/separate-compass.md) document.

## Reasons

There are two main reasons for the Compass project split:

- Separation of concerns: Application Connectivity components are separated from Runtime-related components.
- Change of the Application Connectivity components ownership.

## Implementation

The Compass project is split into two repositories:

- Compass, which contains Application Connectivity components:
  - Director
  - Connector
  - Gateway
  - Pairing Adapter
  - External Services Mock
  - Connectivity Adapter
  - Schema Migrator
- Kyma Control Plane, which contains Runtime-related components:
  - Kyma Environment Broker
  - Provisioner
  - Metris
  - Kubeconfig Service

To preserve Git history, complete Compass history is pushed to the Kyma Control Plane repository. All changes are made on top of the Compass history, including Application Connectivity components removal or code imports change.

### Steps

The following steps have to be executed, to complete Compass project split:

- Create new repository `control-plane` under `kyma-project` GitHub organization

  - follow the process described in kyma-project/community
  - configure repository (labels, branch protection etc.)

- Split Chart

  - Step 1: do it in the same repository
  - Step 2: move KCP chart to new repository
  - Make Compass-KCP communication external
  - Modify components if needed
  - Remove KCP components on Compass repo, and Compass components on KCP repo

- Create Kyma Control Plane Installer

  - base on compass installer (copy & modify needed files)
  - Prepare Prow pipeline to build KCP installer
  - Prepare installation CR for Kyma, Compass and KCP
    - Use separate namespace

- Define fixed Compass release version in the Kyma Control Plane repo

- Copy `KYMA_VERSION` file to Control Plane repo

- Modify run.sh script to install: Kyma (from `KYMA_VERSION`), Compass (from `COMPASS_VERSION`), KCP straight from the repo
- Prepare KCP-integration Prow pipeline for new KCP repo (base on existing compass-integration job)
- Prepare KCP-GKE-integration Prow pipeline for new KCP repo (base on existing compass-gke-integration job on Compass)

- Duplicate GKE-compass-provisioner tests Prow pipeline and run it on KCP repo

- Prepare Prow jobs for building KCP components

- Prepare KCP development artifacts Prowjob

  - Base on Compass artifacts jobs

- Prepare installation CR for KCP

  - Use separate namespace

- Split documentation

  - Discuss how the docs should look on the website (which node)

- Configure Test Grid to display KCP dashboard

- Modify pipelines for the internal environments:

  - Watch KCP repo for changes
  - Install Compass from fixed `COMPASS_VERSION` from KCP repo
  - Modify from where we get Kyma version to install: Install Kyma from fixed `KYMA_VERSION` from KCP repo
  - Install KCP installer as a new step

### Migration of existing Compass installations

Migration of the existing Compass installation doesn't require any custom operations.

Once a pipeline is modified to install Kyma Control Plane components from `control-plane` repository on top of existing Compass installation, the transition period begins. In the transition period, there will be duplicated KCP components on the cluster. They will be deployed in two different namespaces: `compass-system` and `kcp-system`. The transition period ends once the Compass chart is modified and all KCP components are deleted from the Compass chart.

### Cleanup

Once Compass components are migrated to a separate internal environment:

- Remove Compass installation step from internal environment pipelines
