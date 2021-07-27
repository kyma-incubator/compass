BEGIN;

ALTER TABLE labels drop CONSTRAINT valid_refs;

COMMIT;
