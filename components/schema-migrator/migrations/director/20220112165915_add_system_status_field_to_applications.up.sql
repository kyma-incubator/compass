BEGIN;

ALTER TABLE applications
    ADD COLUMN system_status varchar;

COMMIT;
