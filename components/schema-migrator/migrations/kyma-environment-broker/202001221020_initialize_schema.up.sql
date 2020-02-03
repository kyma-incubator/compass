-- Instances

CREATE TABLE IF NOT EXISTS  instances (
    instance_id varchar(255) PRIMARY KEY,
    runtime_id varchar(255) NOT NULL,
    global_account_id varchar(255) NOT NULL,
    service_id varchar(255) NOT NULL,
    service_plan_id varchar(255) NOT NULL,
    dashboard_url varchar(255) NOT NULL,
    provisioning_parameters text NOT NULL
);
