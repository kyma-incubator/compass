CREATE EXTENSION IF NOT EXISTS pglogical;
SELECT pglogical.create_node(
               node_name := 'replica',
               dsn := 'host=test-postgres-replica port=5432 dbname=compass password=pgsql@12345 sslmode=disable'
       );

CREATE TABLE public.event_outbox
(
    id         SERIAL PRIMARY KEY,
    message    json,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

SELECT pglogical.create_subscription(
               subscription_name := 'subscription1',
               provider_dsn := 'host=test-postgres port=5432 dbname=compass password=pgsql@12345 sslmode=disable',
               synchronize_data := TRUE
       );

SELECT pglogical.wait_for_subscription_sync_complete('subscription1');

-- Option for reading the events directly in sql
-- SELECT 'init' FROM pg_create_logical_replication_slot('test_slot_2', 'wal2json');
-- SELECT data FROM pg_logical_slot_get_changes('test_slot_2', NULL, NULL, 'format-version', '2', 'add-msg-prefixes', 'wal2json');

-- SELECT * FROM pglogical.show_subscription_status();
-- SELECT * FROM pglogical.drop_subscription('subscription1') ;
-- SELECT * FROM pglogical.events;