BEGIN;

UPDATE formation_templates SET supports_reset = FALSE WHERE name = 'Side-by-Side Extensibility with Kyma';

COMMIT;