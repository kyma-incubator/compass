BEGIN;

-- delete duplicated entries(tenant_id, name) from formations table
DELETE FROM formations
WHERE id IN (SELECT id FROM (SELECT id, row_number() over (partition by tenant_id, name order by id) AS row_num FROM formations) t WHERE t.row_num > 1);

COMMIT;
