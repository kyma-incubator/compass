BEGIN;

insert into labels (id, tenant_id, app_id, key, value)
    (
        SELECT uuid_generate_v4(), tenant_id, id, 'managed', '"false"'
        from applications
        where id not in (select distinct app_id from labels where key = 'managed')
    );

COMMIT;
