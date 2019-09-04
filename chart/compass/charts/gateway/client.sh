#!/usr/bin/env bash
# https://github.com/kyma-incubator/examples/tree/master/ory-hydra/scenarios/client-credentials

#k apply -f templates/oauth-client.yaml
#k get secrets -n default sample-client -oyaml

# k apply -f templates/hydra-virtualservice-patch.yaml

#export CLIENT_ID= decode from secret
#export CLIENT_SECRET= decode from secret
export DOMAIN=kyma.local

#curl -ik -X POST "https://oauth2-admin.$DOMAIN/clients" -d '{"grant_types":["client_credentials"], "client_id":"'$CLIENT_ID'", "client_secret":"'$CLIENT_SECRET'", "scope":"read write scope-a scope-b"}'

export ENCODED_CREDENTIALS=$(echo -n "$CLIENT_ID:$CLIENT_SECRET" | base64)
curl -ik -X POST "https://oauth2.kyma.local/oauth2/token" -H "Authorization: Basic $ENCODED_CREDENTIALS" -F "grant_type=client_credentials" -F "scope=scope-a scope-b"

#ACCESS_TOKEN= get from curl response
curl -ik https://compass-gateway.kyma.local/healthz -H "Authorization: Bearer ${ACCESS_TOKEN}"