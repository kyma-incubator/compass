BEGIN;

CREATE VIEW scheduled_operations AS
    SELECT * FROM operation
         WHERE status = 'SCHEDULED'
         ORDER BY priority DESC, updated_at;

COMMIT;
