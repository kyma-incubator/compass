BEGIN;

ALTER TABLE cluster DROP COLUMN terraform_state;
ALTER TABLE cluster DROP COLUMN credentials_secret_name;

COMMIT;
