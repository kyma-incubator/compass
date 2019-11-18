---
title: Deprovision clusters
type: Tutorials
---

This tutorial shows how to deprovision clusters with Kyma Runtimes (Runtimes).

## Deprovision Kyma Runtime

> **NOTE:** To access the Runtime Provisioner, make a call from another Pod in the cluster containing the Runtime Provisioner or forward the port on which the GraphQL Server is listening.

  To deprovision a Runtime, make a call to the Runtime Provisioner with a mutation like this:  
  
  ```graphql
  mutation { deprovisionRuntime(id: "61d1841b-ccb5-44ed-a9ec-45f70cd1b0d3") }
  ```

  A successful call returns the ID of the deprovisioning operation:

  ```graphql
  {
    "data": {
      "deprovisionRuntime": "c7e6727f-16b5-4748-ac95-197d8f79d094"
    }
  }
  ```

  The operation of deprovisioning is asynchronous. Use the deprovisioning operation ID (`deprovisionRuntime`) to [check the Runtime Operation Status](08-03-runtime-operation-status.md) and verify that the deprovisioning was successful. Use the Runtime ID (`id`) to [check the Runtime Status](08-04-runtime-status.md). 