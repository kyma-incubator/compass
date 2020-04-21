BEGIN;
DELETE FROM LABELS AS L WHERE L.key='scenarios' AND L.runtime_id IS NOT NULL AND L.value = 'null'::jsonb;
COMMIT;
