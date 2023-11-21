CREATE EXTENSION IF NOT EXISTS pglogical;
SELECT pglogical.create_node(
               node_name := 'replica',
               dsn := 'host=test-postgres-replica port=5432 dbname=compass password=pgsql@12345 sslmode=disable'
       );

SELECT pglogical.create_subscription(
               subscription_name := 'subscription1',
               provider_dsn := 'host=test-postgres port=5432 dbname=compass password=pgsql@12345 sslmode=disable',
               synchronize_structure := true,
               synchronize_data := true
       );

SELECT pglogical.wait_for_subscription_sync_complete('subscription1');

-- SELECT * FROM pglogical.show_subscription_status();
-- SELECT * FROM pglogical.drop_subscription('subscription1') ;