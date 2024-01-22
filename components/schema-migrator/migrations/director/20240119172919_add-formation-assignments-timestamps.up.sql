BEGIN;

ALTER TABLE formations
    ADD COLUMN last_state_change_timestamp TIMESTAMP DEFAULT NOW(),
    ADD COLUMN last_notification_sent_timestamp TIMESTAMP;

ALTER TABLE formation_assignments
    ADD COLUMN last_state_change_timestamp TIMESTAMP DEFAULT NOW(),
    ADD COLUMN last_notification_sent_timestamp TIMESTAMP;

COMMIT;
