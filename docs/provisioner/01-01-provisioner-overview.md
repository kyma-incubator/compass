---
title: Overview
type: Overview
---

The Runtime Provisioner is a Compass component responsible for provisioning, installing, and deprovisioning clusters with Kyma (Kyma Runtimes).
It allows you to provision the clusters in two ways:
- [directly on Google Cloud Platform (GCP)](08-01-provisioning-gcp.md)
- [through Gardener](08-02-provisioning-gardener.md) on:
    * Google Cloud Platform (GCP)
    * Azure
    * Amazon Web Services (AWS).  
Follow the links to access the respective tutorials. 
    
You can access the Runtime Provisioner in two ways:

- by making a call from inside the cluster with the Provisioner (e.g. from another pod)

    ```bash
    kubectl exec {COMMAND_TO_MAKE_A_CALL_FROM_INSIDE_THE_POD}
    ```

- by forwarding the port that the Application is listening on

    ```bash
    kubectl -n compass-system port-forward svc/compass-provisioner 3000:3000
    ```
    <!--- alternatively: by forwarding the port that the GraphQL Server is listening on --->
    

The operations of provisioning and deprovisioning are asynchronous. 