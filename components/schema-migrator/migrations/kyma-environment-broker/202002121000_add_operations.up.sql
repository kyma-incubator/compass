CREATE TABLE IF NOT EXISTS operations (
    id varchar(255) PRIMARY KEY,
    instance_id varchar(255) NOT NULL,
    target_operation_id varchar(255) NOT NULL,
    version integer NOT NULL,
    state varchar(32) NOT NULL,
    description text NOT NULL,
    type varchar(32) NOT NULL,
    data json NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
