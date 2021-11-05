BEGIN;

CREATE TABLE tenant_applications
(
    tenant_id   uuid NOT NULL,
    id uuid NOT NULL,
    owner       bool,

    FOREIGN KEY (id) REFERENCES applications (id) ON DELETE CASCADE,
    FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings (id) ON DELETE CASCADE,
    PRIMARY KEY (tenant_id, id)
);

CREATE TABLE tenant_runtimes
(
    tenant_id   uuid NOT NULL,
    id uuid NOT NULL,
    owner       bool,

    FOREIGN KEY (id) REFERENCES runtimes (id) ON DELETE CASCADE,
    FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings (id) ON DELETE CASCADE,
    PRIMARY KEY (tenant_id, id)
);

CREATE OR REPLACE FUNCTION on_insert_new_parent() RETURNS TRIGGER AS
$$
DECLARE
    tenant_table TEXT;
BEGIN
    IF NEW.parent IS NOT NULL
    THEN
        FOREACH tenant_table IN ARRAY TG_ARGV -- for each top level resource
            LOOP
                -- mark him as owner to everything that is owned by its child
                EXECUTE format('INSERT INTO %I (SELECT %L, id, owner FROM %I WHERE owner = true AND tenant_id = %L) ON CONFLICT DO NOTHING;', tenant_table, NEW.parent, tenant_table, NEW.id);
            END LOOP;
    END IF;
    RETURN NULL;
END
$$ LANGUAGE plpgsql;

CREATE TRIGGER insert_parent AFTER INSERT OR UPDATE ON business_tenant_mappings FOR EACH ROW EXECUTE PROCEDURE on_insert_new_parent('tenant_applications', 'tenant_runtimes');

CREATE OR REPLACE FUNCTION delete_resource() RETURNS TRIGGER AS
$$
DECLARE
    resource_table TEXT;
    count INTEGER;
BEGIN
    resource_table := TG_ARGV[0];
    EXECUTE format('SELECT COUNT(1) FROM %I WHERE id = %L AND owner = true', TG_TABLE_NAME, OLD.id) INTO count;
    IF count = 0 THEN
        EXECUTE format('DELETE FROM %I WHERE id = %L;', resource_table, OLD.id);
    END IF;
    RETURN NULL;
END
$$ LANGUAGE plpgsql;

CREATE TRIGGER delete_application_resource AFTER DELETE ON tenant_applications FOR EACH ROW EXECUTE PROCEDURE delete_resource('applications');
CREATE TRIGGER delete_runtime_resource AFTER DELETE ON tenant_runtimes FOR EACH ROW EXECUTE PROCEDURE delete_resource('runtimes');

CREATE OR REPLACE FUNCTION insert_parent_chain() RETURNS TRIGGER AS
$$
DECLARE
    parent_id uuid;
BEGIN
    SELECT parent INTO parent_id FROM business_tenant_mappings WHERE id = NEW.tenant_id;
    IF (parent_id IS NOT NULL) THEN
        EXECUTE format('INSERT INTO %I VALUES (%L, %L, true) ON CONFLICT DO NOTHING', TG_TABLE_NAME, parent_id, NEW.id);
    END IF;
    RETURN NULL;
END
$$ LANGUAGE plpgsql;

CREATE TRIGGER add_app_to_parent_tenants AFTER INSERT ON tenant_applications FOR EACH ROW EXECUTE PROCEDURE insert_parent_chain();
CREATE TRIGGER add_runtime_to_parent_tenants AFTER INSERT ON tenant_runtimes FOR EACH ROW EXECUTE PROCEDURE insert_parent_chain();

UPDATE runtimes
SET tenant_id =
        (SELECT id
         FROM business_tenant_mappings
         WHERE external_tenant = (SELECT value ->> 0 FROM labels l WHERE l.runtime_id = runtimes.id AND key = 'global_subaccount_id'))
