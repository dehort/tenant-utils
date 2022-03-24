# Tenant utils

This module contains utilities for working with tenant identifiers (such as org_id and account number) on console.redhat.com platform.

## Tenant ID translator client

Package `github.com/RedHatInsights/tenant-utils/pkg/tenantid` contains a client library for the tenant translator service.
The tenant translator service provides operations for conversion between different tenant identifiers.

See the [package readme](./pkg/tenantid/README.md) for more information.

## Org ID column populator util

The org-id-column-populator is a util for populating an org-id column within a database.

See the [package readme](./cmd/org-id-column-populator/README.md) for more information.

## Org ID column populator library

Package `github.com/RedHatInsights/tenant-utils/pkg/tenantconv` contains a library for populating an org-id column
within a database.

See the [package readme](./pkg/tenantconv/README.md) for more information.
