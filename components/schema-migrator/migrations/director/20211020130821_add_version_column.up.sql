BEGIN;

ALTER TABLE label_definitions
    ADD COLUMN version integer default 0;

ALTER TABLE labels
    ADD COLUMN version integer default 0;

COMMIT;