BEGIN;

ALTER TABLE formation_constraints
DROP
COLUMN description;

ALTER TABLE formation_constraints
DROP
COLUMN priority;

ALTER TABLE formation_constraints
DROP
COLUMN created_at;

COMMIT;
