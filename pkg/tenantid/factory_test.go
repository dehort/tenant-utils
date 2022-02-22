package tenantid

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func mockOption(value int) TranslatorOption {
	return &translatorOptionImpl{
		priority: value,
		fn: func(translator *translatorImpl) {
			translator.serviceHost += fmt.Sprintf("%d", value)
		},
	}
}

func TestOptionOrdering(t *testing.T) {
	options := []TranslatorOption{
		mockOption(3),
		mockOption(2),
		mockOption(-5),
		mockOption(8),
	}

	translator := NewTranslator("", options...)
	actual := translator.(*translatorImpl).serviceHost
	expected := "-5238"

	if expected != actual {
		t.Errorf("expected %v to be %v", actual, expected)
	}
}

func TestTimeout(t *testing.T) {
	timeout := 60 * time.Second
	option := WithTimeout(timeout)

	translator := NewTranslator("", option)
	actual := translator.(*translatorImpl).client.(*http.Client).Timeout

	if actual != timeout {
		t.Errorf("expected %v to be %v", actual, timeout)
	}
}
