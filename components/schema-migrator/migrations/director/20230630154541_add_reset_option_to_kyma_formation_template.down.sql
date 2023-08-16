BEGIN;

DELETE FROM formation_constraints WHERE name = 'DoNotSendNotificationsToKymaAdapterInAFormationTypeExceptForServiceCloudVersion2';

UPDATE formation_templates SET supports_reset = FALSE WHERE name = 'Side-by-Side Extensibility with Kyma';

COMMIT;