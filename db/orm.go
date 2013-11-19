package db

import (
	"reflect"
	"errors"
	"fmt"
	"strings"
)

var (
	ormTypes map[string]*ormType
)

func init() {
	ormTypes = make(map[string]*ormType)
}

type ormType struct {
	tableName   string
	reflectType reflect.Type
	columns map[string]string
	columnTypes map[string]reflect.Kind
}

// create new orm type
func newOrmType(obj interface {}) (*ormType, error) {
	reflectType := reflect.TypeOf(obj)
	// check pointer
	if reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
	}
	if reflectType.Kind() != reflect.Struct {
		return nil, errors.New("must define struct for orm")
	}
	key := fmt.Sprint(reflectType)
	// count fields
	fieldNum := reflectType.NumField()
	if fieldNum < 1 {
		return nil, errors.New("must define struct with fields")
	}
	ormTp := &ormType{}
	ormTp.reflectType = reflectType
	ormTp.columns = make(map[string]string)
	ormTp.columnTypes = make(map[string]reflect.Kind)
	// parse fields
	for i := 0; i < fieldNum; i++ {
		rf := reflectType.Field(i)
		table, col := rf.Tag.Get("tbl"), rf.Tag.Get("col")
		if table != "" {
			ormTp.tableName = table
		}
		ormTp.columns[col] = rf.Name
		ormTp.columnTypes[col] = rf.Type.Kind()
	}
	if ormTp.tableName == "" {
		return nil, errors.New("must define struct with table name")
	}
	ormTypes[key] = ormTp
	return ormTp, nil
}

// get defined orm type
func getOrmType(name string) *ormType {
	return ormTypes[name]
}

// define orm type, return errors slice if some of them fail.
// the orm type named as packageName.StructName such as main.XXX
func Define(objects...interface {}) []error {
	errors := make([]error, 0)
	for _, obj := range objects {
		_, e := newOrmType(obj)
		if e != nil {
			errors = append(errors, e)
		}
	}
	return errors
}

// convert map to defined orm struct
func MapToStruct(mapData map[string]string, typeName string) (interface {}, error) {
	ormType := getOrmType(typeName)
	if ormType == nil {
		return nil, errors.New("wrong orm struct type name (forget define ?)")
	}
	rv := reflect.New(ormType.reflectType)
	for col, field := range ormType.columns {
		strValue := mapData[col]
		colType := ormType.columnTypes[col]
		reflectField := rv.Elem().FieldByName(field)
		switch colType{
		case reflect.Float32, reflect.Float64:
			reflectField.SetFloat(strToFloat64(strValue))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			reflectField.SetInt(strToInt64(strValue))
		case reflect.Bool:
			reflectField.SetBool(strToBool(strValue))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			reflectField.SetUint(strToUint64(strValue))
		default :
			reflectField.SetString(strValue)
		}
	}
	return rv.Interface(), nil
}

// convert defined orm struct to map
func StructToMap(obj interface {}, typeName string) (map[string]string, error) {
	ormType := getOrmType(typeName)
	if ormType == nil {
		return nil, errors.New("wrong orm struct type name (forget define ?)")
	}
	rv := reflect.ValueOf(obj)
	// check pointer
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	mapData := make(map[string]string)
	for col, field := range ormType.columns {
		value := fmt.Sprint(rv.FieldByName(field).Interface())
		if value == "<nil>" {
			value = ""
		}
		mapData[col] = value
	}
	return mapData, nil
}

type InkOrm struct {
	db *InkDatabase
}

// insert orm struct
func (this *InkOrm) Insert(obj interface {}) (int, error) {
	objType := strings.TrimPrefix(fmt.Sprint(reflect.TypeOf(obj)), "*")
	ormTp := getOrmType(objType)
	mapData, e := StructToMap(obj, objType)
	if e != nil {
		return -1, e
	}
	this.db.trigger("orm.insert.before", obj)
	keys := make([]string, len(mapData))
	values := make([]interface {}, len(mapData))
	i := 0
	for k, v := range mapData {
		keys[i] = k
		values[i] = v
		i++
	}
	sql := NewSql(ormTp.tableName, keys...)
	result, e := this.db.Exec(sql.Insert(), values...)
	if e != nil {
		return -1, e
	}
	return result.lastInsertId, nil
}

