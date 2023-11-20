CREATE USER replicator REPLICATION PASSWORD 'repl123';
GRANT CONNECT ON DATABASE compass TO replicator;
GRANT replication TO replicator;