package domain

import (
	"fmt"
	"sort"
	"slices"
	"errors"
	"strings"
	"encoding/json"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
	conn "sqldb-ws/lib/infrastructure/connector"
)

// define filter whatever what happen on sql...
func (d *MainService) ViewDefinition(tableName string, innerRestriction... string) (string, string) {
	SQLview := ""; SQLrestriction := ""
	if d.IsSuperCall() { return SQLrestriction, SQLview } // admin can see all on admin view
	SQLrestriction = d.ByEntityUser(tableName)
	SQLview = d.byFields(tableName)
	if len(innerRestriction) > 0 {
		for _, restr := range innerRestriction {
			if len(strings.TrimSpace(restr)) > 0 {
				if len(SQLrestriction) > 0 { SQLrestriction += " AND (" + restr + ")"
				} else { SQLrestriction = restr  }
			}
		}
	}
	return SQLrestriction, SQLview
}

func (s *MainService) ByEntityUser(tableName string, extra ...string) (string) {
	schemas, err := s.Schema(tool.Record{ entities.NAMEATTR : tableName }, false)
	restr := ""
	if len(extra) > 1 {
		restr += "id IN (SELECT " + entities.RootID(extra[1]) + " FROM " + tableName + " WHERE "
	}
	if err != nil && schemas != nil && len(schemas) > 0 { 
		userID := entities.RootID(entities.DBUser.Name)
		entityID := entities.RootID(entities.DBEntity.Name)
		if _, ok := schemas[0][userID]; ok || tableName == entities.DBUser.Name {
			if tableName == entities.DBUser.Name  { userID = tool.SpecialIDParam }
			restr += userID + " IN (SELECT id FROM " + entities.DBUser.Name + " WHERE name=" + conn.Quote(s.GetUser()) + " OR email=" + conn.Quote(s.GetUser()) + ")" 
			if _, ok := schemas[0][entityID]; ok || tableName == entities.DBEntity.Name  {
				if tableName == entities.DBEntity.Name  { entityID = tool.SpecialIDParam }
				if len(restr) > 0 { restr +=  " OR " }
				restr += entityID + " IN (SELECT " + entityID + " FROM " + entities.DBEntityUser.Name + "WHERE " + entities.RootID(entities.DBUser.Name) + " IN (" 
				restr += "SELECT id FROM " + userID + " WHERE name=" + conn.Quote(s.GetUser()) + " OR email=" + conn.Quote(s.GetUser()) + ")"
			}
		}
	}
	if len(extra) > 1 { restr += ")" }
	return restr
}

var fieldsCache = map[string]tool.Results{}

func (d *MainService) byFields(tableName string) (string) {
	SQLview := "id,"
	sqlFilter := entities.RootID(entities.DBSchema.Name) + " IN (SELECT id FROM " + entities.DBSchema.Name + " WHERE name='" + tableName + "')"
	views := []string{}
	if params, ok := d.Params[tool.RootColumnsParam]; ok { 
		views = strings.Split(params, ",") 
	}
	if fieldsCache[tableName] == nil  {
		p := tool.Params{ tool.RootTableParam : entities.DBSchemaField.Name, tool.RootRowsParam : tool.ReservedParam,}
		fields, err := d.SuperCall( p, tool.Record{}, tool.SELECT, "Get", sqlFilter)
		if err != nil || len(fields) == 0 { return "" }
		fieldsCache[tableName] = fields
	}
	for _, field := range fieldsCache[tableName] {
		if len(views) > 0 && !slices.Contains(views, field.GetString(entities.NAMEATTR)) { continue }
		if strings.Contains(strings.ToLower(fmt.Sprintf("%v", field[entities.TYPEATTR])) , "many") { continue }
		if d.PermsCheck(tableName, fmt.Sprintf("%v", field[entities.NAMEATTR]), fmt.Sprintf("%v", field["read_level"]), tool.SELECT) {
			SQLview += fmt.Sprintf("%v",field[entities.NAMEATTR]) + ","
		}
	}
	if len(SQLview) > 0 { SQLview = SQLview[:len(SQLview) - 1] }
	return SQLview
}

var schemaCache = map[string]tool.Results{}

