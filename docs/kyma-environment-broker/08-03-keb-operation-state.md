---
title: Check Kyma Environment Broker operation status
type: Tutorials
---

This tutorial shows how to check the Kyma Environment Broker operation status for the provisioning and deprovisioning operations.

## Steps

1. Export the operation ID that you obtained during [provisioning](./08-01-provisioning-kyma-environment.md) or [deprovisioning](./08-02-deprovisioning-kyma-environment.md) as an environment variable:

```bash
export OPERATION_ID={OBTAINED_OPERATION_ID}
```

> **NOTE:** Ensure that BROKER_URL and INSTANCE_ID environment variables are exported as well before you proceed.

2. Make a call to the Kyma Environment Broker with a proper **Authorization** [request header](./03-05-authorization.md) to verify that provisioning/deprovisioning succeeded.

```bash
curl --request GET "https://$BROKER_URL/v2/service_instances/$INSTANCE_ID/last_operation?operation=$OPERATION_ID&service_id=47c9dcbf-ff30-448e-ab36-d3bad66ba281&plan_id=4deee563-e5ec-4731-b9b1-53b42d855f0c" \
--header 'X-Broker-API-Version: 2.13' \
--header "$AUTHORIZATION_HEADER"
```

A successful call returns the operation status and description:

```json
{
    "state": "succeeded",
    "description": "Operation created : Operation succeeded."
}
```
