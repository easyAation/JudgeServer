package db

import (

	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"online_judge/talcity/scaffold/criteria/merr"
)

var (
	defaultConnName = "default"
	dataBaseCache   = &dbCache{cache: make(map[string]*MySQLConn)}
)

// database cache.
type dbCache struct {
	mux   sync.RWMutex
	cache map[string]*MySQLConn
}

// add database conn with original name.
func (cache *dbCache) add(name string, conn *MySQLConn) (added bool) {
	cache.mux.Lock()
	if _, ok := cache.cache[name]; !ok {
		cache.cache[name] = conn
		added = true
	}
	cache.mux.Unlock()
	return
}

// get database conn if cached.
func (cache *dbCache) get(name string) (conn *MySQLConn, ok bool) {
	cache.mux.RLock()
	conn, ok = cache.cache[name]
	cache.mux.RUnlock()
	return
}

// get default conn.
func (cache *dbCache) getDefault() (conn *MySQLConn) {
	conn, _ = cache.get(defaultConnName)
	return
}

// detect time zone
func detectTzAndEngine(conn *MySQLConn) {
	// default use Local
	conn.TimeZone = time.Local

	row := conn.QueryRow("SELECT TIMEDIFF(NOW(), UTC_TIMESTAMP)")
	var tz string
	row.Scan(&tz)
	if len(tz) >= 8 {
		if tz[0] != '-' {
			tz = "+" + tz
		}

		t, err := time.Parse("-07:00:00", tz)
		if err != nil {
			fmt.Printf("Detect MySQL DB timezone: %s %s \n", tz, err.Error())
		} else {
			conn.TimeZone = t.Location()
		}
	}

	// get default engine from current database
	conn.Engine = "INNODB"
	row = conn.QueryRow("SELECT ENGINE, TRANSACTIONS FROM INFORMATION_SCHEMA.ENGINES WHERE SUPPORT = 'DEFAULT'")
	var (
		engine string
		tx     bool
	)
	row.Scan(&engine, &tx)
	if engine != "" {
		conn.Engine = engine
	}
}

// addConnWithDB add a conn name for db.
func addConnWithDB(name string, db *sql.DB) (*MySQLConn, error) {
	conn := new(MySQLConn)
	conn.Name = name
	conn.DB = db

	err := db.Ping()
	if err != nil {
		return nil, fmt.Errorf("register db Ping %q, %s", name, err.Error())
	}

	if !dataBaseCache.add(name, conn) {
		return nil, fmt.Errorf("database name %q already registered, cannot many register", name)
	}

	return conn, nil
}

// getDbConn get database connection
func getDbConn(name string) *MySQLConn {
	if conn, ok := dataBaseCache.get(name); ok {
		return conn
	}

	panic(fmt.Sprintf("unknown DataBase conn name %s", name))
}

// SetMaxIdleConns change the max idle conns for *sql.DB about specify database.
func SetMaxIdleConns(connName string, maxIdleCons int) {
	conn := getDbConn(connName)
	conn.maxIdleConns = maxIdleCons
	conn.DB.SetMaxIdleConns(maxIdleCons)
}

// SetMaxOpenConns
func SetMaxOpenConns(connName string, maxOpenConns int) {
	conn := getDbConn(connName)
	conn.maxOpenConns = maxOpenConns
	if fun := reflect.ValueOf(conn.DB).MethodByName("SetMaxOpenConns"); fun.IsValid() {
		fun.Call([]reflect.Value{reflect.ValueOf(maxOpenConns)})
	}
}

// GetConn get *MySQLConn from registered database by db conn name.
// use "default" as conn name if not set.
func GetMySQLConn(connNames ...string) (*MySQLConn, error) {
	name := defaultConnName
	if len(connNames) > 0 {
		name = connNames[0]
	}

	conn, ok := dataBaseCache.get(name)
	if ok {
		return conn, nil
	}

	return nil, fmt.Errorf("DataBase of conn name %q not found", name)
}

// GetSqlExecutor get SqlExecutor from registered database by db conn name.
func GetSqlExecutor(ctx context.Context, name string) (SqlExecutor, error) {
	conn, ok := dataBaseCache.get(name)
	if ok {
		return SqlExecWithContext(ctx, conn), nil
	}
	return nil, merr.Wrap(nil, 0, "DataBase of conn name %q not found", name)
}

