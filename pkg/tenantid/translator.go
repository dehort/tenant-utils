package tenantid

import "context"

// Implements single-operation methods using BatchTranslator
type translator struct {
	BatchTranslator
}

func (this *translator) EANToOrgID(ctx context.Context, ean string) (string, error) {
	results, err := this.EANsToOrgIDs(ctx, []string{ean})
	if err != nil {
		return "", err
	}

	if results[0].Err != nil {
		return "", results[0].Err
	}

	return results[0].OrgID, nil
}

func (this *translator) OrgIDToEAN(ctx context.Context, orgId string) (ean *string, err error) {
	results, err := this.OrgIDsToEANs(ctx, []string{orgId})

	if err != nil {
		return nil, err
	}

	if results[0].Err != nil {
		return nil, results[0].Err
	}

	return results[0].EAN, nil
}
