package storage

import "fmt"

const (
	instancesTableName       = "instances"
	provisionParamsTableName = "provisioning_parameters"
)

var schema = map[string]string{
	instancesTableName: fmt.Sprintf(`CREATE TABLE %s (
			instance_id serial PRIMARY KEY,
			runtime_id varchar(255) NOT NULL,
			global_account_id varchar(255) NOT NULL,
			service_id varchar(255) NOT NULL,
			service_plan_id varchar(255) NOT NULL,
			dashboard_url varchar(255),
			parameters_id varchar(255),
			UNIQUE(instance_id),
			UNIQUE(global_account_id),
    		foreign key (parameters_id) REFERENCES %s (id) ON DELETE CASCADE,
		)`, instancesTableName, provisionParamsTableName),

	provisionParamsTableName: fmt.Sprintf(`CREATE TABLE %s (
			id serial PRIMARY KEY,
			name varchar(255) NOT NULL,
			node_Count integer NOT NULL,
    		volume_size_gb varchar(256) NOT NULL,
    		machine_type varchar(256) NOT NULL,
    		region varchar(256) NOT NULL,
			zone varchar(256) NOT NULL,
			auto_scaler_min integer NOT NULL,
    		auto_scaler_max integer NOT NULL,
			max_surge integer NOT NULL,
    		max_unavailable integer NOT NULL,
		)`, provisionParamsTableName),
}
