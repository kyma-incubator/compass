package storage

import "fmt"

const (
	InstancesTableName = "instances"
)

var schema = []string{
	fmt.Sprintf(`CREATE TABLE %s (
			instance_id uuid PRIMARY KEY,
			runtime_id varchar(255) NOT NULL,
			global_account_id varchar(255) NOT NULL,
			service_id varchar(255) NOT NULL,
			service_plan_id varchar(255) NOT NULL,
			dashboard_url varchar(255) NOT NULL,
			parameters_id varchar(2000) NOT NULL,
			UNIQUE(instance_id),
		)`, InstancesTableName),
}
