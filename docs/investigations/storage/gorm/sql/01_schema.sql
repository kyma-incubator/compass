create database compass;

\connect compass

create table applications (
  id varchar(100),
  tenant varchar(100),
--   name varchar(100)
  name varchar(100),
  description varchar (100),
  labels JSON,
  primary key (id)
);


create table apis (
  id varchar(100),
  target_url varchar(100),
--   app_id varchar(100)
  app_id varchar(100),
  primary key (id),
  FOREIGN KEY (app_id) REFERENCES applications(id) on DELETE CASCADE
);

create table events (
  id varchar(100),
  name varchar(100),
--   app_id varchar(100)
  app_id varchar(100),
  primary key (id),
  FOREIGN KEY (app_id) REFERENCES applications(id) on DELETE CASCADE

);

create table documents (
  id varchar(100),
--   app_id varchar(100)
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