---
title: Deprovision Kyma environment through Kyma Environment Broker
type: Tutorials
---

This tutorial shows how to deprovision Kyma environment on Azure.

## Steps

1. Ensure these environment variables are exported:
```bash
export BROKER_URL={KYMA_ENVIRONMENT_BROKER_URL}
export INSTANCE_ID={INSTANCE_ID_FROM_PROVISIONING_CALL}
```

2. Obtain Authorization credentials as described in [Authorization section](./03-05-authorization.md). Export this variable based on chosen authorization method:

```bash
export AUTHORIZATION_HEADER="Authorization: Basic $ENCODED_CREDENTIALS"
```

3. Make a call to the Kyma Environment Broker to delete a runtime on Azure.

```bash
curl  --request DELETE "https://$BROKER_URL/v2/service_instances/$INSTANCE_ID?accepts_incomplete=true&service_id=47c9dcbf-ff30-448e-ab36-d3bad66ba281&plan_id=4deee563-e5ec-4731-b9b1-53b42d855f0c" \
--header 'X-Broker-API-Version: 2.13' \
--header "$AUTHORIZATION_HEADER"
```

A successful call returns the operation ID:

```json
{
    "operation":"8a7bfd9b-f2f5-43d1-bb67-177d2434053c"
}
```

4. Check operation state as described in [this tutorial](./08-03-keb-operation-state.md).