# Schema Migrator

## Overview

The Schema Migrator is responsible for database schema migrations.

## Development

If you want to modify database schema used by compass, add migration files (follow [this](https://github.com/golang-migrate/migrate/blob/master/MIGRATIONS.md) instructions) to `migrations` directory. 
New image of migrator will be produced that contains all migration files so make sure to bump component version value in compass chart.
To test if migration files are correct, execute:
```
make verify
```
