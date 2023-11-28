BEGIN;

-- TODO: Add check constraints for enum fields - sender_type, receiver_type, type
CREATE TABLE technical_integrations
(
    id            uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    sender        UUID         NOT NULL,
    sender_type   VARCHAR(256) NOT NULL,
    receiver      UUID         NOT NULL,
    receiver_type VARCHAR(256) NOT NULL,
    name          VARCHAR(256),
    type          VARCHAR(256) NOT NULL,
    value         JSONB
);

CREATE INDEX idx_technical_integrations_sender
    ON public.technical_integrations (sender);

CREATE INDEX idx_technical_integrations_receiver
    ON public.technical_integrations (receiver);

ALTER TABLE technical_integrations
    ADD CONSTRAINT sender_receiver_uniq UNIQUE (sender, receiver);


CREATE OR REPLACE VIEW app_templates_business_tenants(id, tenant_id)
AS
WITH RECURSIVE tenants AS
                   (SELECT t1.id, t1.parent, l.app_template_id
                    FROM business_tenant_mappings t1
                             JOIN labels l ON key = 'global_subaccount_id' AND app_template_id is not null
                    WHERE external_tenant = value ->> 0
                    UNION ALL
                    SELECT t2.id, t2.parent, t.app_template_id
                    FROM business_tenant_mappings t2
                             INNER JOIN tenants t on t2.id = t.parent)
SELECT DISTINCT tenants.app_template_id as id,
                tenants.id              as tenant_id
FROM tenants;

CREATE OR REPLACE VIEW apps_business_tenants(id, tenant_id)
AS
SELECT DISTINCT id, tenant_id
FROM tenant_applications
WHERE owner = true;

CREATE OR REPLACE VIEW business_tenants(id, external_tenant, external_name, type, license_type, region, subdomain, parent_external_tenant)
AS
SELECT DISTINCT btm.id ,btm.external_tenant, btm.external_name, btm.type, ll.value ->> 0, lr.value ->> 0, ls.value ->> 0, parents.external_tenant
FROM business_tenant_mappings btm
    LEFT JOIN business_tenant_mappings parents ON btm.parent = parents.id
    LEFT JOIN labels ll ON btm.id = ll.tenant_id AND ll.key = 'licensetype'
    LEFT JOIN labels lr ON btm.id = lr.tenant_id AND lr.key = 'region'
    LEFT JOIN labels ls ON btm.id = ls.tenant_id AND ls.key = 'subdomain';

CREATE OR REPLACE VIEW tenant_to_tenant_technical_integrations(id, tenant_id, sender, receiver, name, type)
AS
SELECT DISTINCT ti.id, ta.tenant_id, ti.sender, ti.receiver, ti.name, ti.type
FROM technical_integrations ti
         JOIN tenant_applications ta ON ta.id = ti.sender OR ta.id = ti.receiver
WHERE sender_type = 'APPLICATION'
  AND receiver_type = 'APPLICATION';

CREATE OR REPLACE VIEW oauth2_client_credentials_details(tech_int_id, token_service_url, client_id)
AS
SELECT DISTINCT id                        AS tech_int_id,
                t.value->>'tokenServiceUrl' AS token_service_url,
                t.value->>'clientId'         AS client_id
FROM technical_integrations t
WHERE sender_type = 'APPLICATION'
  AND receiver_type = 'APPLICATION'
  AND type = 'oauth2ClientCredentials';

COMMIT;

--- Dummy Data

INSERT INTO technical_integrations (id, sender, sender_type, receiver, receiver_type, name, type, value)
VALUES ('cba94a64-c55b-4caa-8b46-df13f09261e9', '775115ca-12fc-43f5-90e5-7733abb44792', 'APPLICATION', '7d2ad5ef-8d41-47db-a4e7-f7f35d60dc5a', 'APPLICATION', NULL, 'oauth2ClientCredentials', '{"tokenServiceUrl":"https://consumer-y7sc8ue0.authentication.sap.hana.ondemand.com/oauth/token","clientId":"sb-revenue-cloud!t8509"}'::jsonb)
