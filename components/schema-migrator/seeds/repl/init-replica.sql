CREATE EXTENSION IF NOT EXISTS pglogical;
SELECT pglogical.create_node(
               node_name := 'replica',
               dsn := 'host=thishost port=5432 dbname=compass'
       );

SELECT pglogical.create_subscription(
               subscription_name := 'subscription1',
               provider_dsn := 'host=test-postgres port=5432 dbname=compass'
       );

SELECT pglogical.wait_for_subscription_sync_complete('subscription1');