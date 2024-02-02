package user_service

import (
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
)

type EntityService struct { tool.AbstractSpecializedService }

func (s *EntityService) Entity() tool.SpecializedServiceInfo { return entities.DBEntity }
func (s *EntityService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool, bool) { 
	return record, true, false }
func (s *EntityService) DeleteRowAutomation(results tool.Results, tableName string) { }
func (s *EntityService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *EntityService) WriteRowAutomation(record tool.Record, tableName string) { }
func (s *EntityService) PostTreatment(results tool.Results, tableName string, dest_id... string) tool.Results { 	
	return s.Domain.PostTreat( results, tableName, false) 
}
func (s *EntityService) ConfigureFilter(tableName string, params  tool.Params) (string, string) {
	params[tool.RootSQLFilterParam] = "id IN (SELECT " + entities.RootID(entities.DBEntity.Name) + " FROM " +  entities.DBEntityUser.Name + " " 
	params[tool.RootSQLFilterParam] += "WHERE " + entities.RootID(entities.DBUser.Name) + " IN ("
	params[tool.RootSQLFilterParam] += "SELECT id FROM " + entities.DBUser.Name + " WHERE login='" + s.Domain.GetUser() + "')) "
	return s.Domain.ViewDefinition(tableName, params)
}