BEGIN;

DROP TABLE technical_integrations;

DROP VIEW app_templates_business_tenants;
DROP VIEW apps_business_tenants;
DROP VIEW tenant_to_tenant_technical_integrations;
DROP VIEW oauth2_client_credentials_details;

COMMIT;
