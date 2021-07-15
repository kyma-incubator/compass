BEGIN;

DROP INDEX parent_index;

DROP MATERIALIZED VIEW id_tenant_id_index;
DROP FUNCTION get_id_tenant_id_index();

COMMIT;
