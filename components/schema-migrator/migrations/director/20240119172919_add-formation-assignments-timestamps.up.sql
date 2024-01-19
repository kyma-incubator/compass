BEGIN;

ALTER table formations
    ADD COLUMN last_state_change_timestamp TIMESTAMP,
    ADD COLUMN last_notification_sent_timestamp TIMESTAMP;

UPDATE formations SET last_state_change_timestamp = now()
    WHERE formations.last_state_change_timestamp IS NULL;

UPDATE formations SET last_notification_sent_timestamp = NULL;

ALTER table formation_assignments
    ADD COLUMN last_state_change_timestamp TIMESTAMP,
    ADD COLUMN last_notification_sent_timestamp TIMESTAMP;

UPDATE formation_assignments SET last_state_change_timestamp = now()
    WHERE formation_assignments.last_state_change_timestamp IS NULL;

UPDATE formation_assignments SET last_notification_sent_timestamp = NULL;

COMMIT;
