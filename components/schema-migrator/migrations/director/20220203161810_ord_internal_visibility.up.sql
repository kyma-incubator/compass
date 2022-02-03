BEGIN;

ALTER TABLE bundle_references
ADD COLUMN visibility varchar;

UPDATE bundle_references br
SET visibility = (SELECT visibility FROM api_definitions WHERE api_def_id = api_definitions.id)
WHERE br.api_def_id IS NOT NULL;

UPDATE bundle_references br
SET visibility = (SELECT visibility FROM event_api_definitions WHERE event_def_id = event_api_definitions.id)
WHERE br.event_def_id IS NOT NULL;

-- all null records from APIs/Events mean that those APIs/Events have come from graphql flow - meaning that they are public, because the concept of private visibility is only ORD scoped
UPDATE bundle_references
SET visibility = 'public'
WHERE visibility IS NULL;

ALTER TABLE bundle_references
ALTER COLUMN visibility SET NOT NULL;

COMMIT;
