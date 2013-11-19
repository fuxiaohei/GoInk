package db

import (
	"database/sql"
	"reflect"
	"fmt"
)

type InkDatabaseOption struct {
	Driver         string
	Dsn            string
	MaxConnection  int
	IdleConnection int
	LogsNumber     int
	Mode           string
}

// get debug mode
func (this *InkDatabaseOption) isDebug() bool {
	return this.Mode == "debug"
}

type InkDatabase struct {
	db *sql.DB
	sqlLogs []string
	option *InkDatabaseOption
	filter  databaseFilter
}

// trigger event if filter is initialized
func (this *InkDatabase) trigger(event string, args... interface {}) ([]interface {}, error) {
	if this.filter != nil {
		return this.filter.Trigger(event, args...)
	}
	return nil, nil
}

// query sql string
func (this *InkDatabase) Query(sqlStr string, args...interface {}) (*InkDbResult, error) {
	if this.option.isDebug() {
		fmt.Println(sqlStr)
		this.trigger("database.query.before", &sqlStr, &args)
	}
	result := &InkDbResult{}
	result.lastInsertId = -1
	result.affectedRows = -1
	result.currentSql = sqlStr
	rows, e := this.db.Query(sqlStr, args...)
	if e != nil {
		return nil, e
	}
	result.resultMap, e = RowsToMap(rows)
	if e != nil {
		return nil, e
	}
	return result, nil
}

// execute sql, no result records queries
func (this *InkDatabase) Exec(sqlStr string, args...interface {}) (*InkDbResult, error) {
	if this.option.isDebug() {
		fmt.Println(sqlStr)
		this.trigger("database.exec.before", &sqlStr, &args)
	}
	result := &InkDbResult{}
	result.currentSql = sqlStr
	sqlResult, e := this.db.Exec(sqlStr, args...)
	if e != nil {
		return nil, e
	}
	parsedResult, e2 := ResultToInt(sqlResult)
	if e2 != nil {
		return nil, e2
	}
	result.lastInsertId = int(parsedResult[0])
	result.affectedRows = int(parsedResult[1])
	return result, nil
}

// map row data to string
func RowsToMap(rows *sql.Rows) ([]map[string]string, error) {
	cols, e := rows.Columns()
	if e != nil {
		return nil, e
	}
	tmpItf := make([]interface{}, len(cols))
	for i, _ := range tmpItf {
		var itr interface{}
		tmpItf[i] = &itr
	}
	// set returning result
	rs := make([]map[string]string, 0)
	for rows.Next() {
		rows.Scan(tmpItf...)
		rowRst := make(map[string]string)
		// make all column values to string and append to result
		for i, col := range tmpItf {
			str := fmt.Sprint(reflect.Indirect(reflect.ValueOf(col)).Interface())
			if str == "<nil>" {
				str = ""
			}
			rowRst[cols[i]] = str
		}
		rs = append(rs, rowRst)
	}
	return rs, nil
}

// parse sql.Result to int64
func ResultToInt(rs sql.Result) ([2]int64, error) {
	res := [2]int64{-1, -1}
	i, e := rs.LastInsertId()
	if e != nil {
		return res, e
	}
	res[0] = i
	if i < 1 {
		i, e = rs.RowsAffected()
		if e != nil {
			return res, e
		}
		res[1] = 1
	}
	return res, nil
}

// create new database
func NewDatabase(option *InkDatabaseOption, filter databaseFilter) (*InkDatabase, error) {
	sqlDb, err := sql.Open(option.Driver, option.Dsn)
	if err != nil {
		return nil, err
	}
	sqlDb.SetMaxOpenConns(option.MaxConnection)
	sqlDb.SetMaxIdleConns(option.IdleConnection)
	db := &InkDatabase{}
	db.db = sqlDb
	db.option = option
	db.sqlLogs = make([]string, 0)
	db.filter = filter
	if db.option.isDebug() {
		db.trigger("database.new", db)
	}
	return db, nil
}



