package Db

import (
	"database/sql"
	"reflect"
	"errors"
)

type Result struct {
	Data         []map[string]string
	LastInsertId int
	AffectedRows int
	RowCount     int
}

func (this *Result) One(v interface {}) error {
	if len(this.Data) < 1 {
		return nil
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return errors.New("need pointer struct")
	}
	mapToStruct(this.Data[0], rv)
	return nil
}

func (this *Result) All(v interface {}) error {
	if len(this.Data) < 1 {
		return nil
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return errors.New("need pointer struct")
	}
	rt := rv.Elem().Type().Elem().Elem()
	for _, mapData := range this.Data {
		rvItem := reflect.New(rt)
		mapToStruct(mapData, rvItem)
		rv.Elem().Set(reflect.Append(rv.Elem(), rvItem))
	}
	return nil
}

func newResultFromSqlResult(res sql.Result) *Result {
	lastId, _ := res.LastInsertId()
	affected, _ := res.RowsAffected()
	result := new(Result)
	result.LastInsertId = int(lastId)
	result.AffectedRows = int(affected)
	result.Data = make([]map[string]string, 0)
	return result
}

func newResultFromRows(rows *sql.Rows) *Result {
	result := new(Result)
	result.Data, _ = rowsToMap(rows)
	result.RowCount = len(result.Data)
	return result
}
