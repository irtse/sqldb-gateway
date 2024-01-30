package lib

// All common query params !
type Params map[string]string

const ReservedParam = "all" // IMPORTANT IS THE DEFAULT PARAMS FOR ROWS & COLUMNS PARAMS

const RootTableParam = "table" 
const RootToTableParam = "totable" 
const RootRowsParam = "rows" 
const RootColumnsParam = "columns" 
const RootOrderParam = "orderby" 
const RootDirParam = "dir"
const RootAdminView = "adminview"
const RootShallow = "shallow"
const RootSQLFilterParam = "sqlfilter" 

var RootParamsDesc = map[string]string{
	RootRowsParam : "needed on a rows level request (value=all for post/put method or a get/delete all)",
    RootColumnsParam : "needed on a columns level request (POST/PUT/DELETE with no rows query params) will set up a view on row level (show only expected columns)",
	RootShallow : "activate a lightest response (name only)",
	RootOrderParam : "sets up a sql order in query",
	RootDirParam : "sets up a sql direction in query (ex.ASC)",
	RootSQLFilterParam : "sets up a sql filter in query",
	RootAdminView : "activate a response without the main response format (only available if super admin)",
}

var RootParams = []string{RootRowsParam, RootColumnsParam, RootOrderParam, RootDirParam, RootSQLFilterParam, RootShallow, RootAdminView}

const SpecialIDParam = "id" 