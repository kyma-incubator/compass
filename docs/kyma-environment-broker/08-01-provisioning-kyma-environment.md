---
title: Provision Kyma Runtime through Kyma Environment Broker
type: Tutorials
---

This tutorial shows how to provision Kyma Runtime on Azure.

## Prerequisites

- Compass with:
  * configured Runtime Provisioner and Azure provider [set up](../provisioner/08-02-provisioning-gardener.md)
  * configured Kyma Environment Broker and chosen [overrides](03-06-runtime-overrides.md) set up

## Steps

1. Export these values as environment variables:
```bash
export BROKER_URL={KYMA_ENVIRONMENT_BROKER_URL}
export INSTANCE_ID={INSTANCE_ID}
export GLOBAL_ACCOUNT_ID={GLOBAL_ACCOUNT_ID}
export NAME={RUNTIME_NAME}
```

> **NOTE:** INSTANCE_ID and NAME must be unique. It is recommended to use UUID as an INSTANCE_ID.

2. Get [authorization credentials](./03-05-authorization.md). Export this variable based on the chosen authorization method:

```bash
export AUTHORIZATION_HEADER="Authorization: {Basic OR Bearer} $ENCODED_CREDENTIALS"
```

3. Make a call to the Kyma Environment Broker to create a Runtime on Azure. Find the list of possible request parameters [here](./03-01-service-description.md).

```bash
curl --request PUT "https://$BROKER_URL/v2/service_instances/$INSTANCE_ID?accepts_incomplete=true" \
--header 'X-Broker-API-Version: 2.14' \
--header 'Content-Type: application/json' \
--header "$AUTHORIZATION_HEADER" \
--header 'Content-Type: application/json' \
--data-raw "{
    \"service_id\": \"47c9dcbf-ff30-448e-ab36-d3bad66ba281\",
    \"plan_id\": \"4deee563-e5ec-4731-b9b1-53b42d855f0c\",
    \"context\": {
        \"globalaccount_id\": \"$GLOBAL_ACCOUNT_ID\"
    },
    \"parameters\": {
        \"name\": \"$NAME\"
    }
}"
```

A successful call returns the operation ID:

```json
{
    "operation":"8a7bfd9b-f2f5-43d1-bb67-177d2434053c"
}
```

4. Check the operation status as described [here](./08-03-keb-operation-state.md).
