package utils

import (
	"fmt"
	"sqldb-ws/infrastructure/connector"
	"strings"
	"sync"
)

var ParamsMutex = &sync.Mutex{}

type Params map[string]string

func (p Params) GetAsArgs(key string) []string {
	if arg, ok := p[key]; ok {
		return []string{arg}
	}
	return []string{}
}

func (p Params) GetOrder(condition func(string) bool, order []string) []string {
	if orderBy, ok := p[RootOrderParam]; ok {
		direction := []string{}
		if dir, ok2 := p[RootDirParam]; ok2 {
			direction = strings.Split(fmt.Sprintf("%v", dir), ",")
		}
		for i, el := range strings.Split(fmt.Sprintf("%v", orderBy), ",") {
			if (!condition(el) && el != SpecialIDParam) || len(direction) <= i {
				continue
			} // ???
			upper := strings.Replace(strings.ToUpper(direction[i]), " ", "", -1)
			if upper == "ASC" || upper == "DESC" {
				order = append(order, connector.SQLInjectionProtector(el+" "+upper))
				continue
			}
			order = append(order, connector.SQLInjectionProtector(el+" ASC"))
		}
	} else {
		return []string{"id DESC"}
	}
	return order
}

func (p Params) GetLimit(limited string) string {
	if limit, ok := p[RootLimit]; ok {
		if offset, ok2 := p[RootOffset]; ok2 {
			return connector.FormatLimit(limit, offset)
		}
		return connector.FormatLimit(limit, "")
	}
	return limited
}

func (p Params) RootRaw() Params {
	p[RootRawView] = "enable"
	return p
}

func (p Params) Enrich(param map[string]interface{}) Params {
	for k, v := range param {
		p[k] = fmt.Sprintf("%v", v)
	}
	return p
}

func (p Params) Add(k string, val interface{}, condition func(string) bool) {
	if val == nil || val == "" || !condition(k) {
		return
	}
	p[k] = fmt.Sprintf("%v", val)
}

func (p Params) AddMap(vals map[string]interface{}, condition func(string) bool) {
	for k, val := range vals {
		p.Add(k, val, condition)
	}
}

func (p Params) UpdateParamsWithFilters(view, dir string) {
	if view != "" {
		p[RootColumnsParam] = view
	}
	if dir != "" {
		p[RootDirParam] = dir
	}
}

func (p Params) EnrichCondition(flat map[string]string, condition func(string) bool) Params {
	for k, v := range flat {
		if condition(k) {
			if k == SpecialSubIDParam {
				k = SpecialIDParam
			}
			p[k] = v
		}
	}
	return p
}

func (p Params) Anonymized() map[string]interface{} {
	newM := map[string]interface{}{}
	for k, v := range p {
		newM[k] = v
	}
	return newM
}

func (p Params) Delete(condition func(string) bool) Params {
	toDelete := []string{}
	for k := range p {
		if condition(k) {
			toDelete = append(toDelete, k)
		}
	}
	for _, k := range toDelete {
		delete(p, k)
	}
	return p
}

func AllParams(table string) Params {
	return Params{RootTableParam: table, RootRowsParam: ReservedParam}
}

func GetTableTargetParameters(tableName interface{}) Params {
	if tableName == nil {
		return Params{}
	}
	return Params{RootTableParam: fmt.Sprintf("%v", tableName)}
}

func GetColumnTargetParameters(tableName interface{}, col interface{}) Params {
	if col == nil || fmt.Sprintf("%v", col) == "" {
		col = ReservedParam
	}
	return Params{RootTableParam: fmt.Sprintf("%v", tableName), RootColumnsParam: fmt.Sprintf("%v", col)}
}

func GetRowTargetParameters(tableName interface{}, row interface{}) Params {
	if row == nil || fmt.Sprintf("%v", row) == "" {
		row = ReservedParam
	}
	return Params{RootTableParam: fmt.Sprintf("%v", tableName), RootRowsParam: fmt.Sprintf("%v", row)}
}

const ReservedParam = "all" // IMPORTANT IS THE DEFAULT PARAMS FOR ROWS & COLUMNS PARAMS
const RootTableParam = "table"
const RootRowsParam = "rows"
const RootColumnsParam = "columns"
const RootOrderParam = "orderby"
const RootDirParam = "dir"
const RootFilterNewState = "filter_new" // all - new - old
const RootFilterLine = "filter_line"    // + == "and" | == "or" ~ == "like" : == "=" > == ">" < == "<"
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
const RootLimit = "limit"
const RootOffset = "offset"

var RootParamsDesc = map[string]string{
	RootRowsParam:    "needed on a rows level request (value=all for post/put method or a get/delete all)",
	RootColumnsParam: "needed on a columns level request (POST/PUT/DELETE with no rows query params) will set up a view on row level (show only expected columns)",
	RootShallow:      "activate a lightest response (name only)",
	RootOrderParam:   "sets up a sql order in query (don't even try to inject you little joker)",
	RootDirParam:     "sets up a sql direction in query (ex.ASC|DESC) (don't even try to inject you little joker)",
	RootRawView:      "set 'enable' to activate a response without the main response format",
	RootFilterLine:   "set a filter command line compose as (key~value(+|))",
	RootFilter:       "set filter identifier to activate a specific restrictive filter in db",
	RootViewFilter:   "set view filter identifier to activate a specific view filter in db",
}
var HiddenParams = []string{RootDestTableIDParam}
var RootParams = []string{RootRowsParam, RootColumnsParam, RootOrderParam, RootDirParam, RootLimit, RootOffset,
	RootShallow, RootRawView, RootExport, RootFilename, RootFilterNewState, RootFilterLine, RootFilter, RootViewFilter,
	RootSuperCall, RootCommandRow, SpecialIDParam}

const SpecialIDParam = "id"
const SpecialSubIDParam = "subid"
