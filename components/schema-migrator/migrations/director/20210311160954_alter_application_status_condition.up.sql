BEGIN;

ALTER TABLE applications
    ALTER COLUMN status_condition DROP DEFAULT;

ALTER TYPE application_status_condition RENAME TO application_status_condition_old;
CREATE TYPE application_status_condition AS ENUM (
    'INITIAL',
    'CONNECTED',
    'FAILED',
    'CREATING',
    'CREATE_FAILED',
    'CREATE_SUCCEEDED',
    'UPDATING',
    'UPDATE_FAILED',
    'UPDATE_SUCCEEDED',
    'DELETING',
    'DELETE_FAILED',
    'DELETE_SUCCEEDED'
);

ALTER TABLE applications
    ALTER COLUMN status_condition TYPE application_status_condition
        USING status_condition::text::application_status_condition;

ALTER TABLE applications
    ALTER COLUMN status_condition
        SET DEFAULT 'INITIAL' ::application_status_condition;

DROP TYPE application_status_condition_old;

COMMIT;
