BEGIN;

ALTER TABLE formation_templates
    DROP COLUMN leading_product_ids;

COMMIT;
