---
title: Overview
type: Overview
---

The Runtime Provisioner is a Compass component responsible for provisioning, installing, and deprovisioning clusters with Kyma (Kyma Runtimes). 

> **NOTE:** Kyma installation is not implemented yet. 

It is powered by [Hydroform](https://github.com/kyma-incubator/hydroform) and it allows you to provision the clusters in two ways:
- [directly on Google Cloud Platform (GCP)](08-01-provisioning-gcp.md)
- [through Gardener](08-02-provisioning-gardener.md) on:
    * Google Cloud Platform (GCP)
    * Microsoft Azure
    * Amazon Web Services (AWS).
    
Note that the operations of provisioning and deprovisioning are asynchronous. They return the operation ID, which you can use to [check the Runtime Operation Status](08-03-runtime-operation-status.md).

<!--- You can also use it to [clean up Runtime data](08-06-clean-up-runtime-data.md) when a cluster can no longer be used, but cannot be deprovisioned either. For example, when your cluster dies, or when the operation of deprovisioning has failed. --->

The Runtime Provisioner also allows you to [clean up Runtime data](08-06-clean-up-runtime-data.md). This operation removes all the data concerning a given Runtime from the database and frees up the Runtime ID for reuse. It is useful when your cluster has died or when the operation of deprovisioning has failed.

Follow the links to access the respective tutorials. 
    
To access the Runtime Provisioner, forward the port that the GraphQL Server is listening on:

```bash
kubectl -n compass-system port-forward svc/compass-provisioner 3000:3000
```

