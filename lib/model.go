package lib

import (
	"fmt"
	"strconv"
)

type Method int64
const(
	SELECT Method = 1
	CREATE Method = 2
	UPDATE Method = 3
	DELETE Method = 4
)
func (s Method) String() string {
	switch s {
		case SELECT: return "read"
		case CREATE: return "create"
		case UPDATE: return "update"
		case DELETE: return "delete"
	}
	return "unknown"
}

type Results []Record  
type Record map[string]interface{}

func (ar *Record) GetString(column string) string {
	return fmt.Sprintf("%v", (*ar)[column])
}

func (ar *Record) GetInt(column string) int {
	str := fmt.Sprintf("%v", (*ar)[column])
	val, _ := strconv.Atoi(str)
	return val
}

func (ar *Record) GetFloat(column string) float64 {
	str := fmt.Sprintf("%v", (*ar)[column])
	val, _ := strconv.ParseFloat(str, 64)
	return val
}

