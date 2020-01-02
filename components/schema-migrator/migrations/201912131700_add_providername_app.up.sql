ALTER TABLE applications
    ADD COLUMN provider_display_name varchar(256);

UPDATE applications
SET provider_display_name = name;

ALTER TABLE applications
    ALTER COLUMN provider_display_name SET NOT NULL;
