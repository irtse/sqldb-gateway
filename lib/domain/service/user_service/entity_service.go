package user_service

import (
	"fmt"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/infrastructure/entities"
)

type EntityService struct { tool.AbstractSpecializedService }

func (s *EntityService) Entity() tool.SpecializedServiceInfo { return entities.DBEntity }
func (s *EntityService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool) { return record, true }
func (s *EntityService) DeleteRowAutomation(results tool.Results) { }
func (s *EntityService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *EntityService) WriteRowAutomation(record tool.Record) { }
func (s *EntityService) PostTreatment(results tool.Results) tool.Results {
	res := tool.Results{} // TODO COMPLEXIFY WITH ENTITY ID
	params := tool.Params{ tool.RootTableParam : entities.DBUser.Name, 
							tool.RootRowsParam : tool.ReservedParam,
							"login" : s.Domain.GetUser() }
	users, err := s.Domain.SuperCall( params, tool.Record{}, tool.SELECT, "Get", )	
	if err != nil || len(users) == 0 { return res }
	for _, record := range results {
		params = tool.Params{ tool.RootTableParam : entities.DBEntityUser.Name, 
				                tool.RootRowsParam : tool.ReservedParam,
								entities.RootID(entities.DBEntity.Name): fmt.Sprintf("%v", record[tool.SpecialIDParam]),
				                entities.RootID(entities.DBUser.Name): fmt.Sprintf("%v", users[0][tool.SpecialIDParam]) }
		affectedEntity, err := s.Domain.Call( params, tool.Record{}, tool.SELECT, false, "Get", )	
		if err == nil && len(affectedEntity) > 0 { res = append(res, record) }
	}
	return res
}

func (s *EntityService) ConfigureFilter(tableName string, params  tool.Params) (string, string) {
	return tool.ViewDefinition(s.Domain, tableName, params)
}