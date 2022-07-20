package tenantconv

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"github.com/RedHatInsights/tenant-utils/pkg/tenantid"
)

type blessedIdentifier string
type blessedLiteral string

func MapAccountToOrgId(ctx context.Context, database *sql.DB, table, accountColumn, orgIdColumn string, nullOrgIdPlaceholder string, dbOperationTimeout int, batchSize int, batchTranslator tenantid.BatchTranslator, recordMetrics MetricsRecorder, logger *logrus.Logger) error {

	blessedTableName := blessedIdentifier(pq.QuoteIdentifier(table))
	blessedAccountColumnName := blessedIdentifier(pq.QuoteIdentifier(accountColumn))
	blessedOrgIdColumnName := blessedIdentifier(pq.QuoteIdentifier(orgIdColumn))
	blessedNullOrgIdPlaceholder := blessedLiteral(pq.QuoteLiteral(nullOrgIdPlaceholder))

	uniqueAccounts, rowsUpdated, mappingErr := mapAccountToOrgId(ctx, database, blessedTableName, blessedAccountColumnName, blessedOrgIdColumnName, blessedNullOrgIdPlaceholder, dbOperationTimeout, batchSize, batchTranslator, logger)

	// Record the metrics even if there was an error??
	_ = recordMetrics(table, uniqueAccounts, rowsUpdated)

	return mappingErr
}

func mapAccountToOrgId(ctx context.Context, database *sql.DB, table, accountColumn, orgIdColumn blessedIdentifier, nullOrgIdPlaceholder blessedLiteral, dbOperationTimeout int, batchSize int, batchTranslator tenantid.BatchTranslator, logger *logrus.Logger) (int64, int64, error) {

	logger.Println("table:", table)
	logger.Println("accountColumn:", accountColumn)
	logger.Println("orgIdColumn:", orgIdColumn)
	logger.Println("batchSize:", batchSize)
	logger.Println("nullOrgIdPlaceholder: ", nullOrgIdPlaceholder)

	var totalAccounts int64
	var processedAccounts int64
	var rowsUpdated int64
	var excludeSkippedAccounts []string

	totalAccounts, err := calculateUniqueAccounts(ctx, database, table, accountColumn, orgIdColumn, nullOrgIdPlaceholder)
	if err != nil {
		return processedAccounts, rowsUpdated, err
	}

	logger.Println("totalAccounts: ", totalAccounts)

	for processedAccounts < totalAccounts {

		statement, err := buildAccountsQuery(database, table, accountColumn, orgIdColumn, nullOrgIdPlaceholder, batchSize, excludeSkippedAccounts)
		if err != nil {
			return processedAccounts, rowsUpdated, err
		}

		eans, err := readAccountsFromDatabase(ctx, database, statement, batchSize, dbOperationTimeout)
		if err != nil {
			return processedAccounts, rowsUpdated, err
		}

		statement.Close()

		logger.Printf("Read %d unique accounts from table %s\n", len(eans), table)

		if len(eans) == 0 {
			return processedAccounts, rowsUpdated, nil
		}

		translationResults, err := batchTranslator.EANsToOrgIDs(ctx, eans)
		if err != nil {
			return processedAccounts, rowsUpdated, err
		}

		logger.Printf("Processing batch of translated EANs\n")

		for _, results := range translationResults {

			processedAccounts++

			if skipEAN(results) {
				logger.Printf("Skipping translation of EAN to OrgID: %+v\n", results)
				if results.EAN != nil {
					excludeSkippedAccounts = append(excludeSkippedAccounts, *results.EAN)
				}
				continue
			}

			updatedRowsCount, err := updateOrgIdInDatabase(ctx, database, table, accountColumn, orgIdColumn, nullOrgIdPlaceholder, *results.EAN, results.OrgID, dbOperationTimeout)
			if err != nil {
				return processedAccounts, rowsUpdated, err
			}

			if updatedRowsCount == 0 {
				// I can't really think of why this would happen...but I really don't want
				// this thing to just run forever...
				logger.Println("FAILED TO UPDATED ANY ROWS...Looks like we are done!!")
				return processedAccounts, rowsUpdated, fmt.Errorf("Failed to update any rows.")
			}

			rowsUpdated = rowsUpdated + updatedRowsCount
		}

		logger.Printf("Finished processing batch of translated EANs\n")
	}

	return processedAccounts, rowsUpdated, nil
}

