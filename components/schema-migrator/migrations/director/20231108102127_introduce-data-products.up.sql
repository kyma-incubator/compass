BEGIN;

-- Add runtimeRestriction to packages
ALTER TABLE packages 
       ADD COLUMN runtime_restriction VARCHAR(256);

COMMIT;