WHERE EXISTS((SELECT id
              FROM business_tenant_mappings
              WHERE external_tenant = (SELECT value ->> 0 FROM labels l WHERE l.runtime_id = runtimes.id AND key = 'global_subaccount_id')));

INSERT INTO tenant_applications
SELECT tenant_id, id, true
FROM applications
WHERE tenant_id IS NOT NULL AND id IS NOT NULL;

INSERT INTO tenant_runtimes
SELECT tenant_id, id, true
FROM runtimes
WHERE tenant_id IS NOT NULL AND id IS NOT NULL;

UPDATE documents d SET app_id = (SELECT app_id FROM bundles WHERE id = d.bundle_id);

DROP MATERIALIZED VIEW id_tenant_id_index;
DROP FUNCTION get_id_tenant_id_index();

DROP VIEW tenants_apps;
DROP VIEW tenants_bundles;
DROP VIEW tenants_specifications;
DROP VIEW tenants_apis;
DROP VIEW tenants_events;

DROP FUNCTION IF EXISTS apps_subaccounts_func();
DROP FUNCTION IF EXISTS consumers_provider_for_runtimes_func();
DROP FUNCTION IF EXISTS uuid_or_null();

DROP TABLE api_runtime_auths;

-- Drop Constraints
ALTER TABLE applications DROP CONSTRAINT applications_tenant_constraint;
ALTER TABLE api_definitions DROP  CONSTRAINT  api_definitions_package_id_fk;
ALTER TABLE api_definitions DROP  CONSTRAINT  api_definitions_tenant_constraint;
ALTER TABLE api_definitions DROP  CONSTRAINT  api_definitions_tenant_id_fkey;
ALTER TABLE bundle_instance_auths DROP  CONSTRAINT  package_instance_auths_tenant_id_fkey;
ALTER TABLE bundle_instance_auths DROP  CONSTRAINT  package_instance_auths_tenant_id_fkey1;
ALTER TABLE bundles DROP  CONSTRAINT  packages_tenant_id_fkey;
ALTER TABLE bundles DROP  CONSTRAINT  packages_tenant_id_fkey1;
ALTER TABLE bundle_references DROP  CONSTRAINT  bundle_references_tenant_id_fkey;
ALTER TABLE documents DROP  CONSTRAINT  documents_tenant_id_fkey;
ALTER TABLE documents DROP  CONSTRAINT  documents_tenant_constraint;
ALTER TABLE documents DROP  CONSTRAINT  documents_bundle_id_fk;
ALTER TABLE event_api_definitions DROP  CONSTRAINT  event_api_definitions_package_id_fk;
ALTER TABLE event_api_definitions DROP  CONSTRAINT  event_api_definitions_tenant_constraint;
ALTER TABLE event_api_definitions DROP  CONSTRAINT  event_api_definitions_tenant_id_fkey;
ALTER TABLE fetch_requests DROP  CONSTRAINT  fetch_requests_tenant_constraint;
ALTER TABLE fetch_requests DROP  CONSTRAINT  fetch_requests_tenant_id_fkey2;
ALTER TABLE labels DROP  CONSTRAINT  labels_tenant_id_fkey;
ALTER TABLE labels DROP  CONSTRAINT  labels_tenant_id_fkey1;
ALTER TABLE packages DROP  CONSTRAINT  packages_application_tenant_fk;
ALTER TABLE packages DROP  CONSTRAINT  packages_tenant_id_fkey_cascade;
ALTER TABLE products DROP  CONSTRAINT  products_application_tenant_fk;
ALTER TABLE products DROP  CONSTRAINT  products_tenant_id_fkey_cascade;
ALTER TABLE specifications DROP  CONSTRAINT  specifications_tenant_id_fkey;
ALTER TABLE tombstones DROP  CONSTRAINT  tombstones_application_tenant_fk;
ALTER TABLE tombstones DROP  CONSTRAINT  tombstones_tenant_id_fkey_cascade;
ALTER TABLE vendors DROP  CONSTRAINT  vendors_application_tenant_fk;
ALTER TABLE vendors DROP  CONSTRAINT  vendors_tenant_id_fkey_cascade;
ALTER TABLE webhooks DROP  CONSTRAINT  webhooks_runtime_id_fkey;
ALTER TABLE webhooks DROP  CONSTRAINT  webhooks_tenant_constraint;
ALTER TABLE webhooks DROP  CONSTRAINT  webhooks_tenant_id_fkey;
ALTER TABLE system_auths DROP  CONSTRAINT  system_auths_tenant_id_fkey;
ALTER TABLE system_auths DROP  CONSTRAINT  system_auths_tenant_id_fkey1;

