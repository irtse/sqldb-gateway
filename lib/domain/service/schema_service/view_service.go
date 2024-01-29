package schema_service

import (
	"fmt"
	"strings"
	"encoding/json"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/infrastructure/entities"
)

type ViewService struct { tool.AbstractSpecializedService }

func (s *ViewService) Entity() tool.SpecializedServiceInfo { return entities.DBView }
func (s *ViewService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool) { 
	if _, ok := record["through_perms"]; !ok { return record, false }
	schemas, err := tool.Schema(s.Domain, tool.Record{ 
		entities.RootID(entities.DBSchema.Name) : fmt.Sprintf("%v", record["through_perms"]) })	
	if err != nil && len(schemas) == 0 { return record, false }
	params := tool.Params{ tool.RootTableParam : entities.DBSchemaField.Name, 
	                       tool.RootRowsParam : tool.ReservedParam, 
						   entities.RootID(entities.DBSchema.Name) : fmt.Sprintf("%v", schemas[0][tool.SpecialIDParam]),
						   entities.NAMEATTR : entities.RootID(entities.DBUser.Name),
						 }
	userScheme, _ := s.Domain.SuperCall(params, tool.Record{}, tool.SELECT, "Get")
	params[entities.NAMEATTR] = entities.RootID(entities.DBEntity.Name)
	entityScheme, _ := s.Domain.SuperCall(params, tool.Record{}, tool.SELECT, "Get")
	found := false 
	if len(userScheme) > 0 { found = true }
	if len(entityScheme) > 0 { found = true }
    if found { return record, true }
	return nil, false
}
func (s *ViewService) DeleteRowAutomation(results tool.Results, tableName string) { }
func (s *ViewService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *ViewService) WriteRowAutomation(record tool.Record, tableName string) { }
func (s *ViewService) PostTreatment(results tool.Results, tableName string) tool.Results { 
	if len(results) == 0 { return results }
	res := tool.Results{}
	for _, record := range results {
		readonly := false 
		if r, ok := record["readonly"]; ok && r.(bool) { readonly = true }
		rec := tool.Record{ "name" : record["name"], "description" : record["description"], 
		               "index" : record["index"], "category" : record["category"], "is_list" : record["is_list"], }
		cols := map[string]entities.SchemaColumnEntity{}
		if !s.Domain.IsShallowed() {
			params := tool.Params{ tool.RootTableParam : entities.DBSchemaField.Name, tool.RootRowsParam: tool.ReservedParam, 
				                   entities.RootID(entities.DBSchema.Name) : fmt.Sprintf("%v", record[entities.RootID(entities.DBSchema.Name)]) }
			schemas, err := s.Domain.SuperCall( params, tool.Record{}, tool.SELECT, "Get")
			if err != nil || len(schemas) == 0 { return tool.Results{} }
			for _, r := range schemas {
				var scheme entities.SchemaColumnEntity
				b, _ := json.Marshal(r)
				json.Unmarshal(b, &scheme)
				cols[scheme.Name]=scheme
				scheme.Readonly = readonly
			}
		}
		schemas, err := tool.Schema(s.Domain, record)
		if err != nil && len(schemas) == 0 { continue }
		tName := fmt.Sprintf("%v", schemas[0][entities.NAMEATTR])
		through, err := tool.Schema(s.Domain, tool.Record{  entities.RootID(entities.DBSchema.Name) : record["through_perms"] })
		sqlFilter := ""
		if err == nil && len(through) > 0 { 
			sqlFilter += "id IN (SELECT " + entities.RootID(tName) 
			sqlFilter += " FROM " + fmt.Sprintf("%v", through[0][entities.NAMEATTR])
			sqlFilter += " WHERE " + entities.RootID(entities.DBUser.Name) 
			sqlFilter += " IN (SELECT id FROM " + entities.DBUser.Name + " WHERE login='" + s.Domain.GetUser() + "')" 
			sqlFilter += " OR " + entities.RootID(entities.DBEntity.Name) + " IN ("
			sqlFilter += "SELECT " + entities.RootID(entities.DBEntity.Name) + " FROM " + entities.DBEntityUser.Name + " "
			sqlFilter += "WHERE " + entities.RootID(entities.DBUser.Name) + " IN ("
			sqlFilter += "SELECT id FROM " + entities.DBUser.Name + " WHERE login='" + s.Domain.GetUser() + "')))"
		}
		if restr, ok2 := record["sql_restriction"]; ok2 && restr != nil && restr != "" { 
			if len(sqlFilter) > 0 { sqlFilter += " AND " + strings.Replace(restr.(string), "+", " ", -1) 
			} else { sqlFilter = strings.Replace(restr.(string), "+", " ", -1)  } 
		}
		params := tool.Params{ tool.RootTableParam : tName, 
						  tool.RootRowsParam: tool.ReservedParam, 
						  tool.RootSQLFilterParam : sqlFilter, }
		if columns, ok2 := record["sql_view"]; ok2 && columns != nil && columns != "" { params[tool.RootColumnsParam] = fmt.Sprintf("%v", columns) }
		if order, ok2 := record["sql_order"]; ok2 && order != nil && order != "" { params[tool.RootOrderParam] = fmt.Sprintf("%v", order) }
		if dir, ok2 := record["sql_dir"]; ok2 && dir != nil  && dir != "" { params[tool.RootDirParam] = fmt.Sprintf("%v", dir) }
		datas, err := s.Domain.SuperCall( params, tool.Record{}, tool.SELECT, "Get")
		rec["views"]=[]tool.Record{}
		nR := tool.Record{ }	
		for _, data := range datas {
			r := tool.PostTreatRecord(s.Domain, data, fmt.Sprintf("%v", data[tool.SpecialIDParam]), tName, cols, false, []string{ sqlFilter }...)
			if empty, ok := record["is_empty"]; ok && empty.(bool) { 
				empty := tool.Record{}
				for key, _ := range r["values"].(map[string]interface{}) { empty[key]=nil } 
				rec["values"]=empty
			}			
			nR[tName] = r["schema"]
			delete(r, "schema")
			rec["views"] = append(rec["views"].([]tool.Record), r)
		}	
		rec["schemas"]=nR
		if view, ok3 := record[entities.RootID(entities.DBView.Name)]; ok3 {
			params := tool.Params{ tool.RootTableParam : entities.DBView.Name, 
				                   tool.RootRowsParam: fmt.Sprintf("%v", view), }
			views, err := s.Domain.SuperCall( params, tool.Record{}, tool.SELECT, "Get")
			if err == nil && len(views) > 0 { rec["item_view"]=views[0] }
		}
		params = tool.Params{ tool.RootTableParam : entities.DBAction.Name, tool.RootRowsParam: tool.ReservedParam, }
		params[tool.RootSQLFilterParam] = "id IN (SELECT " + entities.RootID(entities.DBAction.Name) + " FROM " + entities.DBViewAction.Name
		params[tool.RootSQLFilterParam] += " WHERE " + entities.RootID(entities.DBView.Name) + "=" 
		params[tool.RootSQLFilterParam] += fmt.Sprintf("%v", record[tool.SpecialIDParam]) + ")"
		actions, err := s.Domain.Call( params, tool.Record{}, tool.SELECT, false, "Get")
		fmt.Printf("ERR ACT %v %v \n", err, actions)
		if err == nil && len(actions) > 0 {
			rec["actions"] = tool.Results{}
			for _, action := range actions { rec["actions"] = append(rec["actions"].(tool.Results), action) }
		}
		res = append(res,  rec)
	}
	return res
}

func (s *ViewService) ConfigureFilter(tableName string, params tool.Params) (string, string) { 
	return tool.ViewDefinition(s.Domain, tableName, params)
}	