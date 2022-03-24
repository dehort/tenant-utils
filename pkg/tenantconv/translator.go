package tenantconv

import (
	"context"

	"github.com/RedHatInsights/tenant-utils/pkg/tenantid"
)

type TestBatchTranslator struct {
}

func (this *TestBatchTranslator) EANsToOrgIDs(ctx context.Context, eans []string) ([]tenantid.TranslationResult, error) {
	results := make([]tenantid.TranslationResult, 0, len(eans))
	for _, ean := range eans {
		newEAN := new(string)
		*newEAN = ean
		r := tenantid.TranslationResult{OrgID: ean + "-test",
			EAN: newEAN,
			Err: nil,
		}
		results = append(results, r)
	}
	return results, nil
}

func (this *TestBatchTranslator) OrgIDsToEANs(ctx context.Context, orgIDs []string) ([]tenantid.TranslationResult, error) {
	return nil, nil
}
