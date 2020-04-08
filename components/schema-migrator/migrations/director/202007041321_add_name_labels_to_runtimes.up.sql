CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

INSERT INTO labels
    (id ,tenant_id, runtime_id, "key", "value")
SELECT uuid_generate_v4(), unnamed.tenant_id, unnamed.runtime_id, 'name', to_json(unnamed.runtime_name)
FROM (
    SELECT r.id as runtime_id, r.name as runtime_name, r.tenant_id as tenant_id
    FROM runtimes r 
    LEFT OUTER JOIN labels l on l.runtime_id = r.id AND l.key = 'name'
    WHERE "value" IS NULL
) as unnamed; 