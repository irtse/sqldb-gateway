package domain

import (
	"fmt"
	"strings"
	"encoding/json"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
	conn "sqldb-ws/lib/infrastructure/connector"
)
// main func delete Row in user Entry
func (d *MainService) DeleteRow(tableName string, results tool.Results) {
	/*for _, record := range results {
		params := tool.Params{ tool.RootTableParam : entities.DBUserEntry.Name, 
			                   tool.RootRowsParam: fmt.Sprintf("%v", record[tool.SpecialIDParam]), }
		d.SuperCall(params, tool.Record{}, tool.DELETE, "Delete")
	}*/	
}
// main func add Row in user Entry
func (d *MainService) WriteRow(tableName string, record tool.Record) {}
// define filter whatever what happen on sql...
func (d *MainService) ViewDefinition(tableName string, innerRestriction... string) (string, string) {
	SQLview := ""; SQLrestriction := ""; auth := true
	if d.IsRawView() && d.IsSuperCall() { 
		return SQLrestriction, SQLview // admin can see all on admin view
	}
	if d.Method == tool.SELECT {
		for _, exception := range entities.PERMISSIONEXCEPTION {
			if tableName == exception.Name { auth = false; break }
		}
	}
	SQLrestriction = d.ByEntityUser(tableName)
	if auth { 
		restr, v := d.byFields(tableName) 
		SQLview = v
		if len(strings.TrimSpace(restr)) > 0 {
			if len(SQLrestriction) > 0 { SQLrestriction += " AND " + restr
	    	} else {  SQLrestriction = restr }
		}
	}
	if len(innerRestriction) > 0 {
		for _, restr := range innerRestriction {
			if len(strings.TrimSpace(restr)) > 0 {
				if len(SQLrestriction) > 0 { SQLrestriction += " AND " + restr 
				} else { SQLrestriction = restr  }
			}
		}
	}
	return SQLrestriction, SQLview
}

func (s *MainService) ByEntityUser(tableName string) (string) {
	schemas, err := s.Schema(tool.Record{ entities.NAMEATTR : tableName }, false)
	restr := ""
	if err != nil && schemas != nil && len(schemas) > 0 { 
		userID := entities.RootID(entities.DBUser.Name)
		entityID := entities.RootID(entities.DBEntity.Name)
		if _, ok := schemas[0][userID]; ok || tableName == entities.DBUser.Name {
			if tableName == entities.DBUser.Name  { userID = tool.SpecialIDParam }
			restr := userID + " IN (SELECT id FROM " + entities.DBUser.Name + " WHERE name=" + conn.Quote(s.GetUser()) + " OR email=" + conn.Quote(s.GetUser()) + ")" 
			if _, ok := schemas[0][entityID]; ok || tableName == entities.DBEntity.Name  {
				if tableName == entities.DBEntity.Name  { entityID = tool.SpecialIDParam }
				if len(restr) > 0 { restr +=  " OR " }
				restr += entityID + " IN (SELECT " + entityID + " FROM " + entities.DBEntityUser.Name + "WHERE " + entities.RootID(entities.DBUser.Name) + " IN (" 
				restr += "SELECT id FROM " + userID + " WHERE name=" + conn.Quote(s.GetUser()) + " OR email=" + conn.Quote(s.GetUser()) + ")"
			}
		}
	}
	return restr
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
	return "", SQLview
}

func (d *MainService) Schema(record tool.Record, permitted bool) (tool.Results, error) { // check schema auth access
	params := tool.Params{ tool.RootTableParam : entities.DBSchema.Name, 
			               tool.RootRowsParam : tool.ReservedParam }
	sqlFilter := ""
	// fmt.Printf("RECORD %v \n ", record)
	if id, ok := record[entities.RootID(entities.DBSchema.Name)]; ok {
		sqlFilter += "id=" + fmt.Sprintf("%v", id)
	} else if name, ok := record[entities.NAMEATTR]; ok {
		sqlFilter += "name='" + fmt.Sprintf("%v", name) + "'"
	}
	return d.SuperCall( params, tool.Record{}, tool.SELECT, "Get", sqlFilter)
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