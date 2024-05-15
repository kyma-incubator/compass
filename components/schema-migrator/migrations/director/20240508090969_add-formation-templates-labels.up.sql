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

COMMIT;
