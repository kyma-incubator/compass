## Overview

The Kyma Runtime e2e provisioning test checks if Runtime provisioning based on real implementation of the Kyma Environment Broker(KEB), Provisioner and Director.
External dependencies which takes part in this scenario are faked.

This test is executed on the Dev cluster to ensure that provisioning scenario works as expected. It is executed after every merge to Kyma repository that changes the Compass chart.

## Prerequisites

To run this test you must have the following secrets inside your cluster:
- Secret for Gardener (per provider)
- Secret for Service Manager

The Kyma Environment Broker must be configured to use them in order to successfully create a cluster.

## Details

The provisioning e2e test contains a broker client implementation which fakes the external dependency - Registry, which will call the broker in the normal scenario.
The test scenario consist of the following steps:

- Send a provision call to KEB - it creates an operation which we will monitor and sends request to Provisioner
- Wait until the operation is succeeded - the timeout can be configured using the environment variable, it takes about 30min on GCP and a few hours on Azure
- Fetch DashboardURL from the KEB - to achieve that Runtime must be provisioned and registered in Director successfully
- Assert that DashboardURL redirects user to the UUA login page - that steps indicates that Kyma Runtime is accessible by a normal user

In case of any error or at the end of the test, the cleanup logic is executed. It contains following steps:
- Fetch Runtime's kubeconfig from Provisioner and use it to clean resources which would block the cluster from deprovisioning. For now we only must to remove UUA service instance
- Send a deprovision request to KEB - it sends it further to Provisioner which deprovision the Runtime

## Future

In the future we need to implement waiting for the deprovision action in order to be sure if the Runtime was successfully deprovisioned.

That'd be the best if we could use this test before merging our changes to master, but that requires further investigations on how to achieve that.

## Configuration

You can configure the test execution by using the following environment variables:

| Name | Description | Default value |
|-----|---------|:--------:|
| **APP_BROKER_URL** | Specifies the KEB URL. | None |
| **APP_PROVISION_TIMEOUT** | Specifies a timeout on waiting for provision operation to succeed. | `3h` |
| **APP_DEPROVISION_TIMEOUT** | Specifies a timeout on waiting for deprovision operation to succeed. | `1h` |
| **APP_BROKER_PROVISION_GCP** | Specifies if Runtime cluster is host on GCP, if set to false it provisions on Azure. | `true` |
| **APP_BROKER_AUTH_USERNAME** | Specifies the username for basic auth in KEB. | `broker` |
| **APP_BROKER_AUTH_PASSWORD** | Specifies the password for basic auth in KEB. | None |
| **APP_RUNTIME_PROVISIONER_URL** | Specifies the Provisioner URL. | None |
| **APP_RUNTIME_UUA_INSTANCE_NAME** | Specifies the name of UUA instance which is provisioned on the Runtime. | `uua-issuer` |
| **APP_RUNTIME_UUA_INSTANCE_NAMESPACE** | Specifies the namespace of UUA instance which is provisioned on the Runtime. | `kyma-system` |
| **APP_TENANT_ID** | Specifies the TenantID which is used in the test. | None |
| **APP_DIRECTOR_URL** | Specifies the Director's URL. | `http://compass-director.compass-system.svc.cluster.local:3000/graphql` |
| **APP_DIRECTOR_NAMESPACE** | Specifies the Director's Namespace. | `compass-system` |
| **APP_DIRECTOR_OAUTH_CREDENTIALS_SECRET_NAME** | Specifies the name of the Secret created by the Integration System. | `compass-kyma-environment-broker-credentials` |
| **APP_SKIP_CERT_VERIFICATION** | Specifies whether TLS checks the presented certificates. | `false` |
| **APP_DUMMY_TEST** | Specifies if test should success without any action. | `postgres` |

The above environment variables are defined in the TestDefinition.