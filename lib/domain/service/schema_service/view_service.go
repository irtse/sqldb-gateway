package schema_service

import (
	"fmt"
	"strings"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
	conn "sqldb-ws/lib/infrastructure/connector"
)
//WORKING BUT NEED A CLEAN UP
type ViewService struct { tool.AbstractSpecializedService }

func (s *ViewService) Entity() tool.SpecializedServiceInfo { return entities.DBView }
func (s *ViewService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool, bool) { 
	if _, ok := record["through_perms"]; !ok { return record, false, false }
	schemas, err := s.Domain.Schema(tool.Record{ 
		entities.RootID(entities.DBSchema.Name) : fmt.Sprintf("%v", record["through_perms"]) }, true)	
	if err != nil && len(schemas) == 0 { return record, false, false }
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
    if found { return record, true, false }
	return nil, false, false
}
func (s *ViewService) DeleteRowAutomation(results tool.Results, tableName string) { }
func (s *ViewService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *ViewService) WriteRowAutomation(record tool.Record, tableName string) { }
func (s *ViewService) PostTreatment(results tool.Results, tableName string, dest_id... string) tool.Results { 
	if len(results) == 0 { return results }
	res := tool.Results{}
	for _, record := range results {
		readonly := false 
		id := ""
		if r, ok := record["readonly"]; ok && r.(bool) { readonly = true }
		rec := tool.Record{ "id": record["id"], "name" : record["name"], "description" : record["description"],
		                    "index" : record["index"], "category" : record["category"],
							"is_list" : record["is_list"], "readonly" : record["readonly"], }
		for _, dest := range dest_id {
			if id == "" { id = dest 
			} else { id = "," + dest  }
		}
		schemas, err := s.Domain.Schema(record, true)
		if err != nil || len(schemas) == 0 { continue }
		tName := fmt.Sprintf("%v", schemas[0][entities.NAMEATTR])
		through, err := s.Domain.Schema(tool.Record{  entities.RootID(entities.DBSchema.Name) : record["through_perms"] }, true)
		sqlFilter := ""
		if err == nil && len(through) > 0 { 
			sqlFilter += "id IN (SELECT " + entities.RootID(tName) 
			sqlFilter += " FROM " + fmt.Sprintf("%v", through[0][entities.NAMEATTR])
			sqlFilter += " WHERE " + entities.RootID(entities.DBUser.Name) 
			sqlFilter += " IN (SELECT id FROM " + entities.DBUser.Name + " WHERE name=" + conn.Quote(s.Domain.GetUser()) + " OR email=" + conn.Quote(s.Domain.GetUser()) + ")" 
			sqlFilter += " OR " + entities.RootID(entities.DBEntity.Name) + " IN ("
			sqlFilter += "SELECT " + entities.RootID(entities.DBEntity.Name) + " FROM " + entities.DBEntityUser.Name + " "
			sqlFilter += "WHERE " + entities.RootID(entities.DBUser.Name) + " IN ("
			sqlFilter += "SELECT id FROM " + entities.DBUser.Name + " WHERE name=" + conn.Quote(s.Domain.GetUser()) + " OR email=" + conn.Quote(s.Domain.GetUser()) + ")))"
		}
		path, params := s.Domain.GeneratePathFilter("/" + tool.MAIN_PREFIX + "/" + tName, 
		                                            record, tool.Params{ tool.RootTableParam : tName, 
			                                        tool.RootRowsParam: tool.ReservedParam, })
		if id != "" { params[tool.RootRowsParam] = id }
		rec["link_path"]=s.Domain.BuildPath(fmt.Sprintf(entities.DBView.Name), fmt.Sprintf("%v", record[tool.SpecialIDParam]))
		if s.Domain.IsShallowed() { res = append(res, rec); continue }	
		datas, err := s.Domain.PermsSuperCall( params, tool.Record{}, tool.SELECT, "Get")
		empty, ok := record["is_empty"]
		treated := s.Domain.PostTreat(datas, tName, ok && empty.(bool), []string{ sqlFilter }...)
		if len(treated) > 0 {
			for k, v := range treated[0] { 
				if _, ok := rec[k]; !ok { 
					if k == "items" && len(path) > 0 && path[:1] == "/" && record["is_list"].(bool) {
						for _, item := range v.([]interface{}) {
							if strings.Contains(path, entities.DBView.Name) {
								nP :=  "/" + tool.MAIN_PREFIX + path 
								values := item.(map[string]interface{})["values"]
								if valID, ok := values.(map[string]interface{})[tool.SpecialIDParam]; ok {
									nP += "&" + tool.RootDestTableIDParam + "=" + fmt.Sprintf("%v", valID)
								}
								item.(map[string]interface{})["link_path"] = nP
								item.(map[string]interface{})["data_path"] = ""
							}
						}
						rec[k]=v 
					} else if k == "schema" { 
						newV := map[string]interface{}{}
						for fieldName, field := range v.(map[string]interface{}) {
							if readonly { field.(map[string]interface{})["readonly"] = true }
							if view, ok := params[tool.RootColumnsParam]; !ok || view == "" || strings.Contains(view, fieldName) { 
								newV[fieldName] = field 
							}
						}
						rec[k] = newV
					} else { rec[k]=v }
				}
			}
		}
		params = tool.Params{ tool.RootTableParam : entities.DBAction.Name, tool.RootRowsParam: tool.ReservedParam, }
		sqlFilter = "id IN (SELECT " + entities.RootID(entities.DBAction.Name) + " FROM " + entities.DBViewAction.Name
		sqlFilter += " WHERE " + entities.RootID(entities.DBView.Name) + "=" 
		sqlFilter += fmt.Sprintf("%v", record[tool.SpecialIDParam]) + ")"
		actions, err := s.Domain.Call( params, tool.Record{}, tool.SELECT, false, "Get", sqlFilter)
		if err == nil && len(actions) > 0 {
			rec["actions"] = tool.Results{}
			for _, action := range actions { 
				rec["actions"] = append(rec["actions"].(tool.Results), action) 
			}
		}
		res = append(res,  rec)
	}
	return res
}

func (s *ViewService) ConfigureFilter(tableName string) (string, string) { 
	return s.Domain.ViewDefinition(tableName)
}	