# Tenant utils

This module contains utilities for working with tenant identifiers (such as org_id and account number) on console.redhat.com platform.

## Tenant ID translator client

Package `github.com/RedHatInsights/tenant-utils/pkg/tenantid` contains a client library for the tenant translator service.
The tenant translator service provides operations for conversion between different tenant identifiers.

Initiate a translator with default settings:

```go
translator := tenantid.NewTranslator("https://apicast.svc.cluster.local:8892")
```

Initiate a translator with custom settings:

```go
translator := tenantid.NewTranslator(
	"https://apicast.svc.cluster.local:8892",
	tenantid.WithTimeout(15*time.Second),
	tenantid.WithMetrics(),
)
```

Translate org_id to account number:

```go
orgID, err := translator.OrgIDToEAN(context.Background(), "901578")
```

Translate account number to org_id:

```go
account, err := translator.EANToOrgID(context.Background(), "5318290")
```

Note that the returned account number may be nil in case when the given tenant does not define an account number.


### Using the mock implementation

In addition to the HTTP client, a mock implementation is provided.
This implementation is useful for test execution, local development, etc.

```go
mappings := map[string]*string{
	"5318290":  stringRef("901578"),
	"654321":   nil,
}

translator := tenantid.NewTranslatorMockWithMapping(mappings)
```
