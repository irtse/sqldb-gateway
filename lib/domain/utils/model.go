package utils
import ( "fmt"; "strconv"; "sync" )
// API COMMON query params !
var ParamsMutex = &sync.Mutex{}

type Params map[string]string
const ReservedParam = "all" // IMPORTANT IS THE DEFAULT PARAMS FOR ROWS & COLUMNS PARAMS
const RootTableParam = "table" 
const RootRowsParam = "rows" 
const RootColumnsParam = "columns" 
const RootOrderParam = "orderby" 
const RootDirParam = "dir"
const RootFilter = "filter"
const RootViewFilter = "view_filter"
const RootRawView = "rawview"
const RootExport = "export"
const RootFilename = "filename"
const RootShallow = "shallow"
const RootSuperCall = "super"
const RootDestTableIDParam = "dbdest_table_id" 

const RootLimit= "limit"
const RootOffset= "offset"

var RootParamsDesc = map[string]string{
	RootRowsParam : "needed on a rows level request (value=all for post/put method or a get/delete all)",
    RootColumnsParam : "needed on a columns level request (POST/PUT/DELETE with no rows query params) will set up a view on row level (show only expected columns)",
	RootShallow : "activate a lightest response (name only)",
	RootOrderParam : "sets up a sql order in query (don't even try to inject you little joker)",
	RootDirParam : "sets up a sql direction in query (ex.ASC|DESC) (don't even try to inject you little joker)",
	RootRawView : "set 'enable' to activate a response without the main response format",
	RootFilter : "set filter identifier to activate a specific restrictive filter",
	RootViewFilter : "set view filter identifier to activate a specific view filter",
}
var HiddenParams = []string{RootDestTableIDParam}
var RootParams = []string{RootRowsParam, RootColumnsParam, RootOrderParam, RootDirParam, RootLimit, RootOffset, RootShallow, RootRawView, RootExport, RootFilename, RootFilter, RootViewFilter, RootSuperCall, SpecialIDParam}

const SpecialIDParam = "id" 
const SpecialSubIDParam = "subid" 
var MAIN_PREFIX = "generic"

func AllParams(table string) Params { return Params{ RootTableParam : table, RootRowsParam : ReservedParam } }

type Method int64
const(
	SELECT Method = 1
	CREATE Method = 2
	UPDATE Method = 3
	DELETE Method = 4
	COUNT Method = 5
)
func (s Method) String() string {
	switch s {
		case SELECT: return "read"
		case CREATE: return "write"
		case UPDATE: return "update"
		case DELETE: return "delete"
		case COUNT: return "count"
	}
	return "unknown"
}

func (s Method) Method() string {
	switch s {
		case SELECT: return "get"
		case CREATE: return "post"
		case UPDATE: return "put"
		case DELETE: return "delete"
		case COUNT: return "count"
	}
	return "unknown"
}

func (s Method) Calling() string {
	switch s {
		case SELECT: return "Get"
		case CREATE: return "Create"
		case UPDATE: return "Update"
		case DELETE: return "Delete"
		case COUNT: return "Count"
	}
	return "unknown"
}

type Results []Record  
type Record map[string]interface{}

func (ar *Record) GetString(column string) string {
	return fmt.Sprintf("%v", (*ar)[column])
}

func GetString(record map[string]interface{}, column string) string {
	return fmt.Sprintf("%v", record[column])
}

func GetInt(record map[string]interface{},column string) int64 {
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