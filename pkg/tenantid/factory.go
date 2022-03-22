package tenantid

import (
	"net/http"
	"sort"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const defaultTimeout = 10

type TranslatorOption interface {
	apply(*batchTranslatorImpl)
	// Options may need to be applied in certain order (e.g. timeout-setting should be applied before an option that wraps the doer)
	// Priority lets us have complete ordering of all options
	// Options are applied in ascending priority (highest priority options go last)
	Priority() int
}

// NewTranslator returns a new translator instance configured with the provided options
func NewTranslator(serviceHost string, options ...TranslatorOption) Translator {
	sort.SliceStable(options, func(i, j int) bool {
		return options[i].Priority() < options[j].Priority()
	})

	batchTranslator := &batchTranslatorImpl{
		serviceHost: serviceHost,
		client: &http.Client{
			Timeout: defaultTimeout * time.Second,
		},
	}

	for _, option := range options {
		option.apply(batchTranslator)
	}

	return &translator{
		BatchTranslator: batchTranslator,
	}
}

type translatorOptionImpl struct {
	fn       func(*batchTranslatorImpl)
	priority int
}

func (this *translatorOptionImpl) apply(impl *batchTranslatorImpl) {
	this.fn(impl)
}

func (this *translatorOptionImpl) Priority() int {
	return this.priority
}

// WithTimeout allows a custom timeout value to be defined.
// The default value of 10 seconds is used otherwise.
func WithTimeout(timeout time.Duration) TranslatorOption {
	return &translatorOptionImpl{
		fn: func(impl *batchTranslatorImpl) {
			impl.client = &http.Client{
				Timeout: timeout,
			}
		},
		priority: 10,
	}
}

// WithDoer allow a custom http.Client to be provided.
func WithDoer(doer HttpRequestDoer) TranslatorOption {
	return &translatorOptionImpl{
		fn: func(impl *batchTranslatorImpl) {
			impl.client = doer
		},
		priority: 20,
	}
}

// WithDoerWrapper allow for the default http.Client to be wrapped by a custom decorator.
func WithDoerWrapper(fn func(HttpRequestDoer) HttpRequestDoer) TranslatorOption {
	return &translatorOptionImpl{
		fn: func(impl *batchTranslatorImpl) {
			impl.client = fn(impl.client)
		},
		priority: 30,
	}
}

// WithMetrics registers a new histogram measuring latency with the default prometheus Registerer
func WithMetrics() TranslatorOption {
	return WithMetricsWithCustomRegisterer(prometheus.DefaultRegisterer)
}

// WithMetricsWithCustomRegisterer registers a new histogram measuring latency with the provided prometheus Registerer
func WithMetricsWithCustomRegisterer(registerer prometheus.Registerer) TranslatorOption {
	return &translatorOptionImpl{
		fn: func(impl *batchTranslatorImpl) {
			observer := newHistogram()
			registerer.MustRegister(observer)

			impl.client = &measuringHttpRequestDoer{
				delegate: impl.client,
				observer: observer,
			}
		},
		priority: 40,
	}
}
