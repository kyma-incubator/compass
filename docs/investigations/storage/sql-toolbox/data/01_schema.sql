create database compass;

\connect compass


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
