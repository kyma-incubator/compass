BEGIN;

--- vendors ---

ALTER TABLE vendors
ADD COLUMN id UUID CHECK (id <> '00000000-0000-0000-0000-000000000000');

UPDATE vendors
SET id = uuid_generate_v4();

ALTER TABLE vendors DROP CONSTRAINT vendors_pkey;
ALTER TABLE vendors ADD CONSTRAINT vendors_pkey PRIMARY KEY (id);

--- products ---

ALTER TABLE products
    ADD COLUMN id UUID CHECK (id <> '00000000-0000-0000-0000-000000000000');

UPDATE products
SET id = uuid_generate_v4();

ALTER TABLE products DROP CONSTRAINT products_pkey;
ALTER TABLE products ADD CONSTRAINT products_pkey PRIMARY KEY (id);

COMMIT;
