CREATE EXTENSION IF NOT EXISTS pglogical;
SELECT pglogical.create_node(
               node_name := 'replica',
               dsn := 'host=thishost port=5432 dbname=compass'
       );