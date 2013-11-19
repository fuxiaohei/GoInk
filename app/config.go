package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

const (
	CONFIG_JSON = 1
)

var (
	configParseFuncMap map[int]func (bytes []byte) (*InkConfig, error)
)

func init() {
	configParseFuncMap = make(map[int]func (bytes []byte) (*InkConfig, error))
	configParseFuncMap[CONFIG_JSON] = parseJsonConfigBytes
}

type InkConfig map[string]map[string]interface{}

// get config string value
func (this *InkConfig) String(key string) string {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return ""
	}
	str, ok := (*this)[keys[0]][keys[1]]
	if !ok {
		return ""
	}
	return fmt.Sprint(str)
}

// get config string value
// if empty, return replaced value
func (this *InkConfig) StringOr(key string, def string) string {
	value := this.String(key)
	if value == "" {
		return def
	}
	return value
}

// get config int value
func (this *InkConfig) Int(key string) int {
	str := this.String(key)
	i, _ := strconv.Atoi(str)
	return i
}

// get config int value
// if zero, return replaced value
func (this *InkConfig) IntOr(key string, def int) int {
	i := this.Int(key)
	if i == 0 {
		return def
	}
	return i
}

// get config float value
func (this *InkConfig) Float(key string) float64 {
	str := this.String(key)
	f, _ := strconv.ParseFloat(str, 64)
	return f
}

// get config float value
// if 0.0, return replaced value
func (this *InkConfig) FloatOr(key string, def float64) float64 {
	f := this.Float(key)
	if f == 0.0 {
		return def
	}
	return f
}

// get config bool value
func (this *InkConfig) Bool(key string) bool {
	str := this.String(key)
	b, _ := strconv.ParseBool(str)
	return b
}

// get config map value
func (this *InkConfig) Map(key string) map[string]interface{} {
	return (*this)[key]
}

// set config value by name
func (this *InkConfig) Set(key string, value interface{}) {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return
	}
	(*this)[keys[0]][keys[1]] = value
}

// set config map
func (this *InkConfig) SetMap(key string, value map[string]interface{}) {
	(*this)[key] = value
}

// append config object into this one
func (this *InkConfig) Append(value *InkConfig) {
	for k, v := range *value {
		(*this)[k] = v
	}
}

// create new config object
func NewConfig(fileName string, configType int) (*InkConfig, error) {
	bytes, e := ioutil.ReadFile(fileName)
	if e != nil {
		return nil, e
	}
	fn := configParseFuncMap[configType]
	if fn == nil {
		return nil, errors.New("invalid configuration file type")
	}
	return fn(bytes)
}

// set config parse function
func SetConfigParserFunction(configType int, fn func (bytes []byte) (*InkConfig, error)) {
	configParseFuncMap[configType] = fn
}

// parse json bytes
func parseJsonConfigBytes(jsonBytes []byte) (*InkConfig, error) {
	var config InkConfig
	e := json.Unmarshal(jsonBytes, &config)
	return &config, e
}
