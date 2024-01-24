package schema_service

import (
	"fmt"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/infrastructure/entities"
)

type ViewService struct { tool.AbstractSpecializedService }

func (s *ViewService) Entity() tool.SpecializedServiceInfo { return entities.DBView }
func (s *ViewService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool) { 
	schemas, err := Schema(s.Domain, record)	
	if err != nil && len(schemas) == 0 { return record, false }
	params := tool.Params{ tool.RootTableParam : record[entities.RootID(entities.DBSchemaField.Name)].(string), 
	                       tool.RootRowsParam : tool.ReservedParam, 
						   entities.NAMEATTR : entities.RootID(entities.DBUser.Name),
						 }
	userScheme, err := s.Domain.SuperCall(params, tool.Record{}, tool.SELECT, "Get")
	params[entities.NAMEATTR] = entities.RootID(entities.DBEntity.Name)
	entityScheme, err2 := s.Domain.SuperCall(params, tool.Record{}, tool.SELECT, "Get")
    if err == nil && err2 == nil && (len(userScheme) > 0 || len(entityScheme) > 0) { return record, true }
	return nil, false
}
func (s *ViewService) DeleteRowAutomation(results tool.Results) { }
func (s *ViewService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *ViewService) WriteRowAutomation(record tool.Record) { }
func (s *ViewService) PostTreatment(results tool.Results) tool.Results { 
	res := tool.Results{}
	for _, record := range results{
		schemas, err := Schema(s.Domain, tool.Record{entities.RootID(entities.DBSchema.Name) : record[tool.SpecialIDParam].(int64)})
		if err != nil || len(schemas) == 0 { continue }
		params := tool.Params{ tool.RootTableParam : record[entities.RootID(entities.DBSchema.Name)].(string), }
		schemes, err := s.Domain.SuperCall(params, tool.Record{}, tool.SELECT, "Get")
		if err == nil && len(schemes) > 0 {
			recSchemes := map[string]tool.Record{}
			for _, scheme := range schemes { recSchemes[scheme[entities.NAMEATTR].(string)]=scheme }
			record["schemas"]=recSchemes
		}
		params = tool.Params{ tool.RootTableParam : record[entities.RootID(entities.DBViewAction.Name)].(string), 
			                   tool.RootRowsParam : tool.ReservedParam,
							   entities.RootID(entities.DBView.Name): fmt.Sprintf("%d", record[tool.SpecialIDParam].(int64)) }
		actions, err := s.Domain.SuperCall(params, tool.Record{}, tool.SELECT, "Get")
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