SET statement_timeout = 0;
CREATE EXTENSION IF NOT EXISTS plv2;
CREATE PUBLICATION compass_replication FOR ALL TABLES;
ALTER SUBSCRIPTION compass_subscription OWNER TO replicator;
ALTER SUBSCRIPTION compass_subscription CONNECTION 'host=test-postgres port=5432 user=replicator password=repl123';
START SUBSCRIPTION compass_subscription;