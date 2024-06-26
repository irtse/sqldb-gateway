package utils
import ( "fmt"; "strconv"; "sync"; "strings" )
// API COMMON query params !
var ParamsMutex = &sync.Mutex{}

type Params map[string]string
const ReservedParam = "all" // IMPORTANT IS THE DEFAULT PARAMS FOR ROWS & COLUMNS PARAMS
const RootTableParam = "table" 
const RootRowsParam = "rows" 
const RootColumnsParam = "columns" 
const RootOrderParam = "orderby" 
const RootDirParam = "dir"
const RootFilterNewState = "filter_new" // all - new - old	
const RootFilterLine = "filter_line" // + == "and" | == "or" ~ == "like" : == "=" > == ">" < == "<"
const RootFilter = "filter"
const RootViewFilter = "view_filter"
const RootRawView = "rawview"
const RootExport = "export"
const RootFilename = "filename"
const RootShallow = "shallow"
const RootSuperCall = "super"
const RootDestTableIDParam = "dbdest_table_id" 
const RootCommandRow = "command_row"
const RootCommandCols = "command_columns"
const RootLimit= "limit"
const RootOffset= "offset"

var RootParamsDesc = map[string]string{
	RootRowsParam : "needed on a rows level request (value=all for post/put method or a get/delete all)",
    RootColumnsParam : "needed on a columns level request (POST/PUT/DELETE with no rows query params) will set up a view on row level (show only expected columns)",
	RootShallow : "activate a lightest response (name only)",
	RootOrderParam : "sets up a sql order in query (don't even try to inject you little joker)",
	RootDirParam : "sets up a sql direction in query (ex.ASC|DESC) (don't even try to inject you little joker)",
	RootRawView : "set 'enable' to activate a response without the main response format",
	RootFilterLine : "set a filter command line compose as (key~value(+|))",
	RootFilter : "set filter identifier to activate a specific restrictive filter in db",
	RootViewFilter : "set view filter identifier to activate a specific view filter in db",
}
var HiddenParams = []string{RootDestTableIDParam}
var RootParams = []string{RootRowsParam, RootColumnsParam, RootOrderParam, RootDirParam, RootLimit, RootOffset, RootShallow, RootRawView, RootExport, RootFilename, RootFilterNewState, RootFilterLine, RootFilter, RootViewFilter, RootSuperCall, RootCommandRow, SpecialIDParam}

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
	AVG Method = 6
	MIN Method = 7
	MAX Method = 8
	SUM Method = 9
)
func Found(name string) Method {
	switch strings.ToLower(name) {
		case "read": return SELECT
		case "write": return CREATE
		case "update": return UPDATE
		case "delete": return DELETE
		case "count": return COUNT
		case "avg": return AVG
		case "min": return MIN
		case "max": return MAX
		case "sum": return SUM
	}
	return SELECT
}
func (s Method) String() string {
	switch s {
		case SELECT: return "read"
		case CREATE: return "write"
		case UPDATE: return "update"
		case DELETE: return "delete"
		case COUNT: return "count"
		case AVG: return "avg"
		case MIN: return "min"
		case MAX: return "max"
		case SUM: return "sum"
	}
	return "unknown"
}

func (s Method) IsMath() bool {
	switch s {
		case COUNT, AVG, MIN, MAX, SUM: return true
	}
	return false
}

func (s Method) Method() string {
	switch s {
		case SELECT: return "get"
		case CREATE: return "post"
		case UPDATE: return "put"
		case DELETE: return "delete"
		case COUNT: return "count"
		case AVG: return "avg"
		case MIN: return "min"
		case MAX: return "max"
		case SUM: return "sum"
	}
	return "unknown"
}

func (s Method) Calling() string {
	switch s {
		case SELECT: return "Get"
		case CREATE: return "Create"
		case UPDATE: return "Update"
		case DELETE: return "Delete"
		case COUNT: return "Math"
		case AVG: return "Math"
		case MIN: return "Math"
		case MAX: return "Math"
		case SUM: return "Math"
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