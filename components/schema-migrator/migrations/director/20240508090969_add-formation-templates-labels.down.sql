BEGIN;

DROP INDEX labels_formation_template_id;

ALTER TABLE labels DROP CONSTRAINT labels_formation_template_id_fk;

ALTER TABLE labels DROP COLUMN formation_template_id;

ALTER table formation_templates
    DROP created_at,
    DROP updated_at;

COMMIT;
