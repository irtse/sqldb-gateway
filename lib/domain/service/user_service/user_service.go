package user_service

import (
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/infrastructure/entities"
)

type UserService struct { tool.AbstractSpecializedService }

func (s *UserService) Entity() tool.SpecializedServiceInfo { return entities.DBUser }
func (s *UserService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool) { return record, true }
func (s *UserService) DeleteRowAutomation(results tool.Results, tableName string) { }
func (s *UserService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *UserService) WriteRowAutomation(record tool.Record, tableName string) { }
func (s *UserService) PostTreatment(results tool.Results, tableName string) tool.Results { 	
	return tool.PostTreat(s.Domain, results, tableName, false) 
}
func (s *UserService) ConfigureFilter(tableName string, params  tool.Params) (string, string) {
	params[tool.RootSQLFilterParam] = "id IN (SELECT id FROM " + entities.DBUser.Name + " WHERE login='" + s.Domain.GetUser() + "') "
	return tool.ViewDefinition(s.Domain, tableName, params)
}