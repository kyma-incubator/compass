BEGIN;

ALTER TABLE packages
  RENAME TO bundles;

CREATE TABLE packages (
    id UUID PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    owner_app_id uuid NOT NULL, /* if a packag is rreturned by a system, only that system is expeccted iis expected to return this package, otherwise -> validation error */
    FOREIGN KEY (owner_app_id) REFERENCES applications (id) ON DELETE CASCADE,
    title VARCHAR(256) NOT NULL,
    short_description VARCHAR(256) NOT NULL,
    description TEXT NOT NULL,
    version: VARCHAR(64) NOT NULL,
    release_status: VARCHAR(256), /* spec doesn't define NULL fields */
    api_protocol VARCHAR(256),
    licence_type VARCHAR(256),
    terms_of_service VARCHAR(512),
    logo VARCHAR(512),
    image VARCHAR(512),
    provider JSONB, /* duplication w/t provider in application/system resource */
    tags JSONB, /* consider how to store tags to be queriable */
    actions JSONB,
    last_updated TIMESTAMP NOT NULL,
    extensions JSOB; /* The spec MAY be extended with custom properties. Their property names MUST start with "x-"  */
);

ALTER TABLE bundles
    ADD COLUMN package_id UUID; /* NULLable because we don't have packages for existing systems */

ALTER TABLE api_definitions
    RENAME COLUMN name TO title,
    RENAME COLUMN package_id TO bundle_id,
    RENAME COLUMN version_value TO version;

ALTER TABLE api_definitions
DROP CONSTRAINT api_definitions_package_id_fk,
ADD CONSTRAINT api_definitions_package_id_fk
    FOREIGN KEY (bundle_id)
    REFERENCES bundles(id)
    ON DELETE CASCADE;

ALTER TABLE api_definitions
    ADD COLUMN short_description VARCHAR(256) NOT NULL,
    ADD COLUMN api_definitions JSONB NOT NULL, /* spec_data, spec_format and spec_type wiil be populated programatically baased on the first element here */
    ADD COLUMN tags JSONB,
    ADD COLUMN documentation VARCHAR(512),
    ADD COLUMN changelog_entries JSONB,
    ADD COLUMN logo VARCHAR(512),
    ADD COLUMN image VARCHAR(512),
    ADD COLUMN url VARCHAR(512),
    ADD COLUMN release_status VARCHAR(64) NOT NULL, /* should be ENUM */
    ADD COLUMN api_protocol VARCHAR(64) NOT NULL, /* should be ENUM */
    ADD COLUMN entry_point JSONB NOT NULL,
    ADD COLUMN actions JSONB NOT NULL,
    ADD COLUMN last_updated TIMESTAMP NOT NULL,
    ADD COLUMN extensions JSOB; /* The spec MAY be extended with custom properties. Their property names MUST start with "x-"  */


COMMIT;
