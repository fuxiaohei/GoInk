package db

type InkDbResult struct {
	resultMap    []map[string]string
	lastInsertId int
	affectedRows int
	currentSql   string
}

// get the sql whose result is this
func (this *InkDbResult) Sql() string {
	return this.currentSql
}

// get map result, if multi-rows result, return the first one
func (this *InkDbResult) Map() map[string]string {
	if this.resultMap == nil {
		return nil
	}
	if len(this.resultMap) < 1 {
		return nil
	}
	return this.resultMap[0]
}

// get map slice result if multi-rows result
func (this *InkDbResult) Maps() []map[string]string {
	return this.resultMap
}

// convert map to struct
func (this *InkDbResult) One(typeName string) (interface {}, error) {
	mapData := this.Map()
	if mapData == nil {
		return nil, nil
	}
	return MapToStruct(mapData, typeName)
}

// convert whole map slice to struct
func (this *InkDbResult) All(typeName string) ([]interface {}, error) {
	allMapData := this.Maps()
	if allMapData == nil {
		return nil, nil
	}
	result := make([]interface {}, len(allMapData))
	var e error
	for i, mapData := range allMapData {
		result[i], e = MapToStruct(mapData, typeName)
		if e != nil {
			return nil, e
		}
	}
	return result, nil
}

// get last inserted id
func (this *InkDbResult) LastId() int {
	return this.lastInsertId
}

// get affected rows if update or delete
func (this *InkDbResult) Affected() int {
	return this.affectedRows
}

