BEGIN;

ALTER TABLE formations DROP CONSTRAINT formations_state_check;

ALTER TABLE formations
    ADD CONSTRAINT formations_state_check
        CHECK (state = ANY
               (ARRAY['INITIAL'::text, 'READY'::text, 'CREATE_ERROR'::text, 'DELETE_ERROR'::text, 'DELETING'::text));

COMMIT;