// MySQLConfig is mysql config info.
type MySQLConfig struct {
	ConnStr      string
	MaxOpenConns int
	MaxIdleConns int
}

// MySQLConn is mysql conn info.
type MySQLConn struct {
	Name       string
	DataSource string
	*sql.DB
	maxOpenConns int
	maxIdleConns int
	TimeZone     *time.Location
	Engine       string
}

// RegisterDB, init mysql database instance. use the config of dataSource args.
func RegisterDB(connName string, config *MySQLConfig) error {
	var (
		err  error
		db   *sql.DB
		conn *MySQLConn
	)

	if connName == "" {
		connName = defaultConnName
	}

	if config == nil || config.ConnStr == "" {
		err = errors.New("init mysql database instance failed, mysql dataSourceName is empty")
		goto end
	}

	db, err = sql.Open("mysql", config.ConnStr)
	if err != nil {
		err = fmt.Errorf("init mysql database instance failed, DataSourceName is `%s`, Error: %s ", config.ConnStr, err.Error())
		goto end
	}

	conn, err = addConnWithDB(connName, db)
	if err != nil {
		goto end
	}

	conn.DataSource = config.ConnStr

	detectTzAndEngine(conn)

	if config.MaxOpenConns > 0 {
		SetMaxOpenConns(conn.Name, config.MaxOpenConns)
	}

	if config.MaxIdleConns > 0 {
		SetMaxIdleConns(conn.Name, config.MaxIdleConns)
	}

end:
	if err != nil {
		if db != nil {
			db.Close()
		}

		println(err.Error())
	}

	return err
}

// Insert func is insert data to database table by sql.
func (conn *MySQLConn) Insert(sql string, args ...interface{}) (rowsAffected int64, err error) {
	result, err := conn.Exec(sql, args...)
	if err != nil {
		return
	}

	return result.RowsAffected()
}

// Update func is update data to database table by sql
func (conn *MySQLConn) Update(sql string, args ...interface{}) (rowsAffected int64, err error) {
	result, err := conn.Exec(sql, args...)
	if err != nil {
		return
	}

	return result.RowsAffected()
}

// Delete func is delete data from database table by sql.
func (conn *MySQLConn) Delete(sql string, args ...interface{}) (rowsAffected int64, err error) {
	result, err := conn.Exec(sql, args...)
	if err != nil {
		return
	}

	return result.RowsAffected()
}

// Select
func (conn *MySQLConn) Select(sql string, args ...interface{}) (results []map[string]string, err error) {
	rows, err := conn.Query(sql, args...)
	if err != nil {
		return
	}

	defer rows.Close()

	columns, _ := rows.Columns()
	values := make([][]byte, len(columns))
	scans := make([]interface{}, len(columns))
	for i := range values {
		scans[i] = values[i]
	}

	results = make([]map[string]string, 0)
	for rows.Next() {
		if err = rows.Scan(scans...); err != nil {
			println("mysql exec select failed: %s", err.Error())
			return
		}

		row := make(map[string]string)
		for k, v := range values {
			row[columns[k]] = string(v)
		}

		results = append(results, row)
	}

	return
}

type SqlPre struct {
	SqlStr string
	Args   []interface{}
}

// TxExec func is exec sql and args collection by a transaction.
// if a sql exec fail, rollback.
func (conn *MySQLConn) TxExec(sqlPres []*SqlPre) (rowsAffected []int64, err error) {
	tx, err := conn.Begin()
	if err != nil {
		return
	}

	for _, sqlPre := range sqlPres {
		result, err := tx.Exec(sqlPre.SqlStr, sqlPre.Args...)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		num, err := result.RowsAffected()
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		rowsAffected = append(rowsAffected, num)
	}

	err = tx.Commit()
	return
}

// TxExecList is exec sql of args collection by a transaction.
// creates a prepared statement for use within a transaction.
// if statement create or sql exec fail, rollback.
func (conn *MySQLConn) TxExecList(sql string, argsList [][]interface{}) (rowsAffected []int64, err error) {
	tx, err := conn.Begin()
	if err != nil {
		return
	}

	stmt, err := tx.Prepare(sql)
	if err != nil {
		tx.Rollback()
		return
	}

	for _, args := range argsList {
		result, err := stmt.Exec(args...)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		num, err := result.RowsAffected()
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		rowsAffected = append(rowsAffected, num)
	}

	err = tx.Commit()
	return
}
