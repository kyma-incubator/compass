BEGIN;
CREATE TABLE destinations
(
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    name varchar(256) NOT NULL,
    type varchar(256) NOT NULL,
    url varchar(256) NOT NULL,
    authentication varchar(256) NOT NULL,
    bundle_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    revision uuid NOT NULL,
    CONSTRAINT bundle_id_fk FOREIGN KEY (bundle_id)
        REFERENCES bundles (id)
        ON UPDATE NO ACTION
        ON DELETE CASCADE,
    CONSTRAINT tenant_id_fk FOREIGN KEY (tenant_id)
        REFERENCES business_tenant_mappings (id)
        ON UPDATE NO ACTION
        ON DELETE CASCADE
);

ALTER TABLE destinations
    ADD CONSTRAINT destinations_tenant_name_uniqueness UNIQUE(name, tenant_id);

CREATE OR REPLACE VIEW tenants_destinations(tenant_id, id, name, type, url, authentication, bundle_id, revision, sensitive_data) AS
SELECT DISTINCT dst.tenant_id,
                dests.id,
                dests.name,
                dests.type,
                dests.url,
                dests.authentication,
                dests.bundle_id,
                dests.revision,
                '__sensitive_data__' || dests.name || '__sensitive_data__'
FROM destinations dests
         JOIN (SELECT d.id,
                      d.tenant_id::text AS tenant_id
               FROM destinations d
               UNION ALL
               SELECT apps_subaccounts_func.id,
                      apps_subaccounts_func.tenant_id
               FROM apps_subaccounts_func() apps_subaccounts_func(id, tenant_id)) dst ON dests.id = dst.id;
COMMIT;
