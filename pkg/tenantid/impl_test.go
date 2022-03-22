package tenantid

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Batch translator tests", func() {
	Describe("EAN to org_id", func() {
		translate := func(status int, body string, eans ...string) (map[string]TranslationResult, error, *mockHttpRequestDoer) {
			doer := mockHttpClient(status, body)
			translator := batchTranslatorImpl{
				client:      doer,
				serviceHost: "https://example",
			}

			results, err := translator.EANsToOrgIDs(context.Background(), eans)
			if err != nil {
				return nil, err, doer
			}

			indexedResults := make(map[string]TranslationResult)
			for _, result := range results {
				indexedResults[*result.EAN] = result
			}

			return indexedResults, nil, doer
		}

		validateRequest := func(req *http.Request, expectedBody string) {
			data, err := ioutil.ReadAll(req.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal(expectedBody))

			Expect(req.URL.String()).To(Equal("https://example/internal/orgIds"))
			Expect(operation(req.Context())).To(Equal("eans_to_org_ids"))
			Expect(req.Header.Get("accept")).To(Equal("application/json"))
			Expect(req.Header.Get("content-type")).To(Equal("application/json"))
		}

		It("translates a single EAN", func() {
			results, err, doer := translate(200, `{"901578": "5318290"}`, "901578")

			Expect(err).ToNot(HaveOccurred())
			Expect(results).To(HaveLen(1))
			result := results["901578"]
			Expect(*result.EAN).To(Equal("901578"))
			Expect(result.OrgID).To(Equal("5318290"))
			Expect(result.Err).To(BeNil())

			validateRequest(doer.Request, `["901578"]`)
		})

		It("translates multiple EANs", func() {
			results, err, doer := translate(200, `{"901578": "5318290", "901579": "5318291"}`, "901579", "901578")
			Expect(err).ToNot(HaveOccurred())
			Expect(results).To(HaveLen(2))
			result1 := results["901578"]
			Expect(*result1.EAN).To(Equal("901578"))
			Expect(result1.OrgID).To(Equal("5318290"))
			Expect(result1.Err).To(BeNil())
			result2 := results["901579"]
			Expect(*result2.EAN).To(Equal("901579"))
			Expect(result2.OrgID).To(Equal("5318291"))
			Expect(result2.Err).To(BeNil())

			validateRequest(doer.Request, `["901579","901578"]`)
		})
	})

	Describe("org_id to EAN", func() {
		translate := func(status int, body string, orgIDs ...string) (map[string]TranslationResult, error, *mockHttpRequestDoer) {
			doer := mockHttpClient(status, body)
			translator := batchTranslatorImpl{
				client:      doer,
				serviceHost: "https://example",
			}

			results, err := translator.OrgIDsToEANs(context.Background(), orgIDs)
			if err != nil {
				return nil, err, doer
			}

			indexedResults := make(map[string]TranslationResult)
			for _, result := range results {
				indexedResults[result.OrgID] = result
			}

			return indexedResults, nil, doer
		}

		validateRequest := func(req *http.Request, expectedBody string) {
			data, err := ioutil.ReadAll(req.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal(expectedBody))

			Expect(req.URL.String()).To(Equal("https://example/internal/ebsNumbers"))
			Expect(operation(req.Context())).To(Equal("org_ids_to_eans"))
			Expect(req.Header.Get("accept")).To(Equal("application/json"))
			Expect(req.Header.Get("content-type")).To(Equal("application/json"))
		}

		It("translates a single org_id", func() {
			results, err, doer := translate(200, `{"5318290": "901578"}`, "5318290")

			Expect(err).ToNot(HaveOccurred())
			Expect(results).To(HaveLen(1))
			result := results["5318290"]
			Expect(*result.EAN).To(Equal("901578"))
			Expect(result.OrgID).To(Equal("5318290"))
			Expect(result.Err).To(BeNil())

			validateRequest(doer.Request, `["5318290"]`)
		})

		It("translates multiple org_id", func() {
			results, err, doer := translate(200, `{"5318290": "901578", "5318291": "901579"}`, "5318290", "5318291", "5318292")

			Expect(err).ToNot(HaveOccurred())
			Expect(results).To(HaveLen(3))

			result1 := results["5318290"]
			Expect(*result1.EAN).To(Equal("901578"))
			Expect(result1.OrgID).To(Equal("5318290"))
			Expect(result1.Err).To(BeNil())

			result2 := results["5318291"]
			Expect(*result2.EAN).To(Equal("901579"))
			Expect(result2.OrgID).To(Equal("5318291"))
			Expect(result2.Err).To(BeNil())

			result3 := results["5318292"]
			Expect(result3.EAN).To(BeNil())
			Expect(result3.OrgID).To(Equal("5318292"))
			Expect(result3.Err).To(BeNil())

			validateRequest(doer.Request, `["5318290","5318291","5318292"]`)
		})
	})

})

func mockHttpClient(statusCode int, body string) *mockHttpRequestDoer {
	response := http.Response{
		StatusCode: statusCode,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(body))),
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
	}

	return &mockHttpRequestDoer{
		response: &response,
	}
}

type mockHttpRequestDoer struct {
	Request  *http.Request
	response *http.Response
}

func (this *mockHttpRequestDoer) Do(req *http.Request) (*http.Response, error) {
	this.Request = req
	return this.response, nil
}
