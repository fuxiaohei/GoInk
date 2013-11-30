package Db

import (
	"reflect"
	"fmt"
)

var definedStruct map[string]*dbStruct

func init() {
	definedStruct = make(map[string]*dbStruct)
}

type dbStruct struct {
	nativeField map[string]reflect.Kind
	aliasField map[string]reflect.Kind
	aliasMap map[string]string
}

func NewStruct(args...interface {}) {
	for _, v := range args {
		rf := reflect.TypeOf(v)
		if rf.Kind() == reflect.Ptr {
			rf = rf.Elem()
		}
		fieldNum := rf.NumField()
		if fieldNum < 1 {
			continue
		}
		dbStruct := new(dbStruct)
		dbStruct.aliasField = make(map[string]reflect.Kind)
		dbStruct.nativeField = make(map[string]reflect.Kind)
		dbStruct.aliasMap = make(map[string]string)
		for i := 0; i < fieldNum; i++ {
			field := rf.Field(i)
			dbStruct.nativeField[field.Name] = field.Type.Kind()
			dbStruct.aliasField[snakeCasedName(field.Name)] = field.Type.Kind()
			dbStruct.aliasMap[snakeCasedName(field.Name)] = field.Name
		}
		definedStruct[fmt.Sprint(rf)] = dbStruct
	}
}

func mapToStruct(mapData map[string]string, rv reflect.Value) {
	key := fmt.Sprint(rv.Elem().Type())
	dbStruct := definedStruct[key]
	if dbStruct == nil {
		NewStruct(rv.Interface())
		dbStruct = definedStruct[key]
	}
	for k, strValue := range mapData {
		kind := dbStruct.nativeField[k]
		fieldName := k
		if kind == reflect.Invalid {
			kind = dbStruct.aliasField[k]
			fieldName = dbStruct.aliasMap[k]
		}
		if kind == reflect.Invalid {
			continue
		}
		reflectField := rv.Elem().FieldByName(fieldName)
		switch kind{
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
}
