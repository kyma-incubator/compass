BEGIN;

ALTER TABLE formation_assignments
    DROP COLUMN last_operation_initiator,
    DROP COLUMN last_operation_initiator_type,
    DROP COLUMN last_operation;

INSERT INTO formation_assignments (id, formation_id, tenant_id, source, source_type, target, target_type, state)
SELECT uuid_generate_v4(), fa.formation_id, fa.tenant_id, fa.source, fa.source_type, fa.source, fa.source_type, 'READY'
FROM (
         (
             SELECT DISTINCT fa_source.formation_id, fa_source.tenant_id, fa_source.source, fa_source.source_type
             FROM formation_assignments AS fa_source
                      INNER JOIN formation_assignments AS fa_target
                                 ON fa_source.source = fa_target.target AND fa_source.source_type = fa_target.target_type AND fa_source.target = fa_target.source AND fa_source.target_type = fa_target.source_type
         )
     ) AS fa;

COMMIT;
