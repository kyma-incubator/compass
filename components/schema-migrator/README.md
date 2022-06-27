# Schema Migrator

## Overview

The Schema Migrator is responsible for database schema migrations. It runs all _UP_ migrations one by one as a _post-install/post-upgrade_ K8s job. However, if the Compass installation fails, it must be possible to revert the DB state to what it was before the installation. To do this, a _pre-rollback_ K8s job is responsible for running all new down migrations introduced with the failed installation.

### Rollbacks
The _pre-rollback_ `compass-migration-down` job is rendered from the release, to which you want to roll back. Therefore, the version of the `schema-migrator-down` does not include the new _UP_ and _DOWN_ migrations. For this reason, there is `PersistentVolume` (PV) for preserving the migrations from the _post-install_ `compass-migration` job. For more information about using persistent volume, see [Persistent Volumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/).

The rollback has the following process flow:
1. The `compass-migration` job replaces the existing migrations in the PV with the migrations from its container.
2. `compass-migration` runs the migrations.
3. When the installation fails, the `compass-migration-down` automatically tries to migrate to the latest available version in its container (this is the version from the previous installation). To do this, the `compass-migration-down` performs the down migrations one by one until the desired state is reached.
    - In case of clean installation, it performs the down migrations one by one, until there are no migrations left. As a result, the DB is cleaned from any migrations.
    - If the desired version does not exist in the PV, the `compass-migration-down` job fails.

## Development

If you want to modify the database schema used by compass, add migration files to `migrations` directory. To do this, see [Migrations](https://github.com/golang-migrate/migrate/blob/master/MIGRATIONS.md).

In case you changed the `schema-migrator` component in your pull request, a new image of the migrator will be produced. It contains all migration files, so make sure that you bump the component version value in the compass chart.

`Note:` During local installation, if you specified the `--dump-db` flag then in the local k3d docker registry a new image will be built based on the local files. The new image will be used for schema-migrator component.

To test if migration files are correct, run:
```
make verify
```
