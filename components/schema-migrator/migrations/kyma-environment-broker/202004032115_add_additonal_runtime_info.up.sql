ALTER TABLE instances
 ADD COLUMN sub_account_id varchar(255) DEFAULT '',
 ADD COLUMN service_name varchar(255) DEFAULT '',
 ADD COLUMN service_plan_name varchar(255) DEFAULT '';
