package domain

import (
	"fmt"
	"errors"
	"strings"
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
func (d *MainService) ViewDefinition(tableName string, params tool.Params) (string, string) {
	SQLview := ""; SQLrestriction := ""; auth := true
	if admin, ok := params[tool.RootAdminView]; ok && admin == "enable" && d.SuperAdmin { 
		return SQLrestriction, SQLview // admin can see all on admin view
	}
	for _, exception := range entities.PERMISSIONEXCEPTION {
		if tableName == exception.Name { auth = false; break }
	}
	if auth { SQLrestriction, SQLview = d.byFields(tableName) }
	if filter, ok := params[tool.RootSQLFilterParam]; ok {
		if len(SQLrestriction) > 0 { SQLrestriction += " AND " + filter 
	    } else { SQLrestriction = filter  }
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

type View struct {
	Name  		 string 				`json:"name"`
	Description  string 				`json:"description"`
	Path		 string 				`json:"path"`
	Schema		 map[string]interface{} `json:"schema"`
	Items		 []ViewItem				`json:"items"`
}

type ViewItem struct {
	Values 		map[string]interface{} 		   `json:"values"`
	Data 		map[string]interface{} 	 	   `json:"data"`
	Contents 	map[string]interface{} 	 	   `json:"contents"`
}

func (d *MainService) getPath(tableName string, view string, restr string, 
	                          order string, dir string, additonnalRestriction ...string) string {
	path := "/" + tableName + "?rows=all"
	if view != "" { path += "&" + tool.RootColumnsParam + "=" + strings.TrimSpace(view) }
	if strings.TrimSpace(restr) != "" || len(additonnalRestriction) > 0 {
		base := "&" + tool.RootSQLFilterParam + "="
		ext := ""
		if restr != "" { ext += strings.Replace(strings.TrimSpace(restr), " ", "+", -1) + "+" }
		for _, add := range additonnalRestriction { 
			if add != "" { 
				if len(ext) > 0 { ext += strings.Replace(" AND " + add, " ", "+", -1)
				} else { ext += strings.Replace(add, " ", "+", -1) }
			}
		}
		if ext != "" { path += base + ext }
	}
	if strings.TrimSpace(order) != "" { path += "&" + tool.RootOrderParam + "=" + strings.Replace(strings.TrimSpace(order), " ", "+", -1) }
	if strings.TrimSpace(dir) != "" { path += "&" + tool.RootDirParam + "=" + strings.Replace(strings.TrimSpace(dir), " ", "+", -1) }
	return path
}

func (d *MainService) PostTreat(results tool.Results, tableName string, shallow bool, additonnalRestriction ...string) tool.Results {
	res := tool.Results{}
	cols := map[string]entities.SchemaColumnEntity{}
	sqlFilter := entities.RootID(entities.DBSchema.Name) + " IN (SELECT id FROM "
	sqlFilter += entities.DBSchema.Name + " WHERE name='" + tableName + "')"
	// retrive all fields from schema...
	params := tool.Params{ tool.RootTableParam : entities.DBSchemaField.Name, 
		                   tool.RootRowsParam: tool.ReservedParam, 
						   tool.RootSQLFilterParam: sqlFilter }
	schemas, err := d.SuperCall( params, tool.Record{}, tool.SELECT, "Get")
	if err != nil || len(schemas) == 0 { return res }
	var view View
	if !d.IsShallowed() {
		schemes := map[string]interface{}{}
		for _, r := range schemas {
			var scheme entities.SchemaColumnEntity
			var shallowField entities.ShallowSchemaColumnEntity
			b, _ := json.Marshal(r)
			json.Unmarshal(b, &scheme)
			cols[scheme.Name]=scheme
			json.Unmarshal(b, &shallowField)
			schemes[scheme.Name]=shallowField
		}
		view = View{ Name : tableName, Description : tableName + " datas", 
	                  Path : d.getPath(tableName, d.Db.GetSQLView(), d.Db.GetSQLRestriction(),
						                 d.Db.GetSQLOrder(), "", additonnalRestriction... ),
					  Schema : map[string]interface{}{ tableName : schemes }, Items : []ViewItem{} }
	}
	for _, record := range results { 
		if record == nil { continue }
		if !shallow {
			rec := d.PostTreatRecord(record, fmt.Sprintf("%v", schemas[0][tool.SpecialIDParam]), tableName, cols)
			if rec == nil { continue }
			view.Items = append(view.Items, rec.(ViewItem))
		} 
		if !d.IsShallowed() && !d.IsAdminView() {
			var r tool.Record
			b, _ := json.Marshal(view)
			json.Unmarshal(b, &r)
			res = append(res, r)
		} else { res = append(res, record) }
	}
	return res
}

func (d *MainService) PostTreatRecord(record tool.Record, tableID string, tableName string, 
									  cols map[string]entities.SchemaColumnEntity) interface{} {
	if d.IsShallowed() {
		if _, ok := record[entities.NAMEATTR]; ok {
			return tool.Record{ entities.NAMEATTR : record[entities.NAMEATTR] }
		} else { return record }
	} else {
		if d.IsAdminView() { return record } // if admin view avoid.
		contents := map[string]interface{}{}
		vals := map[string]interface{}{}
		data := tool.Record{}
		for _, field := range cols {
			if d.Db.GetSQLView() != "" && !strings.Contains(d.Db.GetSQLView(), field.Name){ continue }
			dest, ok := record[entities.RootID("dest_table")]
			id, ok2 := record[field.Name]
			if strings.Contains(field.Name, entities.DBSchema.Name) && ok2 && ok { 
				schemas, err := d.Schema(tool.Record{ entities.RootID(entities.DBSchema.Name) : id })
				if err != nil || len(schemas) == 0 { continue }
				rec := d.PostTreat(tool.Results{tool.Record{}}, fmt.Sprintf("%v", schemas[0][entities.NAMEATTR]), 
					true, "id=" + fmt.Sprintf("%v", dest))
				if len(rec) == 0 { continue }
				data = rec[0]
				continue
			}
			vals[field.Name]=record[field.Name]
			if field.Link != "" && !strings.Contains(field.Name, "dest") {
				rec := d.PostTreat(tool.Results{tool.Record{}}, field.Link, 
					   true, "id=" + fmt.Sprintf("%v", record[field.Name]))
				if len(rec) == 0 { continue}
				contents[field.Name]=rec[0]
				continue
			}
		}
		vi :=  ViewItem{ Values : vals, Data : data }
		if len(contents) > 0 { vi.Contents=contents }
		return vi
	}
}