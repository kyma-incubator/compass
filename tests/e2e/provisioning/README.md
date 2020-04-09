# Kyma Runtime end-to-end provisioning test

## Overview

Kyma Runtime end-to-end provisioning test checks if [Runtime provisioning](https://github.com/kyma-incubator/compass/blob/master/docs/kyma-environment-broker/02-01-architecture.md) works as expected. The test is based on the Kyma Environment Broker (KEB), Runtime Provisioner and Director implementation. External dependencies relevant for this scenario are mocked. 

The test is executed on a dev cluster. It is executed after every merge to the `kyma` repository that changes the `compass` chart.

## Prerequisites

To run this test, you must have the following Secrets inside your cluster:
- Gardener Secret per provider
- Service Manager Secret

You must also have the Kyma Environment Broker [configured](https://github.com/kyma-incubator/compass/tree/master/components/kyma-environment-broker#configuration) to use these Secrets in order to successfully create a Runtime.

## Details

The provisioning end-to-end test contains a broker client implementation which mocks Registry. It is an external dependency that calls the broker in the regular scenario. The test is divided into two phases:

1. Provisioning:
    
    a. Send a call to KEB to provision a Runtime. KEB creates an operation and sends a request to Runtime Provisioner. Wait until the operation is successful. It takes about 30 minutes on GCP and a few hours on Azure. You can configure the timeout using the environment variable. 
    
    b. Creates a config map with `instanceId` specified.
    
    c. Fetch the DashboardURL from KEB. To do so, the Runtime must be successfully provisioned and registered in the Director.
    
    d. Updates config map with `dashboardUrl` field.
    
    e. Creates a secret with a kubeconfig of the provisioned runtime.
    
    f. Ensure that the DashboardURL redirects you to the UUA login page. It means that the Kyma Runtime is accessible.

2. Cleanup

    The cleanup logic is executed at the end of the e2e test or if the provisioning phase failed. It consists of the following steps:
    
    a. Get `instanceId` from the config map.
    
    b. Remove the test's secret and config map.
    
    c. Fetch the Runtime kubeconfig from Provisioner and use it to clean resources which block the cluster from deprovisioning.
    
    d. Send a request to deprovision the Runtime to KEB. The request is passed to Runtime Provisioner which deprovisions the Runtime.
    
    e. Wait until the deprovisioning is successful. It takes about 20 minutes to complete. You can configure the timeout using the environment variable.

Between the above steps you can execute your own test directly on the provisioned runtime, using a kubeconfig fetched from the Provisioner.

## Configuration

You can configure the test execution by using the following environment variables:

| Name | Description | Default value |
|-----|---------|:--------:|
| **APP_BROKER_URL** | Specifies the KEB URL. | None |
| **APP_PROVISION_TIMEOUT** | Specifies a timeout for the provisioning operation to succeed. | `3h` |
| **APP_DEPROVISION_TIMEOUT** | Specifies a timeout for the deprovisioning operation to succeed. | `1h` |
| **APP_BROKER_PROVISION_GCP** | Specifies if a Runtime cluster is hosted on GCP. If set to `false`, it provisions on Azure. | `true` |
| **APP_BROKER_AUTH_USERNAME** | Specifies the username for the basic authentication in KEB. | `broker` |
| **APP_BROKER_AUTH_PASSWORD** | Specifies the password for the basic authentication in KEB. | None |
| **APP_RUNTIME_PROVISIONER_URL** | Specifies the Provisioner URL. | None |
| **APP_RUNTIME_UUA_INSTANCE_NAME** | Specifies the name of the UUA instance which is provisioned in the Runtime. | `uua-issuer` |
| **APP_RUNTIME_UUA_INSTANCE_NAMESPACE** | Specifies the Namespace of the UUA instance which is provisioned in the Runtime. | `kyma-system` |
| **APP_TENANT_ID** | Specifies TenantID which is used in the test. | None |
| **APP_DIRECTOR_URL** | Specifies the Director URL. | `http://compass-director.compass-system.svc.cluster.local:3000/graphql` |
| **APP_DIRECTOR_NAMESPACE** | Specifies the Director Namespace. | `compass-system` |
| **APP_DIRECTOR_OAUTH_CREDENTIALS_SECRET_NAME** | Specifies the name of the Secret created by the Integration System. | `compass-kyma-environment-broker-credentials` |
| **APP_SKIP_CERT_VERIFICATION** | Specifies whether TLS checks the presented certificates. | `false` |
| **APP_DUMMY_TEST** | Specifies if test should success without any action. | `false` |
| **APP_CLEANUP_PHASE** | Specifies if test is executed in the cleanup phase. | `false` |
| **APP_CONFIG_NAME** | Specifies the name of the config map and secret, created in the test. | `false` |
| **APP_DEPLOY_NAME** | Specifies the namespace of the config map and secret, created in the test. | `false` |
