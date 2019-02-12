package db

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"online_judge/talcity/scaffold/criteria/dsl"
	"online_judge/talcity/scaffold/criteria/merr"
)

// ReadRows
func ReadRows(sqlExec SqlExecutor, table string,
	fields []string, fieldAndValue map[string]interface{}) (*sql.Rows, error) {
	argsLen := len(fieldAndValue)
	if argsLen == 0 {
		return nil, merr.Wrap(nil, 0, "empty query args")
	}
	if len(fields) == 0 {
		return nil, merr.Wrap(nil, 0, "no fields to search")
	}
	for i := range fields {
		fields[i] = "`" + fields[i] + "`"
	}

	var (
		conditions = make([]string, 0, argsLen)
		args       = make([]interface{}, 0, argsLen)
	)

	for f, v := range fieldAndValue {
		conditions = append(conditions, fmt.Sprintf("`%s` = ?", f))
		args = append(args, v)
	}

	SQL := fmt.Sprintf(
		"select %s from `%s` where %s",
		strings.Join(fields, ", "),
		table,
		strings.Join(conditions, " and "),
	)

	rows, err := sqlExec.Query(SQL, args...)
	if err != nil {
		return nil, merr.Wrap(err, 0)
	}
	return rows, nil
}

// Search
// search?filter=team_id,xxx;duty_id,dddd&query=name,kkk;phone,123&order_by=name,desc;created_at,asc&page=1&size=20
func Search(sqlExec SqlExecutor, table string,
	fields []string, filters, querys map[string]interface{},
	orderBys [][]string, from, size int) (*sql.Rows, error) {
	if from < 0 {
		return nil, merr.Wrap(nil, 0, "from not valid")
	}
	if size < 1 {
		return nil, merr.Wrap(nil, 0, "size not valid")
	}
	if len(fields) == 0 {
		return nil, merr.Wrap(nil, 0, "no fields to search")
	}
	for i := range fields {
		fields[i] = "`" + fields[i] + "`"
	}

	filterFields := make([]string, 0, len(filters))
	filterArgs := make([]interface{}, 0, len(filters))
	queryFields := make([]string, 0, len(querys))
	queryArgs := make([]interface{}, 0, len(querys))
	for f, v := range filters {
		filterFields = append(filterFields, fmt.Sprintf("`%s` = ?", f))
		filterArgs = append(filterArgs, v)
	}
	for f, v := range querys {
		queryFields = append(queryFields, fmt.Sprintf("`%s` like ?", f))
		queryArgs = append(queryArgs, fmt.Sprintf("%%%v%%", v))
	}
	sorts := make([]string, 0, len(orderBys))
	for _, sort := range orderBys {
		if len(sort) != 2 {
			return nil, merr.Wrap(nil, 0, "invalid order by parameters")
		}
		sorts = append(sorts, fmt.Sprintf("`%s` %s", sort[0], sort[1]))
	}

	cols := append(filterFields, queryFields...)
	if len(cols) == 0 {
		cols = []string{"1=1"}
	}
	args := append(filterArgs, queryArgs...)
	SQL := fmt.Sprintf(
		"select %s from `%s` where %s",
		strings.Join(fields, ", "),
		table,
		strings.Join(cols, " and "),
	)
	if len(sorts) > 0 {
		SQL = fmt.Sprintf("%s order by %s", SQL, strings.Join(sorts, " , "))
	}
	if size > 0 {
		SQL = SQL + " limit ?, ?"
		args = append(args, from, size)
	}

	rows, err := sqlExec.Query(SQL, args...)
	if err != nil {
		return nil, merr.Wrap(err, 0)
	}
	return rows, nil
}

// Count: Deprecated
func Count(sqlExec SqlExecutor, table string,
	filters, querys map[string]interface{}) (int64, error) {
	filterFields := make([]string, 0, len(filters))
	filterArgs := make([]interface{}, 0, len(filters))
	queryFields := make([]string, 0, len(querys))
	queryArgs := make([]interface{}, 0, len(querys))
	for f, v := range filters {
		filterFields = append(filterFields, fmt.Sprintf("`%s` = ?", f))
		filterArgs = append(filterArgs, v)
	}
	for f, v := range querys {
		queryFields = append(queryFields, fmt.Sprintf("`%s` like ?", f))
		queryArgs = append(queryArgs, fmt.Sprintf("%%%v%%", v))
	}
	cols := append(filterFields, queryFields...)
	if len(cols) == 0 {
		cols = []string{"1=1"}
	}
	args := append(filterArgs, queryArgs...)
	SQL := fmt.Sprintf("select count(1) from `%s` where %s", table, strings.Join(cols, " and "))

	row := sqlExec.QueryRow(SQL, args...)
	var total int64
	if err := row.Scan(&total); err != nil {
		return 0, merr.Wrap(err, 0)
	}
	return total, nil
}

func CountByDSL(sqlExec SqlExecutor, table string, matcher *dsl.Clause) (int64, error) {
	SQL, args, err := dsl.BuildCountSQL(table, matcher)
	if err != nil {
		return 0, err
	}

	row := sqlExec.QueryRow(SQL, args...)
	var total int64
	if err := row.Scan(&total); err != nil {
		return 0, merr.Wrap(err, 0)
	}
	return total, nil
}

// UpdateByID
func UpdateByID(sqlExec SqlExecutor, ID interface{}, tableName string, fieldAndValues map[string]interface{}) (sql.Result, error) {
	return Update(sqlExec, tableName, map[string]interface{}{"id": ID}, fieldAndValues)
}

func Update(sqlExec SqlExecutor, tableName string, filters, fieldAndValues map[string]interface{}) (sql.Result, error) {
	if len(filters) == 0 {
		return nil, merr.Wrap(nil, 0, "empty filters")
	}
	if len(fieldAndValues) == 0 {
		return nil, merr.Wrap(nil, 0, "nothing to update")
	}

	fields := make([]string, 0, len(fieldAndValues))
	args := make([]interface{}, 0, len(fields))
	for f, v := range fieldAndValues {
		refv := reflect.ValueOf(v)
		if refv.Kind() == reflect.Ptr && refv.IsNil() {
			continue
		}
		fields = append(fields, fmt.Sprintf("`%s` = ?", f))
		args = append(args, v)
	}
	if len(fields) == 0 {
		return nil, merr.Wrap(nil, 0, "nothing to update")
	}

	filterCols := make([]string, 0, len(filters))
	for f, v := range filters {
		filterCols = append(filterCols, fmt.Sprintf("`%s` = ?", f))
		args = append(args, v)
	}

	SQL := fmt.Sprintf("update `%s` set %s where %s",
		tableName,
		strings.Join(fields, ", "),
		strings.Join(filterCols, " and "),
	)
	res, err := sqlExec.Exec(SQL, args...)
	if err != nil {
		return nil, merr.Wrap(err, 0)
	}

	return res, nil
}

func InTransaction(ctx context.Context, errCode int, conn *MySQLConn, operations func(sqlExec SqlExecutor) error) (lastErr error) {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return merr.Wrap(err, errCode, "begin tx failed")
	}
	var operationErr error
	defer func() {
		if operationErr != nil {
			if tmpErr := tx.Rollback(); tmpErr != nil {
				lastErr = merr.Wrap(tmpErr, errCode, "tx rollback failed")
				return
			}
			lastErr = operationErr
			return
		}
		if tmpErr := tx.Commit(); tmpErr != nil {
			lastErr = merr.Wrap(tmpErr, errCode, "tx commit failed")
		}
	}()
	operationErr = operations(tx)
	return
}
