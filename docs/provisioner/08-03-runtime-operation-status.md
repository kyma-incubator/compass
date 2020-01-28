---
title: Check Runtime Operation Status
type: Tutorials
---

This tutorial shows how to check the Runtime operation status for the operations of Runtime deprovisioning. 

## Steps

> **NOTE:** To access the Runtime Provisioner, forward the port on which the GraphQL Server is listening.

Make a call to the Runtime Provisioner to verify that deprovisioning succeeded. Pass the ID of the operation as `id`.

```graphql
query { 
  runtimeOperationStatus(id: "e9c9ed2d-2a3c-4802-a9b9-16d599dafd25") { 
    operation 
    state 
    message 
    runtimeID 
  }
}
```

A successful call returns a response which includes the status of the deprovisioning operation (`state`) and the id of the deprovisioned Runtime (`runtimeID`):

```graphql
{
  "data": {
    "runtimeOperationStatus": {
      "operation": "Deprovision",
      "state": "Succeeded",
      "message": "Operation succeeded.",
      "runtimeID": "309051b6-0bac-44c8-8bae-3fc59c12bb5c"
    }
  }
}
```

The `Succeeded` status means that the deprovisioning was successful and the cluster was deleted.

If you get the `InProgress` status, it means that the deprovisioning has not yet finished. In that case, wait a few moments and check the status again.