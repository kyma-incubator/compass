BEGIN;

ALTER TABLE applications
    ADD COLUMN system_status TEXT;

ALTER TABLE applications
    ADD CONSTRAINT system_status_len CHECK (length(system_status) <= 25);

COMMIT;
