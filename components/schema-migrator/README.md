# Schema Migrator

## Overview

The Schema Migrator is responsible for database schema migrations.

It runs all _UP_ migrations as a _post-install/post-upgrade_ K8s Job.

If the Compass installation fails, we should be able to revert the DB state to what it was before the installation. A _pre-rollback_ K8s Job is responsible for executing all new down migrations from the failed installation.

### Rollbacks
The _pre-rollback_ `compass-migration-down` Job is rendered from the target release - the one we’re rolling back to - hence, the version of the `schema-migrator-down` does not include the new _UP_ and _DOWN_ migrations. That's why we have a `PersistentVolume` for keeping the migrations from the _post-install_ `compass-migration` job.

The rollback flow is the following:
1. The `compass-migration` Job replaces the existing migrations in the PV with the migrations from its container.
2. `compass-migration` executes the migrations
3. The installation fails
4. `compass-migration-down` will try to migrate to the latest version present in its container - that's the version from the previous installation.
    - If the version does not exist in the PV, the `compass-migration-down` Job fails.
    - In case of clean installation, it will migrate down, until there are no migrations left, leaving a clean DB.

## Development

If you want to modify database schema used by compass, add migration files (follow [this](https://github.com/golang-migrate/migrate/blob/master/MIGRATIONS.md) instructions) to `migrations` directory. 
New image of migrator will be produced that contains all migration files so make sure to bump component version value in compass chart.
To test if migration files are correct, execute:
```
make verify
```
