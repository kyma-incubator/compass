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
                    WHERE external_tenant = l.value::varchar
                    UNION ALL
                    SELECT t2.id, t2.parent, t.app_template_id
                    FROM business_tenant_mappings t2
                             INNER JOIN tenants t on t2.id = t.parent)
SELECT DISTINCT tenants.app_template_id as id,
                tenants.id              as tenant_id
FROM tenants;

CREATE OR REPLACE VIEW apps_business_tenants(id, tenant_id)
AS
SELECT id, tenant_id
FROM tenant_applications
WHERE owner = true;

CREATE OR REPLACE VIEW tenant_to_tenant_technical_integrations(id, sender, receiver, name, type)
AS
SELECT id, sender, receiver, name, type
FROM technical_integrations
WHERE sender_type = 'APPLICATION'
  AND receiver_type = 'APPLICATION';

CREATE OR REPLACE VIEW oauth2_client_credentials_details(tech_int_id, token_service_url, client_id)
AS
SELECT id                        AS tech_int_id,
       details.token_service_url AS token_service_url,
       details.client_id         AS client_id
FROM technical_integrations t,
     jsonb_to_recordset(t.value) AS details(token_service_url TEXT, client_id TEXT)
WHERE sender_type = 'APPLICATION'
  AND receiver_type = 'APPLICATION'
  AND type = 'oauth2ClientCredentials';

COMMIT;
