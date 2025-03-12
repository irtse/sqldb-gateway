package utils

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func PrepareEnum(enum string) string {
	e := strings.Replace(ToString(enum), " ", "", -1)
	e = strings.Replace(e, "'", "", -1)
	e = strings.Replace(e, "(", "__", -1)
	e = strings.Replace(e, ",", "_", -1)
	e = strings.Replace(e, ")", "", -1)
	return strings.ToLower(e)
}

func ToMap(who interface{}) map[string]interface{} {
	if reflect.TypeOf(who).Kind() == reflect.Map {
		return who.(map[string]interface{})
	}
	return map[string]interface{}{}
}

func ToList(who interface{}) []interface{} {
	if reflect.TypeOf(who).Kind() == reflect.Slice {
		return who.([]interface{})
	}
	return []interface{}{}
}

func ToInt64(who interface{}) int64 {
	if who == nil {
		return 0
	}
	i, err := strconv.Atoi(fmt.Sprintf("%v", who))
	if err != nil {
		return 0
	}
	return int64(i)
}

func ToString(who interface{}) string {
	if who == nil {
		return ""
	}
	return fmt.Sprintf("%v", who)
}

func Compare(who interface{}, what interface{}) bool {
	return who != nil && fmt.Sprintf("%v", who) == fmt.Sprintf("%v", what)
}
