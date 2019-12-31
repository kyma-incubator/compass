ALTER TABLE applications
    ADD COLUMN provider_display_name varchar(256);

UPDATE applications
    SET provider_display_name = provider_name;

ALTER TABLE applications
    DROP COLUMN provider_name;
