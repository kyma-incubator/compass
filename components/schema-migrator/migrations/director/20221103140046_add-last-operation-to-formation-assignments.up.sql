
BEGIN;

DROP TYPE webhook_mode;

CREATE TYPE webhook_mode AS ENUM (
    'ASYNC',
    'SYNC',
    'ASYNC_CALLBACK'
);

ALTER TABLE formation_assignments
    ADD COLUMN last_operation_initiator UUID,
    ADD COLUMN last_operation_initiator_type VARCHAR(256),
    ADD COLUMN last_operation VARCHAR(256);

ALTER TABLE formation_assignments
    ADD CONSTRAINT last_operation_initiator_type_check CHECK (last_operation_initiator_type IN ('APPLICATION', 'RUNTIME', 'RUNTIME_CONTEXT')),
    ADD CONSTRAINT last_operation_check CHECK (last_operation IN ('assign', 'unassign'));

ALTER TABLE formation_assignments
    DROP CONSTRAINT formation_assignments_state_check;

ALTER TABLE formation_assignments
    ADD CONSTRAINT formation_assignments_state_check CHECK ( state IN ('INITIAL', 'READY', 'CREATE_ERROR', 'DELETING', 'DELETE_ERROR', 'CONFIG_PENDING'));

UPDATE formation_assignments SET last_operation = 'assign', last_operation_initiator = source, last_operation_initiator_type = source_type;

-- APP to APP
INSERT INTO formation_assignments (id, formation_id, tenant_id, source, source_type, target, target_type, state,
                                   last_operation_initiator, last_operation_initiator_type, last_operation)
SELECT uuid_generate_v4(),
       f.id,
       f.tenant_id,
       t1.id,
       'APPLICATION',
       t2.id,
       'APPLICATION',
       'READY',
       t1.id,
       'APPLICATION',
       'assign'
FROM formations f
         JOIN tenant_applications t1 ON f.tenant_id = t1.tenant_id
         JOIN tenant_applications t2 ON t1.tenant_id = t2.tenant_id AND t1.id != t2.id
         JOIN labels l1 ON l1.app_id = t1.id AND l1.key = 'scenarios' AND l1.value ? f.name
    JOIN labels l2 ON l2.app_id = t2.id AND l2.key = 'scenarios' AND l2.value ? f.name
    ON conflict (formation_id, source, target) do nothing ;



-- APP to RTM_CTX
INSERT INTO formation_assignments (id, formation_id, tenant_id, source, source_type, target, target_type, state,
                                   last_operation_initiator, last_operation_initiator_type, last_operation)
SELECT uuid_generate_v4(),
       f.id,
       f.tenant_id,
       ta.id,
       'APPLICATION',
       trc.id,
       'RUNTIME_CONTEXT',
       'READY',
       ta.id,
       'APPLICATION',
       'assign'
FROM formations f
         JOIN tenant_applications ta ON f.tenant_id = ta.tenant_id
         JOIN tenant_runtime_contexts trc ON ta.tenant_id = trc.tenant_id
         JOIN labels la ON la.app_id = ta.id AND la.key = 'scenarios' AND la.value ? f.name
         JOIN labels lrc ON lrc.runtime_context_id = trc.id AND lrc.key = 'scenarios' AND lrc.value ? f.name
    ON conflict (formation_id, source, target) do nothing ;

-- RTM_CTX to APP
INSERT INTO formation_assignments (id, formation_id, tenant_id, source, source_type, target, target_type, state,
                                   last_operation_initiator, last_operation_initiator_type, last_operation)
SELECT uuid_generate_v4(),
       f.id,
       f.tenant_id,
       trc.id,
       'RUNTIME_CONTEXT',
       ta.id,
       'APPLICATION',
       'READY',
       trc.id,
       'RUNTIME_CONTEXT',
       'assign'
FROM formations f
         JOIN tenant_applications ta ON f.tenant_id = ta.tenant_id
         JOIN tenant_runtime_contexts trc ON ta.tenant_id = trc.tenant_id
         JOIN labels la ON la.app_id = ta.id AND la.key = 'scenarios' AND la.value ? f.name
         JOIN labels lrc ON lrc.runtime_context_id = trc.id AND lrc.key = 'scenarios' AND lrc.value ? f.name
    ON conflict (formation_id, source, target) do nothing ;

-- APP to RTM
INSERT INTO formation_assignments (id, formation_id, tenant_id, source, source_type, target, target_type, state,
                                   last_operation_initiator, last_operation_initiator_type, last_operation)
SELECT uuid_generate_v4(),
       f.id,
       f.tenant_id,
       ta.id,
       'APPLICATION',
       tr.id,
       'RUNTIME',
       'READY',
       ta.id,
       'APPLICATION',
       'assign'
FROM formations f
         JOIN tenant_applications ta ON f.tenant_id = ta.tenant_id
         JOIN tenant_runtimes tr ON ta.tenant_id = tr.tenant_id
         JOIN labels la ON la.app_id = ta.id AND la.key = 'scenarios' AND la.value ? f.name
         JOIN labels lr ON lr.runtime_id = tr.id AND lr.key = 'scenarios' AND lr.value ? f.name
    ON conflict (formation_id, source, target) do nothing ;

