#!/usr/bin/env bash

INTERNAL_TENANT_ID=$(docker exec -i ${POSTGRES_CONTAINER} psql -qtAX -U "${DB_USER}" -h "${DB_HOST}" -p "${DB_PORT}" -d "${DB_NAME}" -c "SELECT id FROM business_tenant_mappings WHERE external_tenant = '3e64ebae-38b5-46a0-b1ed-9ccee153a0ae'")
echo -e "${GREEN}Internal Tenant ID for default tenant from dump: $INTERNAL_TENANT_ID${NC}"


HEADER=$(echo "{ \"alg\": \"none\", \"typ\": \"JWT\" }" | base64 | tr '/+' '_-' | tr -d '=')
PAYLOAD=$(echo "{ \"scopes\": \"application:read automatic_scenario_assignment:write automatic_scenario_assignment:read health_checks:read application:write runtime:write label_definition:write label_definition:read runtime:read tenant:read formation:write\", \"tenant\": \"$INTERNAL_TENANT_ID\" }" | base64 | tr '/+' '_-' | tr -d '=')
JWT_TOKEN="$HEADER.$PAYLOAD."

echo -e "${GREEN}Use the following JWT token when requesting Director as default tenant:${NC}"
echo $JWT_TOKEN