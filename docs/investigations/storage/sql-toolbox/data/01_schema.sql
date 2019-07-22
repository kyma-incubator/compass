create database compass;

\connect compass

create table applications (
  id varchar(100),
  tenant varchar(100),
  name varchar(100),
  description varchar (100),
  labels JSON,
  primary key (id)
);


create table apis (
  id varchar(100),
  target_url varchar(100),
  app_id varchar(100),
  primary key (id),
  FOREIGN KEY (app_id) REFERENCES applications(id) on DELETE CASCADE
);

create table events (
  id varchar(100),
  name varchar(100),
  app_id varchar(100),
  primary key (id),
  FOREIGN KEY (app_id) REFERENCES applications(id) on DELETE CASCADE

);

create table documents (
  id varchar(100),
  app_id varchar(100),
  title varchar(100),
  format varchar (100),
  data varchar (100),
  primary key (id),
  FOREIGN KEY (app_id) REFERENCES applications(id) on DELETE CASCADE

);

CREATE INDEX  apps_apis ON apis (app_id);
CREATE INDEX  apps_events ON events (app_id);
CREATE INDEX  apps_documents ON documents (app_id);

create table custom (
  id varchar(100),
  data jsonb
);

create table labels (
  id varchar(100),
  tenant varchar(100),
  app_id varchar(100),
  runtime_id varchar(100),
  label_key varchar(100),
  label_id varchar(100),
  value JSONB
);

insert into labels(id,tenant,app_id,label_key, label_id,value) values
  ('1','adidas','app-1','scenarios','label-def-1','["aaa","bbb"]'),
  ('2','adidas','app-2','scenarios', 'label-def-1','["bbb","ccc"]'),
  ('3','adidas','app-3', 'abc','label-def-2','{"name": "John", "age": 32}'),
  ('4','adidas','app-4', 'abc','label-def-2','{"name": "Pamela", "age": 48}');
