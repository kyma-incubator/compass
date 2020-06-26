BEGIN;

ALTER TABLE cluster ADD COLUMN terraform_state bytea;
ALTER TABLE cluster ADD COLUMN credentials_secret_name varchar(256) NOT NULL default '';

COMMIT;
