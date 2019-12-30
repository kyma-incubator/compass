ALTER TABLE applications
    ADD COLUMN provider_name varchar(256);

UPDATE applications
SET provider_name = provider_display_name;

ALTER TABLE applications
    ALTER COLUMN provider_name SET NOT NULL;

ALTER TABLE applications
    DROP COLUMN provider_display_name;
