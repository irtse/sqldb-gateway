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
	schemas, err := Schema(s.Domain, tool.Record{ 
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
func (s *ViewService) PostTreatment(results tool.Results) tool.Results { 
	res := tool.Results{}
	for _, record := range results{
		schemas, err := Schema(s.Domain, record)
		if err != nil || len(schemas) == 0 { continue }
		params := tool.Params{ tool.RootTableParam : schemas[0][entities.NAMEATTR].(string), }
		schemes, err := s.Domain.SuperCall(params, tool.Record{}, tool.SELECT, "Get")
		if err == nil && len(schemes) > 0 {
			recSchemes := map[string]tool.Record{}
			for _, scheme := range schemes { 
				delete(record, entities.RootID(entities.DBSchema.Name))
				record["contents"]=[]string{ scheme[entities.NAMEATTR].(string) }
				recSchemes[scheme[entities.NAMEATTR].(string)]=scheme 
			}
			record["schemas"]=recSchemes
		}
		params = tool.Params{ tool.RootTableParam : entities.DBAction.Name, 
			                  tool.RootRowsParam : tool.ReservedParam, }
		params[tool.RootSQLFilterParam]="id IN (SELECT " + entities.RootID(entities.DBAction.Name) + " FROM " + entities.DBViewAction.Name + " "
		params[tool.RootSQLFilterParam] += "WHERE " + entities.RootID(entities.DBView.Name) + "=" + fmt.Sprintf("%v", record[tool.SpecialIDParam]) + "  );"
		actions, err := s.Domain.Call(params, tool.Record{}, tool.SELECT, false, "Get")
		if err == nil && len(actions) > 0 {
			recActions := []tool.Record{}
			for _, action := range actions { recActions=append(recActions, action) }
			record["actions"]=recActions
		}
		res = append(res, record)
	}
	return res 
}

func (s *ViewService) ConfigureFilter(tableName string, params tool.Params) (string, string) { return "", "" }	