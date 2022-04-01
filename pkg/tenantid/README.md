# Tenant ID translator client

This package contains a client library for the tenant translator service.
The tenant translator service provides operations for conversion between different tenant identifiers.

## Installation

```
go get github.com/RedHatInsights/tenant-utils/pkg/tenantid
```

## Client initialization

Initialize the translator with default settings:

```go
translator := tenantid.NewTranslator("https://apicast.svc.cluster.local:8892")
```

Note that in Red Hat ConsoleDot environments the `TENANT_TRANSLATOR_HOST` and `TENANT_TRANSLATOR_PORT` parameters are configured.
These should be used to initialize the tenant translator service client, i.e. `"http://${TENANT_TRANSLATOR_HOST}:${TENANT_TRANSLATOR_PORT}"`

Additional configuration options can be provided to customize the tenant translator service client.

### Custom timeout

A custom timeout value can be specified with:

```go
translator := tenantid.NewTranslator(
	"https://apicast.svc.cluster.local:8892",
	tenantid.WithTimeout(15*time.Second),
)
```

### Collecting metrics

Collection of metrics can be enabled with:

```go
translator := tenantid.NewTranslator(
	"https://apicast.svc.cluster.local:8892",
	tenantid.WithMetrics(),
)
```

This will cause a new Prometheus histogram named `tenant_translator_request_duration_seconds` to be registered with the default registry.
The histogram uses the default bucket configuration and labels are used to distinguish operations and result (status code).

Alternatively, the metrics can be registered with a given registry:

```go
translator := tenantid.NewTranslator(
	"https://apicast.svc.cluster.local:8892",
	tenantid.WithMetricsWithCustomRegisterer(registry),
)
```

### Custom HTTP client

An entirely custom HTTP client can be provided also:

```go
translator := tenantid.NewTranslator(
	"https://apicast.svc.cluster.local:8892",
	tenantid.WithDoer(customClient),
)
```

## Using the client

Translate org_id to account number:

```go
orgID, err := translator.OrgIDToEAN(context.Background(), "901578")
```

Note that the returned account number may be nil in case when the given tenant does not define an account number.

Translate account number to org_id:

```go
account, err := translator.EANToOrgID(context.Background(), "5318290")
```

Batch version of these operations are also provided.
See [API documentation](https://pkg.go.dev/github.com/RedHatInsights/tenant-utils/pkg/tenantid) for more details.

## Using the mock implementation

In addition to the HTTP client, a mock implementation is provided.
This implementation is useful for test execution, local development, etc.

```go
mappings := map[string]*string{
	"5318290":  stringRef("901578"),
	"654321":   nil,
}

translator := tenantid.NewTranslatorMockWithMapping(mappings)
```
