BEGIN;

CREATE OR REPLACE FUNCTION get_id_tenant_id_index()
    RETURNS TABLE(table_name TEXT, id UUID, tenant_id UUID) AS
$func$
DECLARE
    compass_table text;
    sql varchar;
BEGIN
    FOR compass_table IN
        SELECT DISTINCT t.table_name
        FROM   information_schema.tables t
                   INNER JOIN information_schema.columns c1 ON t.table_name = c1.table_name
                   INNER JOIN information_schema.columns c2 ON t.table_name = c2.table_name
        WHERE  t.table_schema = 'public'
          AND t.table_type = 'BASE TABLE'
          AND c1.column_name = 'id'
          AND c2.column_name = 'tenant_id'
        LOOP
            sql := 'SELECT ''' || compass_table || '''::TEXT as table_name, id, tenant_id FROM public.' || compass_table || ' WHERE tenant_id IS NOT NULL;';
            RAISE NOTICE 'Executing SQL query: %', sql;
            RETURN QUERY EXECUTE sql;
        END LOOP;
END
$func$  LANGUAGE plpgsql;

-- An index view containing all the entity ids with their owning tenant.
-- It is dynamically updated with the data from each table in the public schema containing 'id' and 'tenant_id' columns.
CREATE MATERIALIZED VIEW id_tenant_id_index AS
SELECT id, tenant_id FROM get_id_tenant_id_index();

CREATE UNIQUE INDEX id_tenant_id_index_unique ON id_tenant_id_index(id);

COMMIT;