-- RTM to APP
INSERT INTO formation_assignments (id, formation_id, tenant_id, source, source_type, target, target_type, state,
                                   last_operation_initiator, last_operation_initiator_type, last_operation)
SELECT uuid_generate_v4(),
       f.id,
       f.tenant_id,
       tr.id,
       'RUNTIME',
       ta.id,
       'APPLICATION',
       'READY',
       tr.id,
       'RUNTIME',
       'assign'
FROM formations f
         JOIN tenant_applications ta ON f.tenant_id = ta.tenant_id
         JOIN tenant_runtimes tr ON ta.tenant_id = tr.tenant_id
         JOIN labels la ON la.app_id = ta.id AND la.key = 'scenarios' AND la.value ? f.name
         JOIN labels lr ON lr.runtime_id = tr.id AND lr.key = 'scenarios' AND lr.value ? f.name
    ON conflict (formation_id, source, target) do nothing ;


-- RTM to RTM
INSERT INTO formation_assignments (id, formation_id, tenant_id, source, source_type, target, target_type, state,
                                   last_operation_initiator, last_operation_initiator_type, last_operation)
SELECT uuid_generate_v4(),
       f.id,
       f.tenant_id,
       t1.id,
       'RUNTIME',
       t2.id,
       'RUNTIME',
       'READY',
       t1.id,
       'RUNTIME',
       'assign'
FROM formations f
         JOIN tenant_runtimes t1 ON f.tenant_id = t1.tenant_id
         JOIN tenant_runtimes t2 ON t1.tenant_id = t2.tenant_id AND t1.id != t2.id
         JOIN labels l1 ON l1.runtime_id = t1.id AND l1.key = 'scenarios' AND l1.value ? f.name
    JOIN labels l2 ON l2.runtime_id = t2.id AND l2.key = 'scenarios' AND l2.value ? f.name
    ON conflict (formation_id, source, target) do nothing ;

-- RTM to RTM_CTX
INSERT INTO formation_assignments (id, formation_id, tenant_id, source, source_type, target, target_type, state,
                                   last_operation_initiator, last_operation_initiator_type, last_operation)
SELECT uuid_generate_v4(),
       f.id,
       f.tenant_id,
       tr.id,
       'RUNTIME',
       trc.id,
       'RUNTIME_CONTEXT',
       'READY',
       tr.id,
       'RUNTIME',
       'assign'
FROM formations f
         JOIN tenant_runtimes tr ON f.tenant_id = tr.tenant_id
         JOIN tenant_runtime_contexts trc ON trc.tenant_id = tr.tenant_id
         JOIN labels lr ON lr.runtime_id = tr.id AND lr.key = 'scenarios' AND lr.value ? f.name
         JOIN labels lrc ON lrc.runtime_context_id = trc.id AND lrc.key = 'scenarios' AND lrc.value ? f.name
    ON conflict (formation_id, source, target) do nothing ;

-- RTM_CTX to RTM
INSERT INTO formation_assignments (id, formation_id, tenant_id, source, source_type, target, target_type, state,
                                   last_operation_initiator, last_operation_initiator_type, last_operation)
SELECT uuid_generate_v4(),
       f.id,
       f.tenant_id,
       trc.id,
       'RUNTIME_CONTEXT',
       tr.id,
       'RUNTIME',
       'READY',
       trc.id,
       'RUNTIME_CONTEXT',
       'assign'
FROM formations f
         JOIN tenant_runtimes tr ON f.tenant_id = tr.tenant_id
         JOIN tenant_runtime_contexts trc ON trc.tenant_id = tr.tenant_id
         JOIN labels lr ON lr.runtime_id = tr.id AND lr.key = 'scenarios' AND lr.value ? f.name
         JOIN labels lrc ON lrc.runtime_context_id = trc.id AND lrc.key = 'scenarios' AND lrc.value ? f.name
    ON conflict (formation_id, source, target) do nothing ;


-- RTM_CTX to RTM_CTX
INSERT INTO formation_assignments (id, formation_id, tenant_id, source, source_type, target, target_type, state,
                                   last_operation_initiator, last_operation_initiator_type, last_operation)
SELECT uuid_generate_v4(),
       f.id,
       f.tenant_id,
       t1.id,
       'RUNTIME_CONTEXT',
       t2.id,
       'RUNTIME_CONTEXT',
       'READY',
       t1.id,
       'RUNTIME_CONTEXT',
       'assign'
FROM formations f
         JOIN tenant_runtime_contexts t1 ON f.tenant_id = t1.tenant_id
         JOIN tenant_runtime_contexts t2 ON t1.tenant_id = t2.tenant_id AND t1.id != t2.id
         JOIN labels l1 ON l1.runtime_context_id = t1.id AND l1.key = 'scenarios' AND l1.value ? f.name
    JOIN labels l2 ON l2.runtime_context_id = t2.id AND l2.key = 'scenarios' AND l2.value ? f.name
    ON conflict (formation_id, source, target) do nothing ;

COMMIT;