ALTER TABLE runtimes DROP CONSTRAINT runtimes_tenant_constraint;
ALTER TABLE runtime_contexts DROP CONSTRAINT runtime_contexts_tenant_id_fkey;
ALTER TABLE runtime_contexts DROP CONSTRAINT runtime_contexts_tenant_id_fkey1;

-- Add Constraints
ALTER TABLE api_definitions ADD  CONSTRAINT  api_definitions_package_id_fk  FOREIGN KEY (package_id) REFERENCES packages(id);
ALTER TABLE api_definitions ADD  CONSTRAINT  api_definitions_application_id_fk FOREIGN KEY (app_id) REFERENCES applications(id) ON DELETE CASCADE;
ALTER TABLE bundle_instance_auths ADD  CONSTRAINT  bundle_instance_auths_bundle_id_fk FOREIGN KEY (bundle_id) REFERENCES bundles(id) ON DELETE CASCADE;
ALTER TABLE bundle_instance_auths ADD  CONSTRAINT  bundle_instance_auths_tenant_id_fk FOREIGN KEY (tenant_id) REFERENCES business_tenant_mappings(id) ON DELETE CASCADE;
ALTER TABLE bundles ADD  CONSTRAINT   bundles_application_id_fk FOREIGN KEY (app_id) REFERENCES applications(id) ON DELETE CASCADE;
ALTER TABLE documents ADD  CONSTRAINT  documents_application_id_fk FOREIGN KEY (app_id) REFERENCES applications(id) ON DELETE CASCADE;
ALTER TABLE documents ADD  CONSTRAINT  documents_bundle_id_fk FOREIGN KEY (bundle_id) REFERENCES bundles(id) ON DELETE CASCADE;
ALTER TABLE event_api_definitions ADD  CONSTRAINT  event_api_definitions_package_id_fk  FOREIGN KEY (package_id) REFERENCES packages(id);
ALTER TABLE event_api_definitions ADD  CONSTRAINT  event_api_definitions_application_id_fk FOREIGN KEY (app_id) REFERENCES applications(id) ON DELETE CASCADE;
ALTER TABLE fetch_requests ADD  CONSTRAINT  fetch_requests_document_id_fk FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE;
ALTER TABLE labels ADD  CONSTRAINT  labels_application_id_fk FOREIGN KEY (app_id) REFERENCES applications(id) ON DELETE CASCADE;
ALTER TABLE labels ADD  CONSTRAINT  labels_runtime_id_fk FOREIGN KEY (runtime_id) REFERENCES runtimes(id) ON DELETE CASCADE;
ALTER TABLE packages ADD  CONSTRAINT  packages_application_id_fk FOREIGN KEY (app_id) REFERENCES applications(id) ON DELETE CASCADE;
ALTER TABLE products ADD  CONSTRAINT  products_application_id_fk FOREIGN KEY (app_id) REFERENCES applications(id) ON DELETE CASCADE;
ALTER TABLE tombstones ADD  CONSTRAINT  tombstones_application_id_fk FOREIGN KEY (app_id) REFERENCES applications(id) ON DELETE CASCADE;
ALTER TABLE vendors ADD  CONSTRAINT  vendors_application_id_fk FOREIGN KEY (app_id) REFERENCES applications(id) ON DELETE CASCADE;
ALTER TABLE webhooks ADD  CONSTRAINT  webhooks_runtime_id_fk FOREIGN KEY (runtime_id) REFERENCES runtimes(id) ON DELETE CASCADE;
ALTER TABLE webhooks ADD  CONSTRAINT  webhooks_application_id_fk FOREIGN KEY (app_id) REFERENCES applications(id) ON DELETE CASCADE;
ALTER TABLE system_auths ADD  CONSTRAINT  system_auths_application_id_fk FOREIGN KEY (app_id) REFERENCES applications(id) ON DELETE CASCADE;
ALTER TABLE system_auths ADD  CONSTRAINT  system_auths_runtime_id_fk FOREIGN KEY (runtime_id) REFERENCES runtimes(id) ON DELETE CASCADE;

