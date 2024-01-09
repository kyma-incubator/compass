BEGIN;

UPDATE labels SET
    value = jsonb_build_array(value)
WHERE key = 'cldSystemRole' AND jsonb_typeof(value) != 'array';

COMMIT;