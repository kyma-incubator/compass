ALTER TABLE applications
    ADD COLUMN provider_name varchar(256);

UPDATE APPLICATIONS AS A
SET provider_name = name;

ALTER TABLE applications
    ALTER COLUMN provider_name SET NOT NULL;
