package schema_service

import (
	"fmt"
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
func (s *ViewService) DeleteRowAutomation(results tool.Results) { }
func (s *ViewService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *ViewService) WriteRowAutomation(record tool.Record) { }
func (s *ViewService) PostTreatment(results tool.Results, tableName string) tool.Results { 
	if len(results) == 0 { return results }
	res := tool.Results{}
	var cols map[string]entities.TableColumnEntity
	if !s.Domain.IsShallowed() {
		params := tool.Params{ tool.RootTableParam : tableName, }
			schemas, err := s.Domain.Call( params, tool.Record{}, tool.SELECT, true, "Get")
			if err != nil || len(schemas) == 0 { return tool.Results{} }
			if _, ok := schemas[0]["columns"]; !ok { return tool.Results{} }
			cols = schemas[0]["columns"].(map[string]entities.TableColumnEntity)
	}
	for _, record := range results {
		schemas, err := tool.Schema(s.Domain, record)
		if err != nil && len(schemas) == 0 { continue }
		through, err := tool.Schema(s.Domain, tool.Record{  entities.RootID(entities.DBSchema.Name) : record["through_perms"] })
		sqlFilter := ""
		if err == nil && len(through) > 0 { 
			sqlFilter += "id IN (SELECT " + entities.RootID(fmt.Sprintf("%v", schemas[0][entities.NAMEATTR])) 
			sqlFilter += " FROM " + fmt.Sprintf("%v", through[0][entities.NAMEATTR])
			sqlFilter += " WHERE " + entities.RootID(entities.DBUser.Name) 
			sqlFilter += " IN (SELECT id FROM " + entities.DBUser.Name + " WHERE login='" + s.Domain.GetUser() + "')" 
			sqlFilter += " OR " + entities.RootID(entities.DBEntity.Name) + " IN ("
			sqlFilter += "SELECT " + entities.RootID(entities.DBEntity.Name) + " FROM " + entities.DBEntityUser.Name + " "
			sqlFilter += "WHERE " + entities.RootID(entities.DBUser.Name) + " IN ("
			sqlFilter += "SELECT id FROM " + entities.DBUser.Name + " WHERE login='" + s.Domain.GetUser() + "'))"
		}
		r := tool.PostTreatRecord(s.Domain, record, tableName, cols, []string{ sqlFilter }...)
		r["is_view"]=true
		res = append(res,  r)
	}
	return res
}

func (s *ViewService) ConfigureFilter(tableName string, params tool.Params) (string, string) { 
	return tool.ViewDefinition(s.Domain, tableName, params)
}	