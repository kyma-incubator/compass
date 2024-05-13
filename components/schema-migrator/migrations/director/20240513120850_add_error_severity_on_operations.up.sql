BEGIN;

CREATE TYPE operation_error_severity AS ENUM (
    'Error',
    'Warning',
    'Info'
);

ALTER TABLE operation
    ADD COLUMN error_severity operation_error_severity;

COMMIT;
