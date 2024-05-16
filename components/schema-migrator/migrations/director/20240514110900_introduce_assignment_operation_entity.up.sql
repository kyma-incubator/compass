BEGIN;

CREATE TABLE assignment_operations
(
    id                      UUID PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    type                    TEXT NOT NULL CHECK (type IN ('ASSIGN', 'UNASSIGN')),
    formation_assignment_id UUID NOT NULL CHECK (id <> '00000000-0000-0000-0000-000000000000') REFERENCES formation_assignments (id) ON DELETE CASCADE,
    formation_id            UUID NOT NULL CHECK (id <> '00000000-0000-0000-0000-000000000000') REFERENCES formations (id),
    triggered_by            TEXT NOT NULL CHECK (triggered_by IN ('ASSIGN_OBJECT', 'UNASSIGN_OBJECT', 'RESET', 'RESYNC')),
    started_at_timestamp    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    finished_at_timestamp   TIMESTAMP
);

CREATE INDEX assignment_operations_formation_assignment_id_idx
    ON assignment_operations (formation_assignment_id);

CREATE INDEX assignment_operations_formation_id_idx
    ON assignment_operations (formation_id);

CREATE INDEX assignment_operations_type_idx
    ON assignment_operations (type);

INSERT INTO assignment_operations (id, type, formation_assignment_id, formation_id, triggered_by, started_at_timestamp)
SELECT uuid_generate_v4(), 'ASSIGN', fa.id, fa.formation_id, 'ASSIGN_OBJECT', CURRENT_TIMESTAMP
FROM formation_assignments fa
WHERE fa.state IN ('INITIAL', 'CREATE_ERROR', 'CONFIG_PENDING');

INSERT INTO assignment_operations (id, type, formation_assignment_id, formation_id, triggered_by, started_at_timestamp,
                                   finished_at_timestamp)
SELECT uuid_generate_v4(), 'ASSIGN', fa.id, fa.formation_id, 'ASSIGN_OBJECT', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM formation_assignments fa
WHERE fa.state IN ('READY');

INSERT INTO assignment_operations (id, type, formation_assignment_id, formation_id, triggered_by, started_at_timestamp)
SELECT uuid_generate_v4(), 'UNASSIGN', fa.id, fa.formation_id, 'UNASSIGN_OBJECT', CURRENT_TIMESTAMP
FROM formation_assignments fa
WHERE fa.state IN ('DELETING', 'DELETE_ERROR', 'INSTANCE_CREATOR_DELETING', 'INSTANCE_CREATOR_DELETE_ERROR');

COMMIT;
