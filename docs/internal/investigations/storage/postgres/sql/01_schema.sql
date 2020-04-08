create database compass;

\connect compass

create table applications (
  id int,
  tenant varchar(100),
--   name varchar(100)
  name varchar(100),
  primary key (id)
);


create table apis (
  id int,
  name varchar(100),
--   app_id int
  app_id int,
  primary key (id),
  FOREIGN KEY (app_id) REFERENCES applications(id)
);

create table events (
  id int,
  name varchar(100),
--   app_id int
  app_id int,
  primary key (id),
  FOREIGN KEY (app_id) REFERENCES applications(id)

);

create table documents (
  id int,
  name varchar(100),
--   app_id int
  app_id int,
  primary key (id),
  FOREIGN KEY (app_id) REFERENCES applications(id)

);

CREATE INDEX  apps_apis ON apis (app_id);
CREATE INDEX  apps_events ON events (app_id);
CREATE INDEX  apps_documents ON documents (app_id);

create table custom (
  id int,
  data jsonb
);