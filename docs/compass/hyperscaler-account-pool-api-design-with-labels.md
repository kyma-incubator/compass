Here's a proposal related to the prior [Hyperscaler Account Pool API design](hyperscaler-account-pool-api-design.md) proposal.
This new proposal suggests a lightweight approach using labels on Kubernetes secrets containing hyperscaler account credentials to realize
the required "pool of credentials". 

In the Gardener use case the credentials secrets reside on the Gardener Seed cluster. In the Bring Your Own license
use case the credentials reside on the Compass cluster. Creating and labeling these secrets is part of Hyperscaler
account management, and happens ahead of the provisioning process described here. This introduces a dependency that 
the labels required here are created on the secrets in the expected format. 

A note about ***tenant-name*** and ***account-name***. This document uses tenant-name to represent an end user tenant, as that fits with the
language used in the Kyma story: https://github.com/kyma-incubator/compass/issues/439

Elsewhere this appears to be called Account or AccountName, but these appear to be synonymous with tenant-name as used here.

#### Leverage existing process using Kubernetes credentials secrets

Create secrets with credentials needed for Terraform/Hydroform (as already happens today). 
Add labels to these secrets to manage "pools" of credentials for use by the provisioning process. 

The ***hyperscaler-type*** label is always added during secret creation. The ***tenant-name*** is added at secret creation
time for Bring Your Own license hyperscaler accounts. Otherwise the tenant-name label is 
omitted for new account secrets indicating they are "in the pool" ready for claiming and use.


#### Kubernetes Example Secret Yaml:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: ${SECRET_NAME}
  labels:
    # tenant-name is omitted for new, not yet claimed account credentials
    tenant-name: ${TENANT_NAME}
    hyperscaler-type: ${HYPERSCALER_TYPE}
```

Where hyperscaler-type is one of: GCP, Azure, AWS, etc

In the Gardener provisioner use-case, the secrets are stored in the Gardener Kubernetes cluster. In the Bring Your Own
credentials use-case the credentials are stored in the Compass Kubernetes cluster. When a new cluster is provisioned,
the Provisioner queries for a secret based on ***tenant-name*** and ***hyperscaler-type***. If a secret is found, that
secret is used to provide credentials to the provisioning process (in the same manner as happens today). If a matching
secret is not found, query again for an un-assigned secret (***tenant-name*** label absent). Add the ***tenant-name*** 
label to claim the account and use the credentials to provision (in the same manner as happens today).

####Example process flow
The example uses kubectl commands to illustrate the logical flow. The implementation will use Go code, but
the flow should be the same.

#####1. Get existing account (always try this first):

```kubectl get secret -l tenant-name=${TENANT_NAME}, hyperscaler-type=${HYPERSCALER_TYPE}```

#####2. If secret found, use that secret for Terraform/Hydroform credentials. ***Done***.


#####3. If not found, claim new credentials from the pool. First query for list of unclaimed accounts:

```kubectl get secrets -l !tenant-name, hyperscaler-type=${HYPERSCALER_TYPE}```


#####4. Pick one of the returned secrets and label it to claim, repeating the type and tenant criteria for optimistic locking (Go code might use ResourceVersion for optimistic locking):

```kubectl label secret ${SECRET_NAME} -l !tenant-name, hyperscaler-type=${HYPERSCALER_TYPE} tenant-name=${TENANT_NAME}```


#####5. If step 4 fails with not found, perhaps a concurrent request swiped the account, retry step 3


#####6. Once labeling succeeds, go to step 1

#### Go Code Design and Background

The existing code in [hydroform/configuration/builder.go](../../components/provisioner/internal/hydroform/configuration/builder.go)
uses a client to fetch Kuberenets secrets ``` v1.SecretInterface``` and the secret name, from ```gqlschema.ProvisionRuntimeInput.Credentials.SecretName```
to enable Hydroform/Terraform to communicate with the Kubernetes to be provisioned. We can leverage that same approach of secret client usage
in the new ```HyperscalerAccountPool``` to find the appropriate secret name.  The secret name no longer makes sense as an argument to the graphQL provisioning API.  An additional API will need to be added to allow an end user to add their hyperscaler credentials to compass.

```go

import (
	"github.com/kyma-incubator/hydroform/types"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// HyperscalerAccountPool represents a collection of credentials used by Hydroform/Terraform to provision clusters.
type HyperscalerAccountPool interface {
	
	// Query the secrets for credentials matching the requested providerType (Azure, GCP, etc) and tenantName.
	// If secret with tenantName not found, get an un-assigned secret and add a new tenant-name label to claim.
	// Return the name of the secret.
	CredentialsSecretName(providerType types.ProviderType, tenantName string, secretsClient v1.SecretInterface) (string, error)
}

```

With the interface suggested here one runtime instance of HyperscalerAccountPool could be used for both Gardener 
provisioning and GCP (or other) cases. Since we pass the Kubernetes v1.SecretInterface in the call, it's up to the caller 
to create a secretsClient (v1.SecretInterface) with a connection to the appropriate Kubernetes instance (Gardener or Compass).

#### Questions
 #####1. Different credential types for different use cases
 This proposal assumes that the content of the credentials secrets may vary by use case. For example, a Gardener secret
 may contain a Kubeconfig, where a GCP secret may contain a ServiceAccountKey. The HyperscalerAccountPool is unaware of
 the content of the secret. The design delegates appropriate use of the secret content to Hydroform/Terraform. It 
 appears this may be the case already, so we get to leverage the existing process here.
 
 #####2. Multiple Hyperscaler Accounts for One Tenant
 If one tenant would like to use more than one account for the same hyperscaler, an additional attribute will need to be added,
 for example, if Tenant1 would like to use GCP accounts AccountA and AccountB. The current proposal assumes a one-to-one relationship 
 between (Tenant + HyperscalerType) to Hyperscaler Account. For the sake of simplicity Hyperscaler Account Name has been
 left out for now, but it could be added if and when needed.
