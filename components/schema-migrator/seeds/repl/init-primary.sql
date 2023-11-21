CREATE EXTENSION IF NOT EXISTS pglogical;
SELECT pglogical.create_node(
               node_name := 'primary',
               dsn := 'host=test-postgres port=5432 dbname=compass sslmode=disable'
       );
-- This does not work because there are tables with missing primary key
SELECT pglogical.replication_set_add_all_tables('default', ARRAY['public'], true);

SELECT pglogical.replication_set_add_table('default', 'business_tenant_mappings', true);