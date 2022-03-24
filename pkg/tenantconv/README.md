# Org ID column populator library

This package contains a library for populating an org-id column based
on the contents of an account number column.

Basic algorithm:
- get a chunk of account numbers from the database table where there is not a corresponding org id
- pass the chunk of account numbers to the translator client for mapping to org ids
- update the rows in the database setting org id column

## Installation

```
go get github.com/RedHatInsights/tenant-utils/pkg/tenantconv
```

## Initialization

This library depends the batch operations of the [Tenant ID translator client](https://pkg.go.dev/github.com/RedHatInsights/tenant-utils/pkg/tenantid)
to map an account number to an org-id.

The caller is responsible for opening the connection to the database, initializing the
BatchTranslator (from the tenantid package), initializing the metrics recorder and
passing those objects to the library.

The library allows for different implementations of the [BatchTranslator](../tenantid/types.go)
and [MetricsRecorder](../tenantconv/metrics.go) to be passed in.

A `TestBatchTranslator` implementation of the BatchTranslator interface has been included for testing purposes.
This implementation simply takes each account number and sets the corresponding org id to be the
account number + "-test".  So account number "1234" is mapped to org id "1234-test".  This allows me to quickly verify that
the util is setting the org id as expected in the database.  To use a real instance of the BatchTranslator,
set the address of the translation service using the `--ean-translator-addr` command line option.

A `TestMetricsRecorder` implementation of the MetricsRecorder function has been included for testing purposes.
The `TestMetricsRecorder` implementation simply prints out the metrics.

## Using the client

```go
import (
  "github.com/RedHatInsights/tenant-utils/pkg/tenantconv"
)

 var db *sql.DB = // initialize db connection
 var table string = "ima_table_name"
 var accountColumn string = "ima_account_column_name"
 var orgIdColumn string = "ima_org_id_column_name"
 var nullOrgIdPlaceholder string = "unknown"
 var dbOperationTimeout int = 5
 var batchSize int = 50
 var batchTranslator tenantid.BatchTranslator = // initialize batch translator
 var recordMetrics tenantconv.MetricsRecorder = // initialize metrics recorder
 var logger *logrus.Logger = // initialize logrus logger

 err := tenantconv.MapAccountToOrgId(context.Background(), db, table, accountColumn, orgIdColumn, nullOrgIdPlaceholder, dbOperationTimeout, batchSize, batchTranslator, recordMetrics, logger)
```
