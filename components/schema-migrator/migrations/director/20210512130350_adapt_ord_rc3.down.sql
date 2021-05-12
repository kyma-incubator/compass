BEGIN;

DROP VIEW partners;

ALTER TABLE vendors
    ADD COLUMN sap_partner BOOLEAN;

ALTER TABLE vendors
    DROP COLUMN partners;

COMMIT;
