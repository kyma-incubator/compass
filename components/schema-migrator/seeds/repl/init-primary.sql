CREATE EXTENSION IF NOT EXISTS pglogical;
SELECT pglogical.create_node(
               node_name := 'primary',
               dsn := 'host=test-postgres-replica port=5432 dbname=compass'
       );