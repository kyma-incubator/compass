BEGIN;

ALTER TABLE labels
    ADD COLUMN formation_template_id UUID;

ALTER TABLE labels
    ADD CONSTRAINT labels_formation_template_id_fk FOREIGN KEY (formation_template_id) REFERENCES formation_templates(id) ON DELETE CASCADE;

CREATE INDEX labels_formation_template_id
    ON labels (formation_template_id)
    WHERE (labels.formation_template_id IS NOT NULL);


ALTER TABLE formation_templates
    ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    ADD COLUMN updated_at TIMESTAMP;

------------------------------------------------------------------------------------
------------------------------------------------------------------------------------


CREATE OR REPLACE VIEW formation_template_labels_tenants AS
SELECT l.id, tr.tenant_id, tr.owner FROM labels AS l
                                             INNER JOIN tenant_runtimes tr
                                                        ON l.runtime_id = tr.id AND (l.tenant_id IS NULL OR l.tenant_id = tr.tenant_id);

CREATE OR REPLACE VIEW formation_template_labels_tenants AS
SELECT l.id, tr.tenant_id, tr.owner FROM labels AS l
                                             INNER JOIN formation_templates ft
                                                        ON l.formation_template_id = ft.id AND (l.tenant_id IS NULL OR l.tenant_id = ft.tenant_id);

------------------------------------------------------------------------------------
------------------------------------------------------------------------------------

CREATE OR REPLACE VIEW formation_template_labels_tenants AS
SELECT l.id, tr.tenant_id, tr.owner FROM labels AS l
                                             INNER JOIN runtime_contexts rc ON l.runtime_context_id = rc.id
                                             INNER JOIN tenant_runtimes tr ON rc.runtime_id = tr.id AND (l.tenant_id IS NULL OR l.tenant_id = tr.tenant_id);

CREATE OR REPLACE VIEW formation_template_labels_tenants AS
SELECT l.id, btm.id, btm.owner FROM labels AS l
                                             INNER JOIN formation_templates ft ON l.formation_template_id = ft.id
                                             INNER JOIN business_tenant_mappings btm ON ft.tenant_id = btm.id AND (l.tenant_id IS NULL OR l.tenant_id = btm.id);


COMMIT;
