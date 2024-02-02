package domain

import (
	"fmt"
	"strings"
	"errors"
	"encoding/json"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
)
// main func delete Row in user Entry
func (d *MainService) DeleteRow(tableName string, results tool.Results) {
	for _, record := range results {
		params := tool.Params{ tool.RootTableParam : entities.DBUserEntry.Name, 
			                   tool.RootRowsParam: fmt.Sprintf("%v", record[tool.SpecialIDParam]), }
		d.SuperCall(params, tool.Record{}, tool.DELETE, "Delete")
	}	
}
// main func add Row in user Entry
func (d *MainService) WriteRow(tableName string, record tool.Record) {
	params := tool.Params{  tool.RootTableParam : entities.DBUser.Name, 
		                    tool.RootRowsParam:  tool.ReservedParam, "login" : d.GetUser() }
	// retrieve user your proper user
	users, err := d.SuperCall(params, tool.Record{},tool.SELECT, "Get") // by super procedure in domain
	if err != nil || len(users) == 0 { return }
	params = tool.Params{ tool.RootTableParam : entities.DBSchema.Name, 
		             tool.RootRowsParam:  tool.ReservedParam,
					 entities.NAMEATTR : tableName }
	// retrieve table schema
	schemas, err := d.SuperCall(params, tool.Record{},tool.SELECT, "Get") // by super procedure in domain
	if err != nil || len(schemas) == 0 { return }
	params = tool.Params{  tool.RootTableParam : entities.DBUserEntry.Name, 
		                   tool.RootRowsParam:  tool.ReservedParam, }
	// then create link between schema/user/new row
	users, err = d.SuperCall(params, tool.Record{
		entities.RootID(entities.DBSchema.Name) : schemas[0][tool.SpecialIDParam],
		entities.RootID(entities.DBUser.Name) : users[0][tool.SpecialIDParam],
		entities.RootID("dest_table") : record[tool.SpecialIDParam],
	},tool.CREATE, "CreateOrUpdate")
}
// define filter whatever what happen on sql...
func (d *MainService) ViewDefinition(tableName string, innerRestriction... string) (string, string) {
	SQLview := ""; SQLrestriction := ""; auth := true
	if d.IsRawView() && d.SuperAdmin { 
		return SQLrestriction, SQLview // admin can see all on admin view
	}
	for _, exception := range entities.PERMISSIONEXCEPTION {
		if tableName == exception.Name { auth = false; break }
	}
	if auth { SQLrestriction, SQLview = d.byFields(tableName) }
	if len(innerRestriction) > 0 {
		for _, restr := range innerRestriction {
			if len(SQLrestriction) > 0 { SQLrestriction += " AND " + restr 
			} else { SQLrestriction = restr  }
		}
	}
	return SQLrestriction, SQLview
}

func (d *MainService) byFields(tableName string) (string, string) {
	SQLview := ""
	for _, restricted := range entities.DBRESTRICTED {
		if restricted.Name == tableName { return "id=-1", "" }
	}
	p := tool.Params{ tool.RootTableParam : entities.DBSchema.Name,
	                  tool.RootRowsParam : tool.ReservedParam,
				      entities.NAMEATTR : tableName }
	schemas, err := d.SuperCall( p, tool.Record{}, tool.SELECT, "Get")
	if err != nil || len(schemas) == 0 { return "id=-1", "" }
	d.SuperCall( p, tool.Record{}, tool.SELECT, "Get")
	p = tool.Params{ tool.RootTableParam : entities.DBSchemaField.Name,
	                 tool.RootRowsParam : tool.ReservedParam,
				     entities.RootID(entities.DBSchema.Name) : fmt.Sprintf("%v", schemas[0][tool.SpecialIDParam]) }
	fields, err := d.SuperCall( p, tool.Record{}, tool.SELECT, "Get")
	if err != nil || len(fields) == 0 { return "id=-1", "" }
	for _, field := range fields {
		if name, okName := field["name"]; !okName {
			n := fmt.Sprintf("%v", name)
			if hide, ok := field["hidden"]; (!ok || !hide.(bool)) && n != "id" { SQLview += n + "," }
		}
	}
	return "", SQLview
}

func (d *MainService) Schema(record tool.Record) (tool.Results, error) { // check schema auth access
	if schemaID, ok := record[entities.RootID(entities.DBSchema.Name)]; ok {
		params := tool.Params{ tool.RootTableParam : entities.DBSchema.Name, 
			              tool.RootRowsParam : fmt.Sprintf("%v", schemaID), 
		                }
		schemas, err := d.SuperCall( params, tool.Record{}, tool.SELECT, "Get")
		if err != nil || len(schemas) == 0 { return nil, err }
		if _, ok := d.GetPermission().Verify(schemas[0][entities.NAMEATTR].(string)); !ok { 
			return nil, errors.New("not authorized ") 
		}
		return schemas, nil
	}
	return nil, errors.New("no schemaID refered...")
}

type Filter struct {
	Extra				string 		 			 	 `json:"extra_path"`
	Dir					string 		 			 	 `json:"sql_dir"`
	Order 	 			string 		 			 	 `json:"sql_order"`
	View				string 		 			 	 `json:"sql_view"`
	ViewID 				int64 		 			 	 `json:"dbview_id"`
}

func (d *MainService) GeneratePathFilter(path string, record tool.Record, params tool.Params) (string, tool.Params) { // check schema auth access
	var filter Filter
	b, _:= json.Marshal(record)
	json.Unmarshal(b, &filter) // GET ITS OWN FILTER
	if filter.Extra != "" { 
		path += "/" + filter.Extra
	}
	if filter.View != "" { 
		if params != nil { params[tool.RootColumnsParam] = filter.View }
		path += "&" + tool.RootColumnsParam + "=" + filter.View
	}
	if filter.Order != "" { 
		if params != nil { params[tool.RootOrderParam] =  filter.Order }
		path += "&" + tool.RootOrderParam + "=" + strings.Replace(filter.Order, " ", "+", -1)
	}
	if filter.Dir != "" { 
		if params != nil { params[tool.RootDirParam] = filter.Dir }
		path += "&" + tool.RootDirParam + "=" + strings.Replace(filter.Dir, " ", "+", -1)
	}
	if filter.ViewID > 0 {
		path = "/" + entities.DBView.Name +"?" + tool.RootRowsParam + "=" + fmt.Sprintf("%v", filter.ViewID)
		delete(record, entities.RootID(entities.DBView.Name))
	}
	return path, params
}