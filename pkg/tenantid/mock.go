package tenantid

import (
	"context"
	"fmt"
)

type mockTenantIDTranslator struct {
	orgIDToEAN map[string]*string
}

// NewTranslatorMock returns a mock implementation of translator with a predefined mapping.
func NewTranslatorMock() Translator {
	orgIDToEAN := map[string]*string{
		"5318290":  stringRef("901578"),
		"12900172": stringRef("6377882"),
		"14656001": stringRef("7135271"),
		"11789772": stringRef("6089719"),
		"3340851":  stringRef("0369233"),
		"654321":   nil,
	}

	return &mockTenantIDTranslator{
		orgIDToEAN: orgIDToEAN,
	}
}

// NewTranslatorMockWithMapping returns a mock implementation of translator that operates on the given mapping.
func NewTranslatorMockWithMapping(mapping map[string]*string) Translator {
	return &mockTenantIDTranslator{
		orgIDToEAN: mapping,
	}
}

func (this *mockTenantIDTranslator) OrgIDToEAN(ctx context.Context, orgID string) (ean *string, err error) {
	value, ok := this.orgIDToEAN[orgID]

	if ok {
		return value, nil
	}

	return nil, &TenantNotFoundError{msg: fmt.Sprintf("unknown tenant: %s", orgID)}

}

func (this *mockTenantIDTranslator) EANToOrgID(ctx context.Context, value string) (orgID string, err error) {
	for orgID, ean := range this.orgIDToEAN {
		if ean != nil && *ean == value {
			return orgID, nil
		}
	}

	return "", &TenantNotFoundError{msg: fmt.Sprintf("unknown tenant: %s", value)}
}
