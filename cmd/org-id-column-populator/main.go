package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"

	"github.com/RedHatInsights/tenant-utils/pkg/tenantconv"
	"github.com/RedHatInsights/tenant-utils/pkg/tenantid"
	clowder "github.com/redhatinsights/app-common-go/pkg/api/v1"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	logger := initLogger()
	defer flushLogger()
	cmd := NewRootCommand(logger)
	if err := cmd.Execute(); err != nil {
		flushLogger()
		os.Exit(1)
	}
}

func NewRootCommand(logger *logrus.Logger) *cobra.Command {

	var dbConnMetadata dbConnectionMetadata
	var dbTable string
	var dbAccountColumn string
	var dbOrgIdColumn string
	var dbNullOrgIdPlaceholder string
	var batchSize int
	var clowderConfig bool
	var prometheusPushGatewayAddr string
	var translatorServiceAddr string
	var translatorServiceTimeout int
	var dbOperationTimeout int

	var rootCmd = &cobra.Command{
		Use:   "org-id-column-populator",
		Short: "Run the org_id column population util",
		Run: func(cmd *cobra.Command, args []string) {

			var err error

			if clowderConfig {
				logger.Println("Attempting to read database configuration from clowder config file")
				dbConnMetadata, err = readDBConfigFromClowder()
				if err != nil {
					logger.Fatal(err)
				}
			}

			connStr, err := buildDBConnectionString(dbConnMetadata)
			if err != nil {
				logger.Fatal(err)
			}

			db, err := initializeDBConnection(connStr)
			if err != nil {
				logger.Fatal(err)
			}

			metricsRecorder, registry := buildMetricsRecorder(logger, prometheusPushGatewayAddr)

			accountNumberTranslator, err := buildAccountNumberTranslator(logger, translatorServiceAddr, translatorServiceTimeout, registry)
			if err != nil {
				logger.Fatal(err)
			}

			err = tenantconv.MapAccountToOrgId(context.Background(), db, dbTable, dbAccountColumn, dbOrgIdColumn, dbNullOrgIdPlaceholder, dbOperationTimeout, batchSize, accountNumberTranslator, metricsRecorder, logger)
			if err != nil {
				logger.Fatal(err)
			}
		},
	}

	rootCmd.Flags().StringVarP(&dbConnMetadata.host, "db-host", "H", "localhost", "Database hostname")
	rootCmd.Flags().IntVarP(&dbConnMetadata.port, "db-port", "p", 5432, "Database port")
	rootCmd.Flags().StringVarP(&dbConnMetadata.user, "db-user", "u", "insights", "Database user")
	rootCmd.Flags().StringVarP(&dbConnMetadata.password, "db-password", "w", "1234", "Database password")
	rootCmd.Flags().StringVarP(&dbConnMetadata.name, "db-name", "n", "cloud-connector", "Database name")
	rootCmd.Flags().StringVarP(&dbConnMetadata.sslMode, "db-ssl-mode", "s", "verify-full", "Database ssl mode")
	rootCmd.Flags().StringVarP(&dbConnMetadata.sslRootCert, "db-ssl-root-cert", "c", "", "Location of the root certificate file")

	rootCmd.Flags().StringVarP(&dbTable, "db-table-name", "t", "connections", "Database name")
	rootCmd.Flags().StringVarP(&dbAccountColumn, "db-account-column-name", "a", "account", "Account column within db table")
	rootCmd.Flags().StringVarP(&dbOrgIdColumn, "db-org-id-column-name", "o", "org-id", "OrgID column within db table")
	rootCmd.Flags().StringVar(&dbNullOrgIdPlaceholder, "db-null-org-id-place-holder", "", "Place holder value used in the org_id column that should be replaced with real org id")
	rootCmd.Flags().IntVar(&dbOperationTimeout, "db-operation-timeout", 5, "Timeout for each db operation in number of seconds")

	rootCmd.Flags().BoolVarP(&clowderConfig, "read-clowder-config", "C", false, "Read db config from clowder config file")

	rootCmd.Flags().IntVarP(&batchSize, "batch-size", "b", 100, "Number of accounts to retrieve from the database at once ")

	rootCmd.Flags().StringVarP(&prometheusPushGatewayAddr, "prometheus-push-addr", "g", "", "Address of prometheus push gateway")

	rootCmd.Flags().StringVarP(&translatorServiceAddr, "ean-translator-addr", "T", "", "Address of EAN translator service")

	rootCmd.Flags().IntVarP(&translatorServiceTimeout, "ean-translator-timeout", "O", 20, "Timeout for calling the EAN translator service")

	requiredOptions := []string{"db-table-name", "db-account-column-name", "db-org-id-column-name", "ean-translator-addr"}
	for _, requiredOption := range requiredOptions {
		err := rootCmd.MarkFlagRequired(requiredOption)
		if err != nil {
			logger.Fatal(err)
		}
	}

	return rootCmd
}

