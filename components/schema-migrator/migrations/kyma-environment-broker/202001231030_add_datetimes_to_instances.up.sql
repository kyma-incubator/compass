ALTER TABLE instances
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

ALTER TABLE instances
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

ALTER TABLE instances
    ADD COLUMN delated_at TIMESTAMPTZ NOT NULL;