func calculateUniqueAccounts(ctx context.Context, database *sql.DB, table, accountColumn, orgIdColumn blessedIdentifier, nullOrgIdPlaceholder blessedLiteral) (int64, error) {

	uniqueAccountQuery := buildUniqueAccountQuery(table, accountColumn, orgIdColumn, nullOrgIdPlaceholder, -1, nil)

	selectQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) subQuery", uniqueAccountQuery)

	statement, err := database.Prepare(selectQuery)
	if err != nil {
		return 0, fmt.Errorf("error number of unique accounts from the database - prepare failed %w", err)
	}

	rows, err := statement.QueryContext(ctx)
	if err != nil {
		return 0, fmt.Errorf("error number of unique accounts from the database - query failed %w", err)
	}
	defer rows.Close()

	rows.Next()

	var numberOfAccounts int64
	if err := rows.Scan(&numberOfAccounts); err != nil {
		return 0, fmt.Errorf("error number of unique accounts from the database - scan failed %w", err)
	}

	return numberOfAccounts, nil
}

func buildAccountsQuery(database *sql.DB, table, accountColumn, orgIdColumn blessedIdentifier, nullOrgIdPlaceholder blessedLiteral, batchSize int, excludeList []string) (*sql.Stmt, error) {

	uniqueAccountQuery := buildUniqueAccountQuery(table, accountColumn, orgIdColumn, nullOrgIdPlaceholder, batchSize, excludeList)

	statement, err := database.Prepare(uniqueAccountQuery)
	if err != nil {
		return nil, fmt.Errorf("error reading accounts from the database - prepare failed %w", err)
	}

	return statement, nil
}

func buildUniqueAccountQuery(table, accountColumn, orgIdColumn blessedIdentifier, nullOrgIdPlaceholder blessedLiteral, batchSize int, excludeList []string) string {
	var limitClause string
	if batchSize > 0 {
		limitClause = fmt.Sprintf(" LIMIT %d", batchSize)
	}

	var excludeAccountsClause string
	if len(excludeList) > 0 {
		var buf bytes.Buffer
		for _, e := range excludeList {
			fmt.Fprintf(&buf, "'%s', ", e)
		}
		buf.Truncate(buf.Len() - 2) // Remove trailing ", "

		excludeAccountsClause = fmt.Sprintf(" AND %s NOT IN (%s)", accountColumn, buf.String())
	}

	return fmt.Sprintf("SELECT %s FROM %s WHERE %s IS NOT NULL AND (%s IS NULL OR %s = %s) %s GROUP BY %s ORDER BY %s %s",
		accountColumn,
		table,
		accountColumn,
		orgIdColumn,
		orgIdColumn,
		nullOrgIdPlaceholder,
		excludeAccountsClause,
		accountColumn,
		accountColumn,
		limitClause)
}

func readAccountsFromDatabase(ctx context.Context, database *sql.DB, statement *sql.Stmt, batchSize int, timeout int) ([]string, error) {

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	accountList := make([]string, 0, batchSize)

	rows, err := statement.QueryContext(ctx)
	if err != nil {
		return accountList, fmt.Errorf("error reading accounts from the database - query failed %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var account string

		if err := rows.Scan(&account); err != nil {
			return accountList, fmt.Errorf("error reading accounts from the database - scan failed %w", err)
		}

		accountList = append(accountList, account)
	}

	return accountList, nil
}

func updateOrgIdInDatabase(ctx context.Context, database *sql.DB, table, accountColumn, orgIdColumn blessedIdentifier, nullOrgIdPlaceholder blessedLiteral, account, orgId string, timeout int) (int64, error) {

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	updateQuery := fmt.Sprintf("UPDATE %s SET %s = %s WHERE %s = %s AND (%s IS NULL OR %s = %s)",
		table,
		orgIdColumn,
		pq.QuoteLiteral(orgId),
		accountColumn,
		pq.QuoteLiteral(account),
		orgIdColumn,
		orgIdColumn,
		nullOrgIdPlaceholder)

	statement, err := database.Prepare(updateQuery)
	if err != nil {
		return 0, fmt.Errorf("error updating org id column in the database - prepare failed %w", err)
	}
	defer statement.Close()

	results, err := statement.ExecContext(ctx)
	if err != nil {
		return 0, fmt.Errorf("error updating org id column in the database - exec failed %w", err)
	}

	rowsUpdated, err := results.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("error updating org id column in the database - rows affected failed %w", err)
	}

	return rowsUpdated, nil
}

func skipEAN(translationResult tenantid.TranslationResult) bool {

	if translationResult.EAN == nil {
		return true
	}

	if translationResult.Err != nil {
		return true
	}

	return false
}
