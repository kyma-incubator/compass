-- insert into compass.public.labels (tenant_id, app_id, key, value)
-- VALUES (
--            (SELECT tenant_id, id from applications where id not in (select distinct app_id from labels where key = 'managed')),
--             'managed', false
--     );

-- (SELECT tenant_id, id from applications where id not in (select distinct app_id from labels where key = 'managed'));



insert into labels (id, tenant_id, app_id, key, value)
    (SELECT uuid_generate_v4() as id, tenant_id, id as app_id, 'managed' as key, 'false' as value
    from applications
    where id not in (select distinct app_id from labels where key = 'managed'));