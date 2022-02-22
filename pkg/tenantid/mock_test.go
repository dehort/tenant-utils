package tenantid

import (
	"context"
	"testing"
)

func TestOrgIDToEANLookup(t *testing.T) {
	tests := []struct {
		name          string
		orgID         string
		expectedEAN   *string
		expectedError bool
	}{
		{
			name:        "successful",
			orgID:       "5318290",
			expectedEAN: stringRef("901578"),
		},
		{
			name:        "successful anemic",
			orgID:       "654321",
			expectedEAN: nil,
		},
		{
			name:          "missing",
			orgID:         "654322",
			expectedEAN:   nil,
			expectedError: true,
		},
	}

	translator := NewTranslatorMock()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ean, err := translator.OrgIDToEAN(context.Background(), test.orgID)

			if err != nil != test.expectedError {
				t.Fatal(err)
			}

			if test.expectedEAN != nil && (ean == nil || *ean != *test.expectedEAN) {
				t.Errorf("expected %p to equal %s", ean, *test.expectedEAN)
			}

			if test.expectedEAN == nil && ean != nil {
				t.Errorf("expected %s to be nil", *ean)
			}
		})
	}
}

func TestEANToOrgIDLookup(t *testing.T) {
	tests := []struct {
		name          string
		ean           string
		expectedOrgID string
		expectedError bool
	}{
		{
			name:          "successful",
			ean:           "901578",
			expectedOrgID: "5318290",
		},
		{
			name:          "missing",
			ean:           "654322",
			expectedError: true,
		},
	}

	translator := NewTranslatorMock()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			orgID, err := translator.EANToOrgID(context.Background(), test.ean)

			if err != nil != test.expectedError {
				t.Fatal(err)
			}

			if test.expectedOrgID != orgID {
				t.Errorf("expected %s to equal %s", orgID, test.expectedOrgID)
			}
		})
	}
}
