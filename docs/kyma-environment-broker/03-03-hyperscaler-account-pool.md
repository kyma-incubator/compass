# Hyperscaler Account Pool
​
When provisioning clusters through Gardener with the Runtime Provisioner, the Kyma Environment Broker (KEB) requires a hyperscaler (GCP, Azure, AWS, etc.) account/subscription, but setting it up is outside of the Runtime Provisioner's scope. In order for the user to be able to easily provision Kyma clusters, Hyperscaler Account Pool was created. 

It contains credentials for the hyperscaler accounts that have been set up in advance, stored in Kubernetes Secrets. The credentials are stored separately for each provider and each tenant. The content of the credentials Secrets may vary for different use cases. The Secrets are labeled with the **hyperscaler-type** and **tenant-name** labels to manage pools of credentials for use by the provisioning process. This way, the in-use and unassigned credentials (i.e. credentials available for use) are tracked. Only the **hyperscaler-type** label is added during Secret creation, and the **tenant-name** label is added when the account respective for a given Secret is claimed. The Hyperscaler Account Pool is unaware of the Secrets' content.

The Secrets are stored in the Gardener Kubernetes cluster. They are available within a given Gardener project specified in the KEB and Runtime Provisioner configuration. This configuration uses a kubeconfig that gives KEB and the Runtime Provisioner access to the mentioned Gardener cluster, which enables access to those Secrets. 

When a new cluster is provisioned, the Runtime Provisioner queries for a Secret based on the **tenant-name** and **hyperscaler-type** labels. 
If a Secret is found, the Runtime Provisioner uses the credentials stored in this Secret. If a matching Secret is not found, the user queries again for an unassigned Secret (i.e. a Secret without the **tenant-name** label) and adds the **tenant-name** label to claim the account and use the credentials for provisioning. 

One tenant can use only one account per given hyperscaler type.

This is an example Kubernetes Secret storing hyperscaler credentials:
​
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