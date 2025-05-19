package utils

import (
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

func ToListStr(add []interface{}) []string {
	rec := []string{}
	for _, res := range add {
		rec = append(rec, ToString(res))
	}
	return rec
}

func ToResult(add []map[string]interface{}) Results {
	rec := Results{}
	for _, res := range add {
		rec = append(rec, ToRecord(res))
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
	return ToString((*ar)[column])
}

func GetBool(record map[string]interface{}, column string) bool {
	if record[column] == nil {
		return false
	}
	return ToString(record[column]) == "true"
}

func GetString(record map[string]interface{}, column string) string {
	if record[column] == nil {
		return ""
	}
	return ToString(record[column])
}

func GetFloat(record map[string]interface{}, column string) float64 {
	str := ToString(record[column])
	if str == "" {
		return -1
	}
	val, _ := strconv.ParseFloat(str, 64)
	return float64(val)
}

func GetInt(record map[string]interface{}, column string) int64 {
	str := ToString(record[column])
	if str == "" {
		return -1
	}
	val, _ := strconv.Atoi(str)
	return int64(val)
}

func (ar *Record) GetBool(column string) bool {
	str := ToString((*ar)[column])
	return str == "true"
}

func (ar *Record) GetInt(column string) int64 {
	str := ToString((*ar)[column])
	if str == "" {
		return -1
	}
	val, _ := strconv.Atoi(str)
	return int64(val)
}

func (ar *Record) GetFloat(column string) float64 {
	str := ToString((*ar)[column])
	val, _ := strconv.ParseFloat(str, 64)
	return val
}
