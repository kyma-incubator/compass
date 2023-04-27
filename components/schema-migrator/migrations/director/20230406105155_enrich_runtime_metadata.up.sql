BEGIN;

ALTER TABLE runtimes
ADD COLUMN application_namespace VARCHAR(256);

-- add application_namespace to existing kyma runtimes
UPDATE runtimes
SET application_namespace = 'sap.kyma'
WHERE id IN (SELECT runtime_id FROM labels WHERE key = 'runtimeType' AND value ? 'kyma');


-- add region label to existing kyma runtimes based on the subaccount they are registered in (global_subaccount_id label)
INSERT INTO labels (id, runtime_id, key, value)
SELECT uuid_generate_v4(),
       res.runtime_id,
       'region' AS key,
       res.value
FROM
    (SELECT t2.runtime_id,
            l2.value
     FROM (
              (SELECT uuid_or_null(value ->> 0) AS tnt,
                      runtime_id
               FROM labels
               WHERE KEY='global_subaccount_id') l1 -- get external subaccout_id and the runtime_id for all runtimes that have 'global_subaccount_id' label (kyma ones).
                  JOIN business_tenant_mappings t ON t.external_tenant = l1.tnt::text) AS t2  -- then join the business_mapping_tenants so that you have the internal subaccount id.
              JOIN labels l2 ON l2.tenant_id = t2.id WHERE KEY = 'region') res; -- finally, join labels by tenant_id and key = 'region' so that you can extract the region from the subaccount

COMMIT;