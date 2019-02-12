package db

import (
	"context"
	"database/sql"

	"online_judge/talcity/scaffold/criteria/trace"
)

type SqlExecutor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

type SqlCtxExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type sqlExecWithContext struct {
	SqlCtxExecutor
	ctx context.Context
}

func (sec *sqlExecWithContext) addTrace(query string, args ...interface{}) {
	val := sec.ctx.Value(trace.DefaultKey)
	if val == nil {
		return
	}
	tr := val.(*trace.T)
	tr.Stage("SQL: %s, args: %v", query, args)
}

func (sec *sqlExecWithContext) Exec(query string, args ...interface{}) (sql.Result, error) {
	sec.addTrace(query, args...)
	return sec.ExecContext(sec.ctx, query, args...)
}

func (sec *sqlExecWithContext) QueryRow(query string, args ...interface{}) *sql.Row {
	sec.addTrace(query, args...)
	return sec.QueryRowContext(sec.ctx, query, args...)
}

func (sec *sqlExecWithContext) Query(query string, args ...interface{}) (*sql.Rows, error) {
	sec.addTrace(query, args...)
	return sec.QueryContext(sec.ctx, query, args...)
}

// SqlExecWithContext build SqlExecutor
func SqlExecWithContext(ctx context.Context, db SqlCtxExecutor) SqlExecutor {
	return &sqlExecWithContext{
		SqlCtxExecutor: db,
		ctx:            ctx,
	}
}
