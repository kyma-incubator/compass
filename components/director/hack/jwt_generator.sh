#!/usr/bin/env bash

function get_internal_tenant(){
    local INTERNAL_TENANT_ID=$(docker exec -i ${POSTGRES_CONTAINER} psql -qtAX -U "${DB_USER}" -h "${DB_HOST}" -p "${DB_PORT}" -d "${DB_NAME}" -c "SELECT id FROM business_tenant_mappings WHERE external_tenant = '3e64ebae-38b5-46a0-b1ed-9ccee153a0ae'")
    echo "$INTERNAL_TENANT_ID"
}

function get_token(){
    local INTERNAL_TENANT_ID
    read -r INTERNAL_TENANT_ID <<< $(get_internal_tenant)

    local HEADER=$(echo "{ \"alg\": \"none\", \"typ\": \"JWT\" }" | base64 | tr '/+' '_-' | tr -d '=')
    local PAYLOAD=$(echo "{ \"scopes\": \"runtime.webhooks:read application.local_tenant_id:write tenant_subscription:write tenant:write fetch-request.auth:read webhooks.auth:read application.auths:read application.webhooks:read application_template:write application_template:read application_template.webhooks:read document.fetch_request:read event_spec.fetch_request:read api_spec.fetch_request:read runtime.auths:read integration_system.auths:read bundle.instance_auths:read bundle.instance_auths:read application:read automatic_scenario_assignment:write automatic_scenario_assignment:read health_checks:read application:write runtime:write label_definition:write label_definition:read runtime:read tenant:read formation:read formation:write internal_visibility:read formation_template:read formation_template:write certificate_subject_mapping:read certificate_subject_mapping:write\", \"tenant\":\"{\\\"consumerTenant\\\":\\\"$INTERNAL_TENANT_ID\\\",\\\"externalTenant\\\":\\\"3e64ebae-38b5-46a0-b1ed-9ccee153a0ae\\\"}\" }" | base64 | tr '/+' '_-' | tr -d '=')
    echo "$HEADER.$PAYLOAD."
}

POSTGRES_CONTAINER="${POSTGRES_CONTAINER:-"test-postgres"}"
POSTGRES_VERSION="${POSTGRES_VERSION:-"11"}"
DB_USER="${DB_USER:-"postgres"}"
DB_PWD="${DB_PWD:-"pgsql@12345"}"
DB_NAME="${DB_NAME:-"compass"}"
DB_PORT="${DB_PORT:-"5432"}"
DB_HOST="${DB_HOST:-"127.0.0.1"}"

read -r INTERNAL_TENANT_ID <<< "$(get_internal_tenant)"
echo "Internal Tenant ID for default tenant from dump:"
echo -E "${INTERNAL_TENANT_ID}"

read -r JWT_TOKEN <<< "$(get_token)"
echo "Use the following JWT token when requesting Director as default tenant:"
echo -E "${JWT_TOKEN}"
