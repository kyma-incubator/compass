UPDATE applications SET provider_name='' WHERE provider_name IS NULL;

ALTER TABLE applications
    ALTER COLUMN provider_name SET NOT NULL;
