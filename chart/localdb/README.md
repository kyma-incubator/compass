# LocalDB

## Overview

LocalDB is used only in local env scenario to install and apply the DB Dump if needed. LocalDB consists of the following sub-charts:
- `postgresql` - installing the PostgreSQL database
- `dbdump` - applying the DB dump

## Details

### Configuration

LocalDB has a standard Helm chart configuration. You can check all available configurations in the chart, and sub-charts's `values.yaml` files.

The values from those files can be overridden during installation. 
