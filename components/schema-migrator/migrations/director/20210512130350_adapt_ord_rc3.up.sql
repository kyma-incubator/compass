BEGIN;

ALTER TABLE vendors
    DROP COLUMN sap_partner;

ALTER TABLE vendors
    ADD COLUMN partners JSONB;

CREATE VIEW partners AS
SELECT vendors.ord_id    AS vendor_id,
       elements.value    AS value
FROM vendors,
     jsonb_array_elements_text(vendors.partners) AS elements;

COMMIT;
