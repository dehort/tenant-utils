package tenantid

import (
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func mockOption(value int) TranslatorOption {
	return &translatorOptionImpl{
		priority: value,
		fn: func(translator *batchTranslatorImpl) {
			translator.serviceHost += fmt.Sprintf("%d", value)
		},
	}
}

var _ = Describe("Factory tests", func() {
	It("sets timeout", func() {
		timeout := 60 * time.Second
		option := WithTimeout(timeout)

		t := NewTranslator("", option)
		actual := t.(*translator).BatchTranslator.(*batchTranslatorImpl).client.(*http.Client).Timeout

		Expect(actual).To(Equal(timeout))
	})

	It("applies the options in the given order", func() {
		options := []TranslatorOption{
			mockOption(3),
			mockOption(2),
			mockOption(-5),
			mockOption(8),
		}

		t := NewTranslator("", options...)
		actual := t.(*translator).BatchTranslator.(*batchTranslatorImpl).serviceHost

		Expect(actual).To(Equal("-5238"))
	})
})
