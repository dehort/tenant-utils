package tenantid

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mock batch translator tests", func() {
	Describe("EAN to org_id", func() {
		It("translates multiple EANs", func() {
			translator := NewTranslatorMock()
			results, err := translator.EANsToOrgIDs(context.Background(), []string{"901578", "6377882", "123456"})
			Expect(err).ToNot(HaveOccurred())

			Expect(results).To(HaveLen(3))

			Expect(*results[0].EAN).To(Equal("901578"))
			Expect(results[0].OrgID).To(Equal("5318290"))
			Expect(results[0].Err).To(BeNil())

			Expect(*results[1].EAN).To(Equal("6377882"))
			Expect(results[1].OrgID).To(Equal("12900172"))
			Expect(results[1].Err).To(BeNil())

			Expect(*results[2].EAN).To(Equal("123456"))
			Expect(results[2].OrgID).To(Equal(""))
			Expect(results[2].Err).To(HaveOccurred())
		})
	})

	Describe("org_id to EAN", func() {
		It("translates multiple org_ids", func() {
			translator := NewTranslatorMock()
			results, err := translator.OrgIDsToEANs(context.Background(), []string{"5318290", "654321", "123456"})
			Expect(err).ToNot(HaveOccurred())

			Expect(results).To(HaveLen(3))

			Expect(results[0].OrgID).To(Equal("5318290"))
			Expect(*results[0].EAN).To(Equal("901578"))
			Expect(results[0].Err).To(BeNil())

			Expect(results[1].OrgID).To(Equal("654321"))
			Expect(results[1].EAN).To(BeNil())
			Expect(results[1].Err).To(BeNil())

			Expect(results[2].OrgID).To(Equal("123456"))
			Expect(results[2].EAN).To(BeNil())
			Expect(results[2].Err).To(BeNil())
		})
	})
})
