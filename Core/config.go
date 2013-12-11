package Core

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

type Config map[string]map[string]interface{}

// get config string value.
func (this *Config) String(key string) string {
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

// get config string value.
// if empty, return replaced value
func (this *Config) StringOr(key string, def string) string {
	value := this.String(key)
	if value == "" {
		return def
	}
	return value
}

// get config int value.
func (this *Config) Int(key string) int {
	str := this.String(key)
	i, _ := strconv.Atoi(str)
	return i
}

// get config int value.
// if zero, return replaced value.
func (this *Config) IntOr(key string, def int) int {
	i := this.Int(key)
	if i == 0 {
		return def
	}
	return i
}

// get config float value.
func (this *Config) Float(key string) float64 {
	str := this.String(key)
	f, _ := strconv.ParseFloat(str, 64)
	return f
}

// get config float value.
// if 0.0, return replaced value.
func (this *Config) FloatOr(key string, def float64) float64 {
	f := this.Float(key)
	if f == 0.0 {
		return def
	}
	return f
}

// get config bool value.
func (this *Config) Bool(key string) bool {
	str := this.String(key)
	b, _ := strconv.ParseBool(str)
	return b
}

// set config value by name.
func (this *Config) Set(key string, value interface{}) {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return
	}
	(*this)[keys[0]][keys[1]] = value
}

/* ----------------------------------------*/

// defined config parser function.
// parsing function should parse []byte to map[string]map[string]interface{}
var configFunc map[string]func(bytes []byte) (*Config, error)

func init() {
	configFunc = make(map[string]func(bytes []byte) (*Config, error))
	// add json func as default
	configFunc["json"] = func(bytes []byte) (*Config, error) {
		var config Config
		e := json.Unmarshal(bytes, &config)
		return &config, e
	}
}

// add new config parse function.
func NewConfigFunc(name string, fn func(bytes []byte) (*Config, error)) {
	configFunc[name] = fn
}

// create new config with bytes by config type name
func NewConfig(data []byte, name string) (*Config, error) {
	if configFunc[name] == nil {
		return nil, errors.New("unknown configuration type: "+name)
	}
	return configFunc[name](data)
}

// create new config from file by config type name.
func NewConfigFromFile(fileAbsPath string, name string) (*Config, error) {
	bytes, e := ioutil.ReadFile(fileAbsPath)
	if e != nil {
		return nil, e
	}
	return NewConfig(bytes, name)
}
