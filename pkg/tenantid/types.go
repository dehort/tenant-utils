// Package tenantid is a client library for the Tenant Translator service. The service provides translation between different tenant identifiers.
package tenantid

import (
	"context"
	"net/http"
)

// Provides translation between tenant identifiers.
// Namely, it converts an org_id to EAN (EBS account number) and vice versa.
// Both single-operation and batch variants are provided
type Translator interface {
	BatchTranslator

	// Converts an EAN (EBS account number) to org_id.
	// Returns TenantNotFoundError (second return value) if the EAN is not known.
	EANToOrgID(ctx context.Context, ean string) (orgId string, err error)

	// Converts an org_id to EAN (EBS account number).
	// Returns nil if the org_id belongs to an anemic tenant or the org_id is not known.
	OrgIDToEAN(ctx context.Context, orgId string) (ean *string, err error)
}

// abstraction of http.Client
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Provides translation between tenant identifiers.
type BatchTranslator interface {

	// Converts a slice of EANs (EBS account number) to org_ids
	EANsToOrgIDs(ctx context.Context, eans []string) (results []TranslationResult, err error)

	// Converts a slice of org_ids to EANs (EBS account number)
	OrgIDsToEANs(ctx context.Context, orgIDs []string) (results []TranslationResult, err error)
}

// Holds the result of tenant identifier translation
type TranslationResult struct {
	OrgID string
	EAN   *string
	Err   error
}
