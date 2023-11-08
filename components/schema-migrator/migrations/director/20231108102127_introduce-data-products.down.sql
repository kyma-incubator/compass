BEGIN;

-- Add runtimeRestriction to packages
ALTER TABLE packages 
       DROP COLUMN runtime_restriction VARCHAR(256);

COMMIT;
