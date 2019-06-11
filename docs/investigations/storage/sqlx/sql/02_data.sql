\connect compass
-- COPY applications(id,tenant,name) from '/docker-entrypoint-initdb.d/03_app.csv' DELIMITER ',' CSV;
-- COPY apis(id,name,app_id) from '/docker-entrypoint-initdb.d/04_api.csv' DELIMITER ',' CSV;
-- COPY events(id,name,app_id) from '/docker-entrypoint-initdb.d/05_ev.csv' DELIMITER ',' CSV;
-- COPY documents(id,name,app_id) from '/docker-entrypoint-initdb.d/06_doc.csv' DELIMITER ',' CSV;
--
-- insert into custom(id,data) values(1,'{"name":"John", "age":33}');
-- insert into custom(id,data) values(2,'{"name":"Tom", "age":44}');