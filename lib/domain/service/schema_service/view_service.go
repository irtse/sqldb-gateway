package schema_service

import (
	"fmt"
	"strings"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
)
//WORKING BUT NEED A CLEAN UP
type ViewService struct { tool.AbstractSpecializedService }

func (s *ViewService) Entity() tool.SpecializedServiceInfo { return entities.DBView }
func (s *ViewService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool, bool) { 
	if _, ok := record["through_perms"]; ok { 
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
	return record, true, false
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
		if record["is_empty"] != nil { s.Domain.SetEmpty(record["is_empty"].(bool)) }
		if r, ok := record["readonly"]; ok && r.(bool) { readonly = true }
		rec := tool.Record{ "id": record["id"], "name" : record["name"], "description" : record["description"], "is_empty" : record["is_empty"],
		                    "index" : record["index"], "is_list" : record["is_list"], "readonly" : record["readonly"],
						}
		s.Domain.SetLowerRes(record["is_list"].(bool))
		for _, dest := range dest_id {
			if id == "" { id = dest 
			} else { id = "," + dest  }
		}
		schemas, err := s.Domain.Schema(record, true)
		if err != nil || len(schemas) == 0 { continue }
		rec["category"]=schemas[0]["label"]
		tName := fmt.Sprintf("%v", schemas[0][entities.NAMEATTR])
		path, params := s.Domain.GeneratePathFilter("/" + tool.MAIN_PREFIX + "/" + tName, 
		                                            record, tool.Params{ tool.RootTableParam : tName, 
			                                        tool.RootRowsParam: tool.ReservedParam, })
		if id != "" { params[tool.RootRowsParam] = id }
		rec["link_path"]=s.Domain.BuildPath(fmt.Sprintf(entities.DBView.Name), fmt.Sprintf("%v", record[tool.SpecialIDParam]))
		sqlFilter := ""
		if _, ok := record["through_perms"]; ok { 
			through, err := s.Domain.Schema(tool.Record{  entities.RootID(entities.DBSchema.Name) : record["through_perms"] }, true)
			if len(through) > 0 && err == nil {
				sqlFilter +=  s.Domain.ByEntityUser(fmt.Sprintf("%v", through[0][entities.NAMEATTR]), tName)
			}
		}
		datas := tool.Results{tool.Record{}}
		if !s.Domain.GetEmpty() {
			if restr, ok := record["sql_restriction"]; ok && restr != "" && restr != nil {
				if len(sqlFilter) > 0 { 
					sqlFilter +=  " AND (" 
					sqlFilter += fmt.Sprintf("%v", restr)
					sqlFilter +=  ")"
				} else { sqlFilter += " " + fmt.Sprintf("%v", restr) }
			}
			datas, _ = s.Domain.PermsSuperCall( params, tool.Record{}, tool.SELECT, "Get", sqlFilter)
		}
		treated := s.Domain.PostTreat(datas, tName)
		if len(treated ) > 0 {
			for k, v := range treated[0] { 
				if _, ok := rec[k]; ok { continue }
				if k == "items"  {
					for _, item := range v.([]interface{}) {
						values := item.(map[string]interface{})["values"]
						if list, ok := record["is_list"]; ok && list.(bool) && len(path) > 0 && path[:1] == "/" {
							nP := ""
							if strings.Contains(path, entities.DBView.Name) { nP =  "/" + tool.MAIN_PREFIX + path + "&" + tool.RootDestTableIDParam + "=" + fmt.Sprintf("%v", values.(map[string]interface{})[tool.SpecialIDParam])
							} else { nP =  "/" + tool.MAIN_PREFIX + "/" + tName + "?" + tool.RootRowsParam + "=" + fmt.Sprintf("%v", values.(map[string]interface{})[tool.SpecialIDParam]) }
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
		res = append(res,  rec)
	}
	return res
}

func (s *ViewService) ConfigureFilter(tableName string) (string, string) { 
	return s.Domain.ViewDefinition(tableName)
}	