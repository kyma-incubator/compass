-- runtimes

ALTER TABLE runtimes
    ALTER COLUMN status_condition TYPE VARCHAR(255);

ALTER TABLE runtimes
    ALTER COLUMN status_condition DROP DEFAULT;

UPDATE runtimes
    SET status_condition = 'READY'
    WHERE status_condition ='CONNECTED';

ALTER TYPE runtime_status_condition 
    RENAME TO runtime_status_condition_old;

CREATE TYPE runtime_status_condition AS ENUM (
    'INITIAL',
    'READY',
    'FAILED'
);

ALTER TABLE runtimes 
    ALTER COLUMN status_condition TYPE runtime_status_condition 
    USING status_condition::runtime_status_condition;

ALTER TABLE runtimes
    ALTER COLUMN status_condition
    SET DEFAULT 'INITIAL' ::runtime_status_condition;

DROP TYPE runtime_status_condition_old;

-- applications

ALTER TABLE applications 
    ALTER COLUMN status_condition TYPE VARCHAR(255);

ALTER TABLE applications 
    ALTER COLUMN status_condition DROP DEFAULT;

UPDATE applications
    SET status_condition = 'READY'
    WHERE status_condition = 'CONNECTED';

ALTER TYPE application_status_condition 
    RENAME TO application_status_condition_old;

CREATE TYPE application_status_condition AS ENUM (
    'INITIAL',
    'UNKNOWN'
    'CONNECTED',
    'FAILED'
);

ALTER TABLE applications 
    ALTER COLUMN status_condition TYPE application_status_condition 
    USING status_condition::application_status_condition;

ALTER TABLE applications
    ALTER COLUMN status_condition
    SET DEFAULT 'INITIAL' ::application_status_condition;

DROP TYPE application_status_condition_old;
