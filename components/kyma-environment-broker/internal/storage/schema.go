package storage

import "fmt"

const (
	InstancesTableName       = "instances"
	ProvisionParamsTableName = "provisioning_parameters"
)

var schema = []string{
	fmt.Sprintf(`CREATE TABLE %s (
			params_id uuid PRIMARY KEY,
			name varchar(255) NOT NULL,
			node_Count integer NOT NULL,
    		volume_size_gb varchar(256) NOT NULL,
    		machine_type varchar(256) NOT NULL,
    		region varchar(256) NOT NULL,
			zone varchar(256) NOT NULL,
			auto_scaler_min integer NOT NULL,
    		auto_scaler_max integer NOT NULL,
			max_surge integer NOT NULL,
    		max_unavailable integer NOT NULL
		)`, ProvisionParamsTableName),

	fmt.Sprintf(`CREATE TABLE %s (
			instance_id uuid PRIMARY KEY,
			runtime_id varchar(255) NOT NULL,
			global_account_id varchar(255) NOT NULL,
			service_id varchar(255) NOT NULL,
			service_plan_id varchar(255) NOT NULL,
			dashboard_url varchar(255) NOT NULL,
			parameters_id uuid NOT NULL,
			UNIQUE(instance_id),
			UNIQUE(global_account_id),
			UNIQUE(parameters_id),
    		foreign key (parameters_id) REFERENCES %s (params_id) ON DELETE CASCADE
		)`, InstancesTableName, ProvisionParamsTableName),
}
