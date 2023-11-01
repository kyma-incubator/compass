BEGIN;

UPDATE labels SET
    value = jsonb_array_element(value, 0)
WHERE key = 'cldSystemRole' AND jsonb_typeof(value) = 'array';

COMMIT;