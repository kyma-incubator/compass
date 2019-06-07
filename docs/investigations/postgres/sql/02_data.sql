\connect compass
COPY applications(id,tenant,name) from '/docker-entrypoint-initdb.d/03_app.csv' DELIMITER ',' CSV;
COPY apis(id,name,app_id) from '/docker-entrypoint-initdb.d/04_api.csv' DELIMITER ',' CSV;
COPY events(id,name,app_id) from '/docker-entrypoint-initdb.d/05_ev.csv' DELIMITER ',' CSV;
COPY documents(id,name,app_id) from '/docker-entrypoint-initdb.d/06_doc.csv' DELIMITER ',' CSV;

insert into custom(id,f1,f2) values(1,'{"name":"Adam", "age":33}','[{"age":44, "name":"John"},{"age":55, "name":"Margaret"}]');
insert into custom(id,f1,f2) values(2,'{"name":"Tom", "age":44}','[{"age":44, "name":"John"},{"age":55, "name":"Margaret"}]');