BEGIN;

ALTER TABLE formation_assignments
    ALTER COLUMN source_type TYPE TEXT,
    ALTER COLUMN target_type TYPE TEXT;

ALTER TABLE formation_assignments
    ADD CONSTRAINT source_type_check CHECK (source_type IN ('APPLICATION', 'RUNTIME', 'RUNTIME_CONTEXT')),
    ADD CONSTRAINT target_type_check CHECK (target_type IN ('APPLICATION', 'RUNTIME', 'RUNTIME_CONTEXT'));

COMMIT;
