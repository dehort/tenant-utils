package tenantconv

import (
	"github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

type MetricsRecorder func(tableName string, uniqueAccounts int64, rowsUpdated int64) error

func TestMetricsRecorder(logger *logrus.Logger) (MetricsRecorder, error) {
	return func(tableName string, uniqueAccounts int64, rowsUpdated int64) error {
		logger.Println("Discarding metrics:")
		logger.Println("tableName: ", tableName)
		logger.Println("uniqueAccounts:", uniqueAccounts)
		logger.Println("rowsUpdated:", rowsUpdated)
		return nil
	}, nil
}

func PrometheusMetricsRecorder(logger *logrus.Logger, pushGateway string, jobName string, registry *prometheus.Registry) (MetricsRecorder, error) {
	return func(tableName string, uniqueAccounts int64, rowsUpdated int64) error {
		logger.Println("Metrics:")
		logger.Println("pushGateway: ", pushGateway)
		logger.Println("tableName: ", tableName)
		logger.Println("uniqueAccounts:", uniqueAccounts)
		logger.Println("rowsUpdated:", rowsUpdated)

		uniqueAccountsMetric := prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "org_id_column_populator_unique_accounts",
			Help: "Number unique accounts processed per table.",
		}, []string{"table"})
		rowsUpdatedMetric := prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "org_id_column_populator_rows_updated",
			Help: "Number rows updated per table.",
		}, []string{"table"})

		registry.MustRegister(uniqueAccountsMetric)
		registry.MustRegister(rowsUpdatedMetric)

		uniqueAccountsMetric.With(prometheus.Labels{"table": tableName}).Add(float64(uniqueAccounts))
		rowsUpdatedMetric.With(prometheus.Labels{"table": tableName}).Add(float64(rowsUpdated))

		err := push.New(pushGateway, jobName).
			Gatherer(registry).
			Push()
		if err != nil {
			logger.Println("Unable to push metrics to prometheus gateway:", err)
			return err
		}
		return nil
	}, nil
}
