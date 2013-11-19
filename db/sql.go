package db

import (
	"strings"
	"fmt"
)

var cacheSql map[string]string

func init() {
	cacheSql = make(map[string]string)
}

type InkSql struct {
	Table       string
	Columns     []string
	WhereStr    string
	OrderStr    string
	LimitNum    [2]int
	PageNum     [2]int
	GroupStr    string
	HavingStr   string
	buildString string
}

// add where condition.
// several where conditions are joined by AND.
func (this *InkSql) Where(where string) *InkSql {
	if len(this.WhereStr) > 0 {
		this.WhereStr += " AND " + where
	} else {
		this.WhereStr = where
	}
	return this
}

// add order condition
func (this *InkSql) Order(order string) *InkSql {
	this.OrderStr = order
	return this
}

// add limit condition
func (this *InkSql) Limit(limit int, offset int) *InkSql {
	this.LimitNum[0] = limit
	this.LimitNum[1] = offset
	return this
}

// add pagination condition
func (this *InkSql) Page(page int, size int) *InkSql {
	this.PageNum[0] = (page - 1)*size
	this.PageNum[1] = size
	return this
}

// add group by and having condition
func (this *InkSql) Group(group string, having string) *InkSql {
	this.GroupStr = group
	this.HavingStr = having
	return this
}

// create select query string
func (this *InkSql) Select() string {
	sql := "SELECT"
	if len(this.Columns) < 1 {
		sql += " * "
	} else {
		sql += " `" + strings.Join(this.Columns, "`,`") + "` "
	}
	sql += "FROM " + this.Table
	if len(this.WhereStr) > 0 {
		sql += " WHERE " + this.WhereStr
	}
	if len(this.GroupStr) > 0 {
		sql += " GROUP BY `" + this.GroupStr + "`"
		if len(this.HavingStr) > 0 {
			sql += " HAVING " + this.HavingStr
		}
	}
	if len(this.OrderStr) > 0 {
		sql += " ORDER BY " + this.OrderStr
	}
	if this.PageNum[1] > 0 {
		sql += " LIMIT " + fmt.Sprint(this.PageNum[0], " , ", this.PageNum[1])
		return sql
	}
	if this.LimitNum[0] != -1 {
		sql += " LIMIT " + fmt.Sprint(this.LimitNum[0])
		if this.LimitNum[1] > 0 {
			sql += " OFFSET " + fmt.Sprint(this.LimitNum[1])
		}
	}
	return sql
}

// create insert query string
func (this *InkSql) Insert() string {
	sql := "INSERT INTO " + this.Table + "(`"
	sql += strings.Join(this.Columns, "`,`") + "`) VALUES ("
	sql += strings.TrimSuffix(strings.Repeat("?,", len(this.Columns)), ",") + ")"
	return sql
}

// create update query string
func (this *InkSql) Update() string {
	sql := "UPDATE " + this.Table + " SET `"
	sql += strings.Join(this.Columns, "` = ?,`") + "` = ?"
	if len(this.WhereStr) > 0 {
		sql += " WHERE " + this.WhereStr
	}
	return sql
}

// create delete query string
func (this *InkSql) Delete() string {
	sql := "DELETE FROM " + this.Table
	if len(this.WhereStr) > 0 {
		sql += " WHERE " + this.WhereStr
	}
	if len(this.OrderStr) > 0 {
		sql += " ORDER BY " + this.OrderStr
	}
	if this.LimitNum[0] != -1 {
		sql += " LIMIT " + fmt.Sprint(this.LimitNum[0])
	}
	return sql
}

// create new sql builder with table name or some columns
func NewSql(table string, columns ...string) *InkSql {
	sql := &InkSql{}
	sql.Table = table
	sql.Columns = columns
	sql.LimitNum = [2]int{-1, -1}
	sql.PageNum = [2]int{-1, -1}
	return sql
}

// cache sql string or get cached sql string
func CacheSql(name string, sql... string) string {
	if len(sql) < 1 {
		return cacheSql[name]
	}
	cacheSql[name] = sql[0]
	return ""
}