ALTER TABLE runtime_contexts ADD CONSTRAINT runtime_contexts_runtime_id_fkey FOREIGN KEY (runtime_id) REFERENCES runtimes(id) ON DELETE CASCADE;

-- Drop Columns
ALTER TABLE applications DROP COLUMN  tenant_id;
ALTER TABLE api_definitions DROP  COLUMN  tenant_id;
ALTER TABLE bundles DROP  COLUMN  tenant_id;
ALTER TABLE bundle_references DROP  COLUMN  tenant_id;
ALTER TABLE documents DROP  COLUMN  tenant_id;
ALTER TABLE event_api_definitions DROP  COLUMN  tenant_id;
ALTER TABLE fetch_requests DROP  COLUMN  tenant_id;
ALTER TABLE packages DROP  COLUMN  tenant_id;
ALTER TABLE products DROP  COLUMN  tenant_id;
ALTER TABLE specifications DROP  COLUMN  tenant_id;
ALTER TABLE tombstones DROP  COLUMN  tenant_id;
ALTER TABLE vendors DROP  COLUMN  tenant_id;
ALTER TABLE webhooks DROP  COLUMN  tenant_id;

ALTER TABLE runtimes DROP COLUMN tenant_id;
ALTER TABLE runtime_contexts DROP COLUMN tenant_id;

-- Rename Column
ALTER TABLE bundle_instance_auths RENAME COLUMN  tenant_id to owner_id;

-- Create indices
CREATE INDEX tenant_applications_app_id ON tenant_applications(id);
CREATE INDEX api_definitions_app_id ON api_definitions(app_id);
CREATE INDEX bundle_instance_auths_bundle_id ON bundle_instance_auths(bundle_id);
CREATE INDEX bundle_instance_auths_owner_id ON bundle_instance_auths(owner_id);
CREATE INDEX bundles_app_id ON bundles(app_id);
CREATE INDEX documents_app_id ON documents(app_id);
CREATE INDEX event_api_definitions_app_id ON event_api_definitions(app_id);
CREATE INDEX api_specifications_tenants_app_id ON specifications(api_def_id) WHERE api_def_id IS NOT NULL;
CREATE INDEX event_specifications_tenants_app_id ON specifications(event_def_id) WHERE specifications.event_def_id IS NOT NULL;
CREATE INDEX fetch_request_document_id ON fetch_requests(document_id) WHERE fetch_requests.document_id IS NOT NULL;
CREATE INDEX fetch_request_specification_id ON fetch_requests(document_id) WHERE fetch_requests.spec_id IS NOT NULL;
CREATE INDEX labels_tenant_id ON labels(tenant_id) WHERE labels.tenant_id IS NOT NULL; -- This is not an isolation tenant. Tenants are labelabel objects.
CREATE INDEX labels_app_id ON labels(app_id) WHERE labels.app_id IS NOT NULL;
CREATE INDEX labels_runtime_id ON labels(runtime_id) WHERE labels.runtime_id IS NOT NULL;
CREATE INDEX packages_app_id ON packages(app_id);
CREATE INDEX products_app_id ON products(app_id);
CREATE INDEX tombstones_app_id ON tombstones(app_id);
CREATE INDEX vendors_app_id ON vendors(app_id);
CREATE INDEX webhooks_app_id ON webhooks(app_id) WHERE webhooks.app_id IS NOT NULL;
CREATE INDEX webhooks_runtime_id ON webhooks(runtime_id) WHERE webhooks.runtime_id IS NOT NULL;
CREATE INDEX system_auths_app_id ON system_auths(app_id) WHERE system_auths.app_id IS NOT NULL;
CREATE INDEX system_auths_runtime_id ON system_auths(runtime_id) WHERE system_auths.runtime_id IS NOT NULL;