type dbConnectionMetadata struct {
	host        string
	port        int
	user        string
	password    string
	name        string
	sslMode     string
	sslRootCert string
}

func buildDBConnectionString(metadata dbConnectionMetadata) (string, error) {

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		metadata.host,
		metadata.port,
		metadata.user,
		metadata.password,
		metadata.name,
		metadata.sslMode)

	if metadata.sslRootCert != "" {
		connStr = connStr + " sslrootcert=" + metadata.sslRootCert
	}

	return connStr, nil
}

func initializeDBConnection(connStr string) (*sql.DB, error) {

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error creating connection to the database - %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error testing connection to the database - %w", err)
	}

	return db, nil
}

func readDBConfigFromClowder() (dbConnectionMetadata, error) {

	if !clowder.IsClowderEnabled() {
		return dbConnectionMetadata{}, fmt.Errorf("Clowder is not enabled!  Set \"ACG_CONFIG\" environment variable!")
	}

	cfg := clowder.LoadedConfig

	var pathToDBCertFile string
	var err error
	if cfg.Database.RdsCa != nil {
		pathToDBCertFile, err = cfg.RdsCa()
		if err != nil {
			return dbConnectionMetadata{}, err
		}
	}

	return dbConnectionMetadata{
		host:        cfg.Database.Hostname,
		port:        cfg.Database.Port,
		user:        cfg.Database.Username,
		password:    cfg.Database.Password,
		name:        cfg.Database.Name,
		sslMode:     cfg.Database.SslMode,
		sslRootCert: pathToDBCertFile,
	}, nil
}

func buildAccountNumberTranslator(logger *logrus.Logger, translatorServiceAddr string, timeout int, registry *prometheus.Registry) (tenantid.BatchTranslator, error) {

	if translatorServiceAddr == "test" {
		logger.Println("Using fake EAN translator impl")

		return &tenantconv.TestBatchTranslator{}, nil
	}

	if translatorServiceAddr != "" {
		logger.Printf("Using real EAN translator impl (addr => %s)\n", translatorServiceAddr)

		options := []tenantid.TranslatorOption{
			tenantid.WithTimeout(time.Duration(timeout) * time.Second),
		}

		if registry != nil {
			options = append(options, tenantid.WithMetricsWithCustomRegisterer(registry))
		}

		return tenantid.NewTranslator(translatorServiceAddr, options...), nil
	}

	return nil, fmt.Errorf("EAN translator impl not set!")
}

func buildMetricsRecorder(logger *logrus.Logger, pushGatewayAddr string) (tenantconv.MetricsRecorder, *prometheus.Registry) {

	if pushGatewayAddr != "" {
		logger.Println("Configured to send metrics to prometheus")

		hostName, err := os.Hostname()
		if err != nil {
			logger.Fatal("unable to determine host name for use a prometheus push job name")
		}

		registry := prometheus.NewRegistry()

		metricsRecorder, err := tenantconv.PrometheusMetricsRecorder(logger, pushGatewayAddr, hostName, registry)
		if err != nil {
			logger.Fatal(err)
		}

		return metricsRecorder, registry
	}

	logger.Println("Configured to discard metrics")
	metricsRecorder, _ := tenantconv.TestMetricsRecorder(logger)

	return metricsRecorder, nil
}
