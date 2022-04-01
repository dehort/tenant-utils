package tenantid

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func newHistogram() *prometheus.HistogramVec {
	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "tenant_translator_request_duration_seconds",
		Help: "Translator service request duration",
	}, []string{"operation", "result"})

	for _, operation := range []string{"eans_to_org_ids", "org_ids_to_eans"} {
		for _, result := range []string{"200", "error"} {
			histogram.WithLabelValues(operation, result)
		}
	}

	return histogram
}

type measuringHttpRequestDoer struct {
	delegate HttpRequestDoer
	observer *prometheus.HistogramVec
}

func (this *measuringHttpRequestDoer) Do(req *http.Request) (resp *http.Response, err error) {
	t := time.Now()
	resp, err = this.delegate.Do(req)
	d := time.Since(t)

	result := "error"
	if err == nil {
		result = fmt.Sprintf("%d", resp.StatusCode)
	}

	op := operation(req.Context())
	this.observer.WithLabelValues(op, result).Observe(d.Seconds())

	return
}
