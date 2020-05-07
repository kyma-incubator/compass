BEGIN;

ALTER TABLE runtimes
    ALTER COLUMN status_condition TYPE VARCHAR(255);

ALTER TABLE runtimes
    ALTER COLUMN status_condition DROP DEFAULT;

ALTER TYPE runtime_status_condition
    RENAME TO runtime_status_condition_old;

CREATE TYPE runtime_status_condition AS ENUM (
    'INITIAL',
    'PROVISIONING',
    'CONNECTED',
    'FAILED'
);

ALTER TABLE runtimes
    ALTER COLUMN status_condition TYPE runtime_status_condition
    USING status_condition::runtime_status_condition;

ALTER TABLE runtimes
    ALTER COLUMN status_condition
    SET DEFAULT 'INITIAL' ::runtime_status_condition;

DROP TYPE runtime_status_condition_old;

COMMIT;
