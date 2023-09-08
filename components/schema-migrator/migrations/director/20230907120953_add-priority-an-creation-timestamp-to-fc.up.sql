BEGIN;

ALTER TABLE formation_constraints
    ADD COLUMN description VARCHAR(255);

ALTER TABLE formation_constraints
    ADD COLUMN priority INTEGER default 0;

ALTER TABLE formation_constraints
    ADD COLUMN created_at timestamp;

COMMIT;
