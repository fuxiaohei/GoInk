/**
 * Created by FuXiaoHei on 13-11-30.
 */
package Db

import (
	"fmt"
	"strings"
)

type Sql struct{
	Table       string
	columns     []string
	distinct map[string]bool
	whereString []string
	orderString string
	limitString string
	groupString string
}

func (this *Sql) Column(str string) *Sql {
	this.columns = strings.Split(str, ",")
	return this
}

func (this *Sql) Distinct(str string) *Sql {
	columns := strings.Split(str, ",")
	for _, col := range columns {
		this.distinct[col] = true
	}
	return this
}

func (this *Sql) Where(column, condition string) *Sql {
	str := "`"+column+"` " + condition
	this.whereString = append(this.whereString, str)
	return this
}

func (this *Sql) WhereIn(column string, valueNum int) *Sql {
	str := "`"+column+"` IN ("+strings.TrimSuffix(strings.Repeat("?,", valueNum), ",") + ")"
	this.whereString = append(this.whereString, str)
	return this
}

func (this *Sql) WhereBetween(column string) *Sql {
	str := "`"+column + "` BETWEEN ? AND ?"
	this.whereString = append(this.whereString, str)
	return this
}

func (this *Sql) Order(cond string) *Sql {
	this.orderString = " ORDER BY " + cond
	return this
}

func (this *Sql) Group(column string) *Sql {
	this.groupString = " GROUP BY `"+column + "`"
	return this
}

func (this *Sql) Limit(limit int) *Sql {
	if limit == -1 {
		this.limitString = " LIMIT ?"
		return this
	}
	this.limitString = " LIMIT " + fmt.Sprint(limit)
	return this
}

func (this *Sql) Page(page, size int) *Sql {
	if page == -1 && size == -1 {
		this.limitString = " LIMIT ?,?"
		return this
	}
	this.limitString = " LIMIT "+fmt.Sprint((page-1)*size)+"," + fmt.Sprint(size)
	return this
}

func (this *Sql) Select() string {
	sql := "SELECT "
	if len(this.columns) < 1 {
		sql += "* FROM "
	}else {
		for _, col := range this.columns {
			if this.distinct[col] {
				sql += "DISTINCT `"+col + "`,"
			}else {
				sql += "`"+col + "`,"
			}
		}
		sql = strings.TrimSuffix(sql, ",") + " FROM "
	}
	sql += this.Table
	if len(this.whereString) > 0 {
		sql +=" WHERE " + strings.Join(this.whereString, " AND ")
	}
	sql +=this.groupString
	sql +=this.orderString
	sql +=this.limitString
	return sql
}

func (this *Sql) Count() string {
	this.columns = []string{}
	this.limitString = ""
	sql := this.Select()
	return strings.Replace(sql, "*", "count(*) AS countNum", -1)
}

func (this *Sql) Update() string {
	sql := "UPDATE "+this.Table + " SET `"
	if len(this.columns) < 1{
		return ""
	}
	sql += strings.Join(this.columns, "` = ?,`") + "` = ?"
	if len(this.whereString) > 0 {
		sql +=" WHERE " + strings.Join(this.whereString, " AND ")
	}
	return sql
}

func (this *Sql) Delete()string{
	sql := "DELETE FROM " + this.Table
	if len(this.whereString) > 0 {
		sql +=" WHERE " + strings.Join(this.whereString, " AND ")
	}
	sql +=this.orderString
	sql +=this.limitString
	return sql
}

func (this *Sql) Insert() string {
	sql := "INSERT INTO " + this.Table + "(`"
	sql += strings.Join(this.columns, "`,`") + "`) VALUES ("
	sql += strings.TrimSuffix(strings.Repeat("?,", len(this.columns)), ",") + ")"
	return sql
}

func (this *Engine) NewSql(table string) *Sql {
	sql := new(Sql)
	sql.Table = table
	sql.columns = []string{}
	sql.distinct = make(map[string]bool)
	return sql
}




