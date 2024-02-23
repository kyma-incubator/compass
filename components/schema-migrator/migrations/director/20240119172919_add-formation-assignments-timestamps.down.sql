BEGIN;

ALTER table formations
    DROP last_state_change_timestamp,
    DROP last_notification_sent_timestamp;

ALTER table formation_assignments
    DROP last_state_change_timestamp,
    DROP last_notification_sent_timestamp;

COMMIT;
