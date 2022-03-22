package tenantid

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("translator tests", func() {

	translator := NewTranslatorMock()

	DescribeTable("EAN to org_id",
		func(ean, expectedOrgID string, shouldErr bool) {
			orgID, err := translator.EANToOrgID(context.Background(), ean)

			if !shouldErr {
				Expect(orgID).To(Equal(expectedOrgID))
				Expect(err).ToNot(HaveOccurred())
			} else {
				Expect(err).To(HaveOccurred())

			}
		},
		Entry("ean1", "901578", "5318290", false),
		Entry("ean2", "6377882", "12900172", false),
		Entry("unknown ean", "123456", "", true),
	)

	DescribeTable("org_id to EAN",
		func(orgID string, expectedEAN *string, shouldErr bool) {
			ean, err := translator.OrgIDToEAN(context.Background(), orgID)

			if !shouldErr {
				Expect(ean).To(Equal(expectedEAN))
				Expect(err).ToNot(HaveOccurred())
			} else {
				Expect(err).To(HaveOccurred())
			}
		},
		Entry("org_id1", "5318290", stringRef("901578"), false),
		Entry("anemic tenant", "654321", nil, false),
		Entry("unknown tenant", "123456", nil, false),
	)
})
