BEGIN;

ALTER TABLE formation_templates
    ADD COLUMN leading_product_ids JSONB;

COMMIT;