CREATE INDEX tenant_runtimes_runtimes_id ON tenant_runtimes(id);
CREATE INDEX runtime_contexts_runtime_id ON runtime_contexts(runtime_id);

CREATE UNIQUE INDEX fetch_requests_tenant_id_coalesce_coalesce1_coalesce2_idx ON fetch_requests (coalesce(document_id, '00000000-0000-0000-0000-000000000000'),
                                                                                                 coalesce(spec_id, '00000000-0000-0000-0000-000000000000'));
ALTER TABLE webhooks ADD CONSTRAINT webhook_owner_id_unique
    CHECK ((app_template_id IS NOT NULL AND app_id IS NULL AND runtime_id IS NULL AND integration_system_id IS NULL)
        OR (app_template_id IS NULL AND app_id IS NOT NULL AND runtime_id IS NULL AND integration_system_id IS NULL)
        OR (app_template_id IS NULL AND app_id IS NULL AND runtime_id IS NOT NULL AND integration_system_id IS NULL)
        OR (app_template_id IS NULL AND app_id IS NULL AND runtime_id IS NULL AND integration_system_id IS  NOT NULL));

ALTER TABLE labels ALTER COLUMN tenant_id DROP NOT NULL;
ALTER TABLE labels ADD CONSTRAINT check_scenario_label_is_tenant_scoped CHECK ((key = 'scenarios' AND tenant_id IS NOT NULL) OR key <> 'scenarios');

UPDATE labels SET tenant_id = NULL::uuid WHERE (app_id IS NOT NULL AND key <> 'scenarios') OR (runtime_id IS NOT NULL AND key <> 'scenarios');

-- APIs
CREATE OR REPLACE VIEW api_definitions_tenants AS
SELECT ad.*, ta.tenant_id, ta.owner FROM api_definitions AS ad
                                             INNER JOIN tenant_applications ta ON ta.id = ad.app_id;

-- BundleInstanceAuth
CREATE OR REPLACE VIEW bundle_instance_auths_tenants AS
SELECT bia.*, ta.tenant_id, ta.owner  FROM bundle_instance_auths AS bia
                                               INNER JOIN bundles b ON b.id = bia.bundle_id
                                               INNER JOIN tenant_applications ta ON ta.id = b.app_id;

-- Bundles
CREATE OR REPLACE VIEW bundles_tenants AS
SELECT b.*, ta.tenant_id, ta.owner FROM bundles AS b
                                            INNER JOIN tenant_applications ta ON ta.id = b.app_id;

-- Docs
CREATE OR REPLACE VIEW documents_tenants AS
SELECT d.*, ta.tenant_id, ta.owner FROM documents AS d
                                            INNER JOIN tenant_applications ta ON ta.id = d.app_id;

-- Events
CREATE OR REPLACE VIEW event_api_definitions_tenants AS
SELECT e.*, ta.tenant_id, ta.owner FROM event_api_definitions AS e
                                            INNER JOIN tenant_applications ta ON ta.id = e.app_id;

-- Specs
CREATE VIEW api_specifications_tenants AS
(SELECT s.*, ta.tenant_id, ta.owner FROM specifications AS s
                                            INNER JOIN api_definitions AS ad ON ad.id = s.api_def_id
                                            INNER JOIN tenant_applications ta on ta.id = ad.app_id);

