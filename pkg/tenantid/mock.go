package tenantid

import (
	"context"
	"fmt"
)

type mockBatchTranslator struct {
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
		"10001":    stringRef("010101"),
	}

	return NewTranslatorMockWithMapping(orgIDToEAN)
}

// NewTranslatorMockWithMapping returns a mock implementation of translator that operates on the given mapping.
func NewTranslatorMockWithMapping(mapping map[string]*string) Translator {
	return &translator{
		BatchTranslator: &mockBatchTranslator{
			orgIDToEAN: mapping,
		},
	}
}

func (this *mockBatchTranslator) OrgIDsToEANs(ctx context.Context, orgIDs []string) (results []TranslationResult, err error) {
	results = make([]TranslationResult, len(orgIDs))

	for i, orgID := range orgIDs {
		value := this.orgIDToEAN[orgID]
		results[i] = newTranslationResult(orgID, value)
	}

	return
}

func (this *mockBatchTranslator) EANsToOrgIDs(ctx context.Context, eans []string) (results []TranslationResult, err error) {
	results = make([]TranslationResult, len(eans))

	for i, requestedEAN := range eans {
		results[i] = this.findByEAN(requestedEAN)
	}

	return
}

func (this *mockBatchTranslator) findByEAN(requestedEAN string) TranslationResult {
	for orgID, ean := range this.orgIDToEAN {
		if ean != nil && *ean == requestedEAN {
			return newTranslationResult(orgID, ean)
		}
	}

	return TranslationResult{
		EAN: stringRef(requestedEAN),
		Err: &TenantNotFoundError{msg: fmt.Sprintf("unknown tenant: %s", requestedEAN)},
	}
}
