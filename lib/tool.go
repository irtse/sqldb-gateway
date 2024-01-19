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

type Record map[string]interface{}

func (ar *Record) GetString(column string) string {
	str := fmt.Sprintf("%v", (*ar)[column])
	return str
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

type Results []Record  

type SpecializedServiceInfo interface { GetName() string }
type SpecializedService interface {
	Entity() SpecializedServiceInfo
	WriteRowWorkflow(record Record)
	VerifyRowWorkflow(record Record, create bool) (Record, bool)
	DeleteRowWorkflow(results Results)
	UpdateRowWorkflow(results Results, record Record) 
}
type CustomService struct { SpecializedService }

func (s *CustomService) UpdateRowWorkflow(results Results, record Record) {}
func (s *CustomService) WriteRowWorkflow(record Record) {}
func (s *CustomService) DeleteRowWorkflow(results Results) { }
func (s *CustomService) Entity() SpecializedServiceInfo { return nil }
func (s *CustomService) VerifyRowWorkflow(record Record, create bool) (Record, bool) { return record, true }

type Params map[string]string

const ReservedParam = "all" 

const RootTableParam = "table" 
const RootRowsParam = "rows" 
const RootColumnsParam = "columns" 
const RootOrderParam = "orderby" 
const RootDirParam = "dir"
const RootSQLFilterParam = "sqlfilter" 

var RootParamsDesc = map[string]string{
	RootRowsParam : "needed on a rows level request (value=all for post/put method or a get/delete all)",
    RootColumnsParam : "needed on a columns level request (POST/PUT/DELETE with no rows query params)",
}

var RootParams = []string{RootRowsParam, RootColumnsParam, RootOrderParam, RootDirParam, RootSQLFilterParam}

const SpecialIDParam = "id" 
const SpecialModeParam = "mode" 