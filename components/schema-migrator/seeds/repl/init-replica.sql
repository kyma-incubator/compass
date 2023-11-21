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



-- SELECT * FROM pglogical.show_subscription_status();
-- SELECT * FROM pglogical.drop_subscription('subscription1') ;
-- SELECT * FROM pglogical.events;