CREATE VIEW event_specifications_tenants AS
SELECT s.*, ta.tenant_id, ta.owner FROM specifications AS s
                                            INNER JOIN event_api_definitions AS ead ON ead.id = s.event_def_id
                                            INNER JOIN tenant_applications ta on ta.id = ead.app_id;

-- Fetch Requests
CREATE OR REPLACE VIEW document_fetch_requests_tenants AS
SELECT fr.*, ta.tenant_id, ta.owner FROM fetch_requests AS fr
                                             INNER JOIN documents d ON fr.document_id = d.id
                                             INNER JOIN tenant_applications ta ON ta.id = d.app_id;

CREATE OR REPLACE VIEW api_specifications_fetch_requests_tenants AS
SELECT fr.*, ta.tenant_id, ta.owner FROM fetch_requests AS fr
                                             INNER JOIN specifications s ON fr.spec_id = s.id
                                             INNER JOIN api_definitions AS ad ON ad.id = s.api_def_id
                                             INNER JOIN tenant_applications ta on ta.id = ad.app_id;

CREATE OR REPLACE VIEW event_specifications_fetch_requests_tenants AS
SELECT fr.*, ta.tenant_id, ta.owner FROM fetch_requests AS fr
                                             INNER JOIN specifications s ON fr.spec_id = s.id
                                             INNER JOIN event_api_definitions AS ead ON ead.id = s.event_def_id
                                             INNER JOIN tenant_applications ta on ta.id = ead.app_id;

-- Labels - since labels can be put on tenants we cannot get l.*, however this is
-- not a problem because labels are not necessary for the ORD service which is the only component reading from the view
CREATE OR REPLACE VIEW application_labels_tenants AS
SELECT l.id, ta.tenant_id, ta.owner FROM labels AS l
                                             INNER JOIN tenant_applications ta
                                                        ON l.app_id = ta.id AND (l.tenant_id IS NULL OR l.tenant_id = ta.tenant_id);

CREATE OR REPLACE VIEW runtime_labels_tenants AS
SELECT l.id, tr.tenant_id, tr.owner FROM labels AS l
                                             INNER JOIN tenant_runtimes tr
                                                        ON l.runtime_id = tr.id AND (l.tenant_id IS NULL OR l.tenant_id = tr.tenant_id);

CREATE OR REPLACE VIEW runtime_contexts_labels_tenants AS
SELECT l.id, tr.tenant_id, tr.owner FROM labels AS l
                                             INNER JOIN runtime_contexts rc ON l.runtime_context_id = rc.id
                                             INNER JOIN tenant_runtimes tr ON rc.runtime_id = tr.id AND (l.tenant_id IS NULL OR l.tenant_id = tr.tenant_id);


-- Aggregated labels view
CREATE OR REPLACE VIEW labels_tenants AS
(SELECT l.id, ta.tenant_id, ta.owner FROM labels AS l
                                             INNER JOIN tenant_applications ta
                                                        ON l.app_id = ta.id AND (l.tenant_id IS NULL OR l.tenant_id = ta.tenant_id))
UNION ALL
(SELECT l.id, tr.tenant_id, tr.owner FROM labels AS l
                                             INNER JOIN tenant_runtimes tr
                                                        ON l.runtime_id = tr.id AND (l.tenant_id IS NULL OR l.tenant_id = tr.tenant_id))
UNION ALL
(SELECT l.id, tr.tenant_id, tr.owner FROM labels AS l
                                             INNER JOIN runtime_contexts rc ON l.runtime_context_id = rc.id
                                             INNER JOIN tenant_runtimes tr ON rc.runtime_id = tr.id AND (l.tenant_id IS NULL OR l.tenant_id = tr.tenant_id));

-- Runtime Context
CREATE OR REPLACE VIEW runtime_contexts_tenants AS
SELECT rtc.*, tr.tenant_id, tr.owner FROM runtime_contexts AS rtc
                                              INNER JOIN tenant_runtimes tr ON rtc.runtime_id = tr.id;

-- Packages
CREATE OR REPLACE VIEW packages_tenants AS
SELECT p.*, ta.tenant_id, ta.owner FROM packages AS p
                                            INNER JOIN tenant_applications AS ta ON ta.id = p.app_id;

-- Products
CREATE OR REPLACE VIEW products_tenants AS
SELECT p.*, ta.tenant_id, ta.owner FROM products AS p
                                            INNER JOIN tenant_applications AS ta ON ta.id = p.app_id;

-- Tombstones
CREATE OR REPLACE VIEW tombstones_tenants AS
SELECT t.*, ta.tenant_id, ta.owner FROM tombstones AS t
                                            INNER JOIN tenant_applications AS ta ON ta.id = t.app_id;

-- Vendors
CREATE OR REPLACE VIEW vendors_tenants AS
SELECT v.*, ta.tenant_id, ta.owner FROM vendors AS v
                                            INNER JOIN tenant_applications AS ta ON ta.id = v.app_id;

-- Webhooks
CREATE OR REPLACE VIEW application_webhooks_tenants AS
SELECT w.*, ta.tenant_id, ta.owner FROM webhooks AS w
                                            INNER JOIN tenant_applications ta ON w.app_id = ta.id;

CREATE OR REPLACE VIEW runtime_webhooks_tenants AS
SELECT w.*, tr.tenant_id, tr.owner FROM webhooks AS w
                                            INNER JOIN tenant_runtimes tr ON w.runtime_id = tr.id;


-- Aggregated Webhooks
CREATE OR REPLACE VIEW webhooks_tenants AS
(SELECT w.*, ta.tenant_id, ta.owner FROM webhooks AS w
                                            INNER JOIN tenant_applications ta ON w.app_id = ta.id)
UNION ALL
(SELECT w.*, tr.tenant_id, tr.owner FROM webhooks AS w
                                            INNER JOIN tenant_runtimes tr ON w.runtime_id = tr.id);

-- ASAs Redesign
ALTER TABLE automatic_scenario_assignments ADD COLUMN target_tenant_id UUID REFERENCES business_tenant_mappings(id) ON DELETE CASCADE;

UPDATE automatic_scenario_assignments asa SET target_tenant_id = (SELECT id
                                                              FROM business_tenant_mappings
                                                              WHERE external_tenant = asa.selector_value)
WHERE selector_key = 'global_subaccount_id' AND EXISTS(SELECT id
                                                        FROM business_tenant_mappings
                                                        WHERE external_tenant = asa.selector_value);

DELETE FROM automatic_scenario_assignments WHERE target_tenant_id IS NULL;

ALTER TABLE automatic_scenario_assignments ALTER COLUMN target_tenant_id SET NOT NULL;

ALTER TABLE automatic_scenario_assignments DROP COLUMN selector_key;
ALTER TABLE automatic_scenario_assignments DROP COLUMN selector_value;

CREATE OR REPLACE FUNCTION check_tenant_id_is_direct_parent_of_target_tenant_id() RETURNS TRIGGER AS
$$
BEGIN
    EXECUTE format('SELECT 1 FROM business_tenant_mappings WHERE id = %L AND parent = %L', TG_TABLE_NAME, NEW.target_tenant_id, NEW.tenant_id);
    IF NOT FOUND THEN
        RAISE EXCEPTION 'target_tenant_id should be direct child of tenant_id';
    END IF;
    RETURN NULL;
END
$$ LANGUAGE plpgsql;

CREATE CONSTRAINT TRIGGER tenant_id_is_direct_parent_of_target_tenant_id AFTER INSERT ON automatic_scenario_assignments
    FOR EACH ROW EXECUTE PROCEDURE check_tenant_id_is_direct_parent_of_target_tenant_id();

-- Label Definitions restrictions

DELETE FROM label_definitions WHERE key <> 'scenarios';
ALTER TABLE label_definitions ADD CONSTRAINT key_is_scenario CHECK(key = 'scenarios');

-- TODO: ORD Service Views

COMMIT;
