BEGIN;

--- vendors ---

ALTER TABLE vendors DROP CONSTRAINT vendors_pkey;
ALTER TABLE vendors ADD CONSTRAINT vendors_pkey PRIMARY KEY (ord_id);

ALTER TABLE vendors
DROP COLUMN id;

--- products ---

ALTER TABLE products DROP CONSTRAINT products_pkey;
ALTER TABLE products ADD CONSTRAINT products_pkey PRIMARY KEY (ord_id);

ALTER TABLE products
    DROP COLUMN id;

COMMIT;