func (d *MainService) Schema(record tool.Record, permitted bool) (tool.Results, error) { // check schema auth access
	params := tool.Params{ tool.RootTableParam : entities.DBSchema.Name, 
			               tool.RootRowsParam : tool.ReservedParam }
	var schemas tool.Results
	var err error
	sqlFilter := ""
	if id, ok := record[entities.RootID(entities.DBSchema.Name)]; ok {
		if schemaCache[fmt.Sprintf("%v", id)] != nil {  schemas = schemaCache[fmt.Sprintf("%v", id)] }
		sqlFilter += "id=" + fmt.Sprintf("%v", id)
	} else if name, ok := record[entities.NAMEATTR]; ok {
		if schemaCache[fmt.Sprintf("%v", name)] != nil { schemas = schemaCache[fmt.Sprintf("%v", name)] }
		sqlFilter += "name='" + fmt.Sprintf("%v", name) + "'"
	}
	if schemas == nil {
		schemas, err = d.SuperCall( params, tool.Record{}, tool.SELECT, "Get", sqlFilter)
		if err != nil || len(schemas) == 0 { return nil, err }
	}
	if id, ok := record[entities.RootID(entities.DBSchema.Name)]; ok { schemaCache[fmt.Sprintf("%v", id)] = schemas
	} else if name, ok := record[entities.NAMEATTR]; ok { schemaCache[fmt.Sprintf("%v", name)] = schemas }
	if permitted && !d.PermsCheck(
		fmt.Sprintf("%v", schemas[0][entities.NAMEATTR]), "", "", tool.SELECT) {
		return nil, errors.New("not authorized")
	}
	return schemas, err
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

func (d *MainService) BuildPath(tableName string, rows string, extra... string) string {
	path := "/" + tool.MAIN_PREFIX + "/" + tableName + "?rows=" + rows
	for _, ext := range extra { path += "&" + ext }
	return path
}

func (d *MainService) GetScheme(tableName string, isId bool) (
	map[string]interface{}, int64, []string, map[string]entities.SchemaColumnEntity, []string) {
	cols := map[string]entities.SchemaColumnEntity{}
	keysOrdered := []string{}
	sqlFilter := ""
	additionnalAction := []string{}
	var id int64
	schemes := map[string]interface{}{}
	if fieldsCache[tableName] == nil {
		if isId { sqlFilter += entities.RootID(entities.DBSchema.Name) + "=" + tableName
		} else { 
			sqlFilter += entities.RootID(entities.DBSchema.Name) + " IN (SELECT id FROM "
			sqlFilter += entities.DBSchema.Name + " WHERE name=" + conn.Quote(tableName) + ")" 
		}
		// retrive all fields from schema...
		params := tool.Params{ tool.RootTableParam : entities.DBSchemaField.Name, 
							tool.RootRowsParam: tool.ReservedParam, }
		schemas, err := d.SuperCall( params, tool.Record{}, tool.SELECT, "Get", sqlFilter)
		if err != nil || len(fieldsCache[tableName]) == 0 { return schemes, id, keysOrdered, cols, additionnalAction }
		fieldsCache[tableName] = schemas
	}
	for _, r := range fieldsCache[tableName] {
		var scheme entities.SchemaColumnEntity
		var shallowField entities.ShallowSchemaColumnEntity
		b, _ := json.Marshal(r)
		json.Unmarshal(b, &scheme)
		if !d.PermsCheck(tableName, scheme.Name, scheme.Level, tool.SELECT) { continue }
		cols[scheme.Name]=scheme
		id = scheme.SchemaId
		json.Unmarshal(b, &shallowField)
		shallowField.ActionPath = ""
		shallowField.Actions=[]string{}
		if scheme.Link != "" && !d.LowerRes {
			shallowField.LinkPath = "/" + tool.MAIN_PREFIX + "/" + scheme.Link + "?rows=all"
			if scheme.LinkView != "" { shallowField.LinkPath += "&" + tool.RootColumnsParam + "=" + scheme.LinkView  
			} else if !strings.Contains(scheme.Type, "many") { shallowField.LinkPath += "&" + tool.RootShallow + "=enable" 
			} else {
				isSkipped := false
				for _, meth := range []tool.Method{ tool.SELECT, tool.CREATE, tool.UPDATE, tool.DELETE } {
					if d.PermsCheck(scheme.Link, "", "", meth) { 
						additionnalAction = append(additionnalAction, meth.Method())
						sch, _, ordered, _, _ := d.GetScheme(scheme.Link, false)
						shallowField.DataSchema = sch
						shallowField.DataSchemaOrder = ordered
						shallowField.ActionPath = "/" + tool.MAIN_PREFIX + "/" + scheme.Link + "?rows=" + tool.ReservedParam
						shallowField.Actions=append(shallowField.Actions, meth.Method())
					} else if meth == tool.UPDATE && !d.Empty { shallowField.Readonly = true 
					} else if meth == tool.CREATE && d.Empty { shallowField.Readonly = true 
					} else if meth == tool.SELECT { isSkipped = true }
				} 
				if isSkipped { continue }
			}
			if scheme.LinkOrder != "" { shallowField.LinkPath += "&" + tool.RootOrderParam + "=" + scheme.LinkOrder  }
		}
		if !d.Empty && !d.PermsCheck(tableName, scheme.Name, scheme.Level, tool.UPDATE) || d.Empty && !d.PermsCheck(tableName, scheme.Name, scheme.Level, tool.CREATE) {
			shallowField.Readonly=true
		}
		keysOrdered = append(keysOrdered, scheme.Name)
		schemes[scheme.Name]=shallowField
	}
	sort.SliceStable(keysOrdered, func(i, j int) bool{
        return schemes[keysOrdered[i]].(entities.ShallowSchemaColumnEntity).Index <= schemes[keysOrdered[j]].(entities.ShallowSchemaColumnEntity).Index
    })
	return schemes, id, keysOrdered, cols, additionnalAction
}