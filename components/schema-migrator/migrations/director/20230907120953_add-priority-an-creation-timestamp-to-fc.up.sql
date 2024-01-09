BEGIN;

ALTER TABLE formation_constraints
    ADD COLUMN description VARCHAR(255) DEFAULT '';

ALTER TABLE formation_constraints
    ADD COLUMN priority INTEGER DEFAULT 0;

ALTER TABLE formation_constraints
    ADD COLUMN created_at TIMESTAMP;

UPDATE
    formation_constraints
SET created_at = CURRENT_TIMESTAMP
WHERE created_at IS NULL;

COMMIT;
