package lib

import ()

type Params map[string]string

const ReservedParam = "all" 

const RootTableParam = "table" 
const RootToTableParam = "totable" 
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