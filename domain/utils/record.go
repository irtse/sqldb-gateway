package utils

import (
	"fmt"
	"strconv"
	// API COMMON query params !
)

var MAIN_PREFIX = "generic"

type Results []Record
type Record map[string]interface{}

func AddMap(rec map[string]interface{}, add ...map[string]interface{}) map[string]interface{} {
	for _, res := range add {
		for k, v := range res {
			rec[k] = v
		}
	}
	return rec
}

func ToRecord(add ...map[string]interface{}) Record {
	rec := Record{}
	for _, res := range add {
		for k, v := range res {
			rec[k] = v
		}
	}
	return rec
}

func Add(m map[string]interface{}, k string, v interface{}, condition func(interface{}) bool,
	transform func(interface{}) interface{}) {
	if condition(v) {
		m[k] = transform(v)
	}
}

func (ar Record) Add(k string, v interface{}, condition func() bool) {
	if condition() {
		ar[k] = v
	}
}

func (ar Record) Copy() Record {
	new := Record{}
	for k, v := range ar {
		new[k] = v
	}
	return new
}

func (ar *Record) GetString(column string) string {
	if (*ar)[column] == nil {
		return ""
	}
	return fmt.Sprintf("%v", (*ar)[column])
}

func GetString(record map[string]interface{}, column string) string {
	if record[column] == nil {
		return ""
	}
	return fmt.Sprintf("%v", record[column])
}

func GetInt(record map[string]interface{}, column string) int64 {
	str := fmt.Sprintf("%v", record[column])
	val, _ := strconv.Atoi(str)
	return int64(val)
}

func (ar *Record) GetInt(column string) int64 {
	str := fmt.Sprintf("%v", (*ar)[column])
	val, _ := strconv.Atoi(str)
	return int64(val)
}

func (ar *Record) GetFloat(column string) float64 {
	str := fmt.Sprintf("%v", (*ar)[column])
	val, _ := strconv.ParseFloat(str, 64)
	return val
}
