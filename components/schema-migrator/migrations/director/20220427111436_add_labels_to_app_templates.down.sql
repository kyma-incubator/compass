BEGIN;

ALTER TABLE labels DROP CONSTRAINT labels_app_template_id_fk;

ALTER TABLE labels DROP COLUMN app_template_id;

COMMIT;