// delete orm struct by key column value
func (this *InkOrm) Delete(obj interface {}, keyColumn string) (int, error) {
	objType := strings.TrimPrefix(fmt.Sprint(reflect.TypeOf(obj)), "*")
	ormTp := getOrmType(objType)
	mapData, e := StructToMap(obj, objType)
	if e != nil {
		return -1, e
	}
	this.db.trigger("orm.delete.before", obj, &keyColumn)
	value := mapData[keyColumn]
	sql := NewSql(ormTp.tableName).Where("`" + keyColumn + "` = ?")
	result, e := this.db.Exec(sql.Delete(), value)
	if e != nil {
		return -1, e
	}
	return result.affectedRows, nil
}

// update orm struct by key column value ( with specific columns )
func (this *InkOrm) Update(obj interface {}, keyColumn string, columns...string) (int, error) {
	objType := strings.TrimPrefix(fmt.Sprint(reflect.TypeOf(obj)), "*")
	ormTp := getOrmType(objType)
	mapData, e := StructToMap(obj, objType)
	if e != nil {
		return -1, e
	}
	this.db.trigger("orm.update.before", obj, &keyColumn, &columns)
	keyValue := mapData[keyColumn]
	delete(mapData, keyColumn)
	keys := make([]string, 0)
	values := make([]interface {}, 0)
	if len(columns) > 0 {
		for _, col := range columns {
			if col == keyColumn {
				return -1, errors.New("can not update column markd as update-key column")
			}
			v, ok := mapData[col]
			if !ok {
				return -1, errors.New("no column named '" + col + "' in this object")
			}
			values = append(values, v)
		}
		keys = columns;
	}else {
		for k, v := range mapData {
			keys = append(keys, k)
			values = append(values, v)
		}
	}
	values = append(values, keyValue)
	sql := NewSql(ormTp.tableName, keys...).Where("`" + keyColumn + "` = ?").Update()
	result, e := this.db.Exec(sql, values...)
	if e != nil {
		return -1, e
	}
	return result.affectedRows, nil
}

// find as orm struct type with sql conditions and arguments,
// return interface slice, you need assert type by yourself.
func (this *InkOrm) Find(typeName string, sqlCondition *InkSql, args...interface {}) ([]interface {}, error) {
	ormTp := getOrmType(typeName)
	if sqlCondition == nil{
		sqlCondition = NewSql("")
	}
	sqlCondition.Table = ormTp.tableName
	data, e := this.db.Query(sqlCondition.Select(), args...)
	if e != nil {
		return nil, e
	}
	this.db.trigger("orm.find.before", &typeName, sqlCondition, &args)
	resultData := make([]interface {}, len(data.Maps()))
	for i, v := range data.Maps() {
		resultData[i], e = MapToStruct(v, typeName)
		if e != nil {
			return nil, e
		}
	}
	return resultData, nil
}

// find one orm struct
func (this *InkOrm) FindOne(typeName string, sqlCondition *InkSql, args...interface {}) (interface {}, error) {
	ormTp := getOrmType(typeName)
	if sqlCondition == nil{
		sqlCondition = NewSql("")
	}
	sqlCondition.Table = ormTp.tableName
	this.db.trigger("orm.find.one.before", &typeName, sqlCondition, &args)
	data, e := this.db.Query(sqlCondition.Select(), args...)
	if e != nil {
		return nil, e
	}
	return MapToStruct(data.Map(), typeName)
}

// create new struct
func NewOrm(db *InkDatabase) *InkOrm {
	orm := &InkOrm{}
	orm.db = db
	return orm
}
