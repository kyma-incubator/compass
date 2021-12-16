BEGIN;

-- delete all applications that have no tenant having access to them - orphaned data
DELETE FROM applications WHERE id NOT IN (SELECT id FROM tenant_applications);

-- delete all runtimes that have no tenant having access to them - orphaned data
DELETE FROM runtimes WHERE id NOT IN (SELECT id FROM tenant_runtimes);

COMMIT